// Commands that are only available for admins

package main

import (
	"os/exec"
)

func handleLoad(bot tgbotapi.BotAPI) {
	w := exec.Command("w")
	var out bytes.Buffer
	w.Stdout = &out
	err := w.Run()
	if err != nil {
		logErr.Println(err)
	}
	logInfo.Printf("Load is: %v\n", out.String())
	msg := tgbotapi.NewMessage(cmd.Chat.ID, out.String())
	bot.Send(msg)

}
