// Commands that are only available for admins

package main

import (
	"bytes"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"os/exec"
)

func adminCommandHandler(cmd *tgbotapi.Message) bool {
	// These commands are only available if:
	// - You're in a private chat
	// - You're in a group chat, but you specify "override"
	logInfo.Printf("Admin command '%s' requested\n", cmd.Command())

	switch cmd.Command() {
	case "load":
		handleLoad(cmd)
	case "uptime":
		handleUptime(cmd)
	default:
		return true // Is clearly not an admin command
	}
	return false
}

func handleLoad(cmd *tgbotapi.Message) {
	w := exec.Command("uptime")
	var out bytes.Buffer
	w.Stdout = &out

	err := w.Run()
	if err != nil {
		logErr.Println(err)
	}
	uptime := out.String()
	load := uptime[len(uptime)-30:]
	// Format: 19:02:27 up  4:47,  1 user,  load average: 0.21, 0.43, 0.45

	msg := tgbotapi.NewMessage(cmd.Chat.ID, "L"+load)
	Global.bot.Send(msg)
}

func handleUptime(cmd *tgbotapi.Message) {
	uptime := exec.Command("uptime", "-p")
	var out bytes.Buffer
	uptime.Stdout = &out

	err := uptime.Run()
	if err != nil {
		logErr.Println(err)
	}

	msg := tgbotapi.NewMessage(cmd.Chat.ID, out.String())
	Global.bot.Send(msg)
}
