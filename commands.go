// Command.go
package main

import (
	"bytes"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"sort"
	"strconv"
	"time"
)

func commandHandler(cmd *tgbotapi.Message) {

	switch cmd.Command() {
	case "hi":
		handleHi(cmd)
	case "hallodaar":
		handleHallodaar(cmd)
	case "ping":
		handlePing(cmd)
	case "stats":
		handleStats(cmd)
	case "time":
		handleTime(cmd)
	case "kutbot":
		handleInsults(cmd)
	default:
		if contains(strconv.Itoa(cmd.From.ID), Global.config.Admins) {
			_ = adminCommandHandler(cmd)
		}
	}
}

func handleHi(cmd *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(cmd.Chat.ID, "Hello to *you* too!")
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := Global.bot.Send(msg)
	if err != nil {
		logErr.Println(err)
	}
}

func handleHallodaar(cmd *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(cmd.Chat.ID, "Hai _schat_")
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := Global.bot.Send(msg)
	if err != nil {
		logErr.Println(err)
	}
}

func handlePing(cmd *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(cmd.Chat.ID, "Pong")
	_, err := Global.bot.Send(msg)
	if err != nil {
		logErr.Println(err)
	}
}

func handleInsults(cmd *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(cmd.Chat.ID, "No, you, you little *bitch* fucktard assfuck, what the fuck do you think you are doing with the fucking fuck?")
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyToMessageID = cmd.MessageID

	_, err := Global.bot.Send(msg)
	if err != nil {
		logErr.Println(err)
	}

}

func handleTime(cmd *tgbotapi.Message) {
	t1 := time.Now()
	msg1 := tgbotapi.NewMessage(cmd.Chat.ID, "First message")
	msg1.ReplyToMessageID = cmd.MessageID
	_, err := Global.bot.Send(msg1)
	if err != nil {
		logErr.Println(err)
	}

	t2 := time.Now()

	timeDiff := t2.Sub(t1)
	msg2 := tgbotapi.NewMessage(cmd.Chat.ID, fmt.Sprintf("Time difference: %.1f ms", timeDiff.Seconds()*1000))
	msg2.ReplyToMessageID = cmd.MessageID
	_, err = Global.bot.Send(msg2)
	if err != nil {
		logErr.Println(err)
	}
}

func handleStats(cmd *tgbotapi.Message) {
	Global.statsLock.RLock()
	memstats := Global.stats[cmd.Chat.ID]
	logInfo.Printf("%#v", Global.stats)
	Global.statsLock.RUnlock()

	thisChat := chatStats{}
	thisChat.people = make(map[int]personStats)
	if Global.useDB {
		// get data from db
		thisChat = getStatsFromDB(cmd.Chat.ID)
	}

	// Add data that is currently in memory
	thisChat.messageTotal += memstats.messageTotal
	thisChat.charTotal += memstats.charTotal

	for i, c := range memstats.people {
		var thisperson personStats
		thisperson.name = c.name
		thisperson.charcount = thisChat.people[i].charcount + c.charcount
		thisperson.msgcount = thisChat.people[i].msgcount + c.msgcount
		thisChat.people[i] = thisperson
	}

	// Results have been fetched, create the message
	var b bytes.Buffer
	b.WriteString("Message count, character count\n")

	var cname string
	switch thisChat.Type {
	case "private":
		logInfo.Println("Private chat statistics requested")
		cname = "Private chat"
	case "group":
		fallthrough
	case "supergroup":
		logInfo.Printf("Statistics requested by %v\n", cmd.From.String())
		cname = thisChat.name
	}
	b.WriteString(fmt.Sprintf("*%s*\n", cname))

	// Sort people by messagecount
	var keys []int
	for k := range thisChat.people {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// Iterate over the people in the chat
	for _, l := range keys {
		j := thisChat.people[l]
		b.WriteString(fmt.Sprintf("%s: *%d*; %d\n", j.name, j.msgcount, j.charcount))
	}

	b.WriteString(fmt.Sprintf("\n_Total_: *%d*; %d\n", thisChat.messageTotal, thisChat.charTotal))

	curr := time.Now().String()[:19]
	b.WriteString(fmt.Sprintf("%s", curr))

	m := tgbotapi.NewMessage(cmd.Chat.ID, b.String())
	m.ParseMode = tgbotapi.ModeMarkdown
	_, err := Global.bot.Send(m)
	if err != nil {
		logErr.Println(err)
	}
}

func getStatsFromDB(chatid int64) chatStats {
	c := chatStats{}
	c.people = make(map[int]personStats)
	// get thisChatinfo
	getChatInfo(&c, chatid)

	// get person information
	getPersonInfo(&c, chatid)

	return c
}

func getChatInfo(c *chatStats, chatid int64) {
	chatinfo, err := Global.db.Query(`SELECT name, messageTotal, charTotal, Type FROM chats WHERE chatid=? LIMIT 1`, chatid)
	if err != nil {
		logErr.Println(err)
	}
	defer chatinfo.Close()

	var charTotal int64
	var msgTotal int
	var name, Type string

	for chatinfo.Next() {
		err = chatinfo.Scan(&name, &msgTotal, &charTotal, &Type)
		if err != nil {
			logErr.Println(err)
		}

		c.Type = Type
		c.charTotal = charTotal
		c.messageTotal = msgTotal
		c.name = name
	}
	return
}

func getPersonInfo(c *chatStats, chatfk int64) {
	rows, err := Global.db.Query(`SELECT name, personid, msgcount, charcount FROM personstats WHERE chatfk = ?`, chatfk)
	if err != nil {
		logErr.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var charcount int64
		var msgcount, personid int

		err = rows.Scan(&name, &personid, &msgcount, &charcount)
		if err != nil {
			logErr.Println(err)
		}

		p := personStats{
			name:      name,
			msgcount:  msgcount,
			charcount: charcount,
		}
		c.people[personid] = p
	}
}
