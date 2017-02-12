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
	case "stats":
		handleStats(g.bot, cmd.Chat.ID)
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

const statsBeginStub = `%s
*Name: Messages: Characters:*
`

const statsLoop = `%d. %s: *%d*; %d
`

const statsEnd = `
_Total_: *%d*; %d
`

const statsTime = `_%s_`

func handleStats(bot *tgbotapi.BotAPI, from int64) {
	logInfo.Println("Printing statistics")
	var b bytes.Buffer
	chat := stats[from]
	b.WriteString(fmt.Sprintf(statsBeginStub, chat.name))

	for i, j := range chat.people {
		b.WriteString(fmt.Sprintf(statsLoop, i, j.name, j.msgcount, j.charcount))
	}

	b.WriteString(fmt.Sprintf(statsEnd, chat.messageTotal, chat.charTotal))

	curr := time.Now().String()[:20]
	b.WriteString(fmt.Sprintf(statsTime, curr))

	m := tgbotapi.NewMessage(from, b.String())
	m.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(m)
	if err != nil {
		logErr.Println(err)
	}
}
