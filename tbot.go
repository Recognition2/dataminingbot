package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Tbot struct {
	Username string // Bot username
	apikey   string // Communication token with Telegram API
}

type User struct {
	Id         int    `json:"id"`         // Unique identifier for this user/bot
	First_name string `json:"first_name"` // User's or bot's first name
	Last_name  string `json:"last_name"`  // OPTIONAL: User's or bot's last name
	Username   string `json:"username"`   // OPTIONAL
}

type Update struct {
	Update_id           int64   `json:"update_id"`           // Unique identifier
	Message             Message `json:"message,omitempty"`   // OPTIONAL:	Message that was sent
	Edited_message      Message `json:"edited_message"`      // OPTIONAL: New version of a message
	Channel_post        Message `json:"channel_post"`        // OPTIONAL
	Edited_channel_post Message `json:"edited_channel_post"` // OPTIONAL
}

type Message struct {
	Message_id int64  `json:"message_id"` // Unique identifier
	From       User   `json:"from"`       // OPTIONAL: May be empty for channels
	Date       int    `json:"date"`       // Date message was sent, Unix time
	Chat       Chat   `json:"chat"`       // Conversation message belongs to
	Text       string `json:"text"`       // Raw UTF-8 text of message
}

type MessageEntity struct {
	Type string `json:"type"` // One of "mention", "hashtag", "bot_command", "text_link" (clickable text urls),
	// "text_mention" (users without usernames), "url", "email", "bold", "italic", "code", "pre".
	Offset int    `json:"offset"` // Offset in UTF-16 code units
	Length int    `json:"length"` // Length of entity in UTF-16 code units
	Url    string `json:"url"`    // OPTIONAL: for text_link only
	User   User   `json:"user"`   // OPTIONAL: For text_mention only
}

type Chat struct {
	Id       int64  `json:"id"`       // Unique identifier, max 52 bits.
	Type     string `json:"type"`     // One of: "private", "group", “supergroup” or “channel”
	Title    string `json:"title"`    // OPTIONAL
	Username string `json:"username"` // OPTIONAL: Not for groups.
}

type APIError struct {
	Ok          bool   `json:"ok"` // Is response successful?
	Error_code  int    `json:"error_code"`
	Description string `json:"description"`
}

type BasicReturn struct {
	Ok     bool        `json:"ok"`
	Result interface{} `json:"result"`
}

var turl = "https://api.telegram.org/bot"

// Get information about self
func (t Tbot) getUsername() (string, error) {
	resp, err := http.Get(turl + t.apikey + "/getMe")
	if err != nil {
		return "", err
	}

	self, err := ioutil.ReadAll(resp.Body) // Read stream
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	// Parse data
	var thisBot User
	var result json.RawMessage
	var isSucces = BasicReturn{
		Result: &result,
	}

	err = json.Unmarshal(self, &isSucces)
	if err != nil {
		return "", err
	}
	if isSucces.Ok != true {
		return "", errors.New(fmt.Sprintf("Telegram bot token is incorrect: %v", isSucces.Result))
	}

	err = json.Unmarshal(result, &thisBot)
	if err != nil {
		return "", err
	}

	// Print data
	logInfo.Printf("Telegram token valid, starting as @%v\n", thisBot.Username)
	return thisBot.Username, nil
}

func (t Tbot) getUpdates(offset int64, shutdown chan bool) ([]Update, error) {
	logInfo.Println("Getting an Update")

	var arg struct {
		Timeout         int      `json:"timeout"`
		Offset          int64    `json:"offset"`
		Limit           int      `json:"limit"`
		Allowed_updates []string `json:"allowed_updates"`
	}
	arg.Timeout = 120
	arg.Offset = offset
	arg.Limit = 1

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(arg)

	client := &http.Client{}

	r, err := http.Post(turl+t.apikey+"/getUpdates", "application/json; charset=utf-8", b)
	// http.Get has an inherent delay, while doing DNS lookup etc. (guaranteeing that your connection is valid)
	if err != nil {
		return nil, err
	}

	rawData := make(chan []byte)
	rawDataCompleted := make(chan bool)

	go func() {
		defer r.Body.Close()
		rawdata, err := ioutil.ReadAll(r.Body) // Read stream
		if err != nil {
			logErr.Printf("Error reading from response: %v\n", err)
			return
		}
		rawDataCompleted <- true
		rawData <- rawdata
	}()

	select {
	case <-shutdown:
		logWarn.Println("Forced to shut down")
		return nil, nil
	case <-rawDataCompleted:
		break
	}

	raw := <-rawData
	// Parse raw results
	var result json.RawMessage
	var isSucces = BasicReturn{
		Result: &result,
	}

	err = json.Unmarshal(raw, &isSucces)
	if err != nil {
		return nil, err
	}
	if isSucces.Ok != true {
		logErr.Printf("ERROR from Telegram API: %+v\n", isSucces.Result)
		return nil, errors.New("Telegram bot token is incorrect")
	}

	// Create Update objects
	updates := make([]Update, 100)
	err = json.Unmarshal(result, &updates)
	if err != nil {
		return nil, err
	}

	return updates, nil
}

func (t Tbot) sendMessage(chat_id int64, text string) error {
	var s struct {
		Chat_id int64  `json:"chat_id"`
		Text    string `json:"text"`
	}
	s.Chat_id = chat_id
	s.Text = text

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(s)

	r, err := http.Post(turl+t.apikey+"/sendMessage", "application/json; charset=utf-8", b)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	rdata, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var apierr APIError
	err = json.Unmarshal(rdata, &apierr)
	if err != nil {
		logWarn.Printf("Error decoding sent response: %v", err)
	}

	if apierr.Ok == false {
		logWarn.Printf("Error from Telegram API. Error code: %d. Description: %s", apierr.Error_code, apierr.Description)
		return errors.New("Error from Telegram API")
	}

	logInfo.Println(b.String())
	return nil
}
