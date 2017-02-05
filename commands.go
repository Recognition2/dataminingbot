// Command.go
package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"strconv"
)

func commandHandler(g global, cmd *tgbotapi.Message) {
	defer g.wg.Done()

	switch cmd.Command() {
	case "hi":
		handleHi(g.bot, cmd)
	default:
		if contains(strconv.Itoa(cmd.From.ID), g.c.Admins) {
			adminCommandHandler(g, cmd)
		}
	}
}

func handleHi(bot *tgbotapi.BotAPI, cmd *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(cmd.Chat.ID, "Hello to you too!")
	bot.Send(msg)
}
