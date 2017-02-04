// Command.go
package main

import (
	"gopkg.in/telegram-bot-api.v4"
)

func handleHi(bot tgapi.BotAPI) {
	msg := tgbotapi.NewMessage(cmd.Chat.ID, "Hello to you too!")
	bot.Send(msg)
}
