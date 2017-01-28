package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
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
	Update_id int64   `json:"update_id"`         // Unique identifier
	Message   Message `json:"message,omitempty"` // OPTIONAL:	Message that was sent
}

type Message struct {
	Message_id int64  // Unique identifier
	From       User   // OPTIONAL: May be empty for channels
	Date       int    // Date message was sent, Unix time
	Chat       Chat   // Conversation message belongs to
	Text       string // Raw UTF-8 text of message
}

type Chat struct {
	Id       int64  // Unique identifier, max 52 bits.
	Type     string // One of: "private", "group", “supergroup” or “channel”
	Title    string // OPTIONAL
	Username string //OPTIONAL: Not for groups.
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

type sendMessage struct {
	Chat_id int64  `json:"chat_id"`
	Text    string `json:"text"`
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
	timeout := 120
	var arg = "offset=" + strconv.FormatInt(offset, 10) + "&timeout=" + strconv.Itoa(timeout)
	httpResponse, err := http.Get(turl + t.apikey + "/getUpdates?" + arg)
	// http.Get has an inherent delay, while doing DNS lookup etc. (guaranteeing that your connection is valid)
	if err != nil {
		return nil, err
	}

	rawData := make(chan []byte)
	rawDataCompleted := make(chan bool)

	go func() {
		defer httpResponse.Body.Close()
		rawdata, err := ioutil.ReadAll(httpResponse.Body) // Read stream
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
