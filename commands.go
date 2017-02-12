// Command.go
package main

import (
	"bytes"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"time"
)

func commandHandler(g global, cmd *tgbotapi.Message) {
	defer g.wg.Done()

	switch cmd.Command() {
	case "hi":
		handleHi(g.bot, cmd)
	case "ping":
		handlePing(g.bot, cmd)

	case "stats":
		handleStats(g.bot, cmd)
	case "time":
		handleTime(g.bot, cmd)
	default:
		if contains(strconv.Itoa(cmd.From.ID), g.c.Admins) {
			_ = adminCommandHandler(g, cmd)
		}

	}
}

func handleHi(bot *tgbotapi.BotAPI, cmd *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(cmd.Chat.ID, "Hello to *you* too!")
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(msg)
	if err != nil {
		logErr.Println(err)
	}
}

func handlePing(bot *tgbotapi.BotAPI, cmd *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(cmd.Chat.ID, "Pong")
	_, err := bot.Send(msg)
	if err != nil {
		logErr.Println(err)
	}
}

func handleTime(bot *tgbotapi.BotAPI, cmd *tgbotapi.Message) {
	t1 := time.Now()
	msg1 := tgbotapi.NewMessage(cmd.Chat.ID, "First message")
	msg1.ReplyToMessageID = cmd.MessageID
	_, err := bot.Send(msg1)
	if err != nil {
		logErr.Println(err)
	}

	t2 := time.Now()

	timeDiff := t2.Sub(t1)
	msg2 := tgbotapi.NewMessage(cmd.Chat.ID, fmt.Sprintf("Time difference: %f seconds", timeDiff.Seconds()))
	msg2.ReplyToMessageID = cmd.MessageID
	_, err = bot.Send(msg2)
	if err != nil {
		logErr.Println(err)
	}
}

func handleStats(bot *tgbotapi.BotAPI, cmd *tgbotapi.Message) {
	var b bytes.Buffer
	chat := stats[cmd.Chat.ID]

	b.WriteString("Message count, character count\n")

	var cname string
	switch chat.Type {
	case "private":
		logInfo.Println("Private chat statistics requested")
		cname = "Private chat"
	case "group":
		fallthrough
	case "supergroup":
		logInfo.Printf("Statistics requested by %v\n", cmd.From.String())
		cname = chat.name
	}
	b.WriteString(fmt.Sprintf("*%s*\n", cname))

	for _, j := range chat.people {
		b.WriteString(fmt.Sprintf("%s: *%d*; %d\n", j.name, j.msgcount, j.charcount))
	}

	b.WriteString(fmt.Sprintf("\n_Total_: *%d*; %d\n", chat.messageTotal, chat.charTotal))

	curr := time.Now().String()[:19]
	b.WriteString(fmt.Sprintf("%s", curr))

	m := tgbotapi.NewMessage(cmd.Chat.ID, b.String())
	m.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(m)
	if err != nil {
		logErr.Println(err)
	}

}
