package main

import "github.com/go-telegram-bot-api/telegram-bot-api"

func messageMonitor(messages chan *tgbotapi.Message) {
	defer Global.wg.Done()
	logInfo.Println("Starting message monitor")
	defer logWarn.Println("Stopping message monitor")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 120
	updates, err := Global.bot.GetUpdatesChan(u)
	if err != nil {
		logErr.Printf("Update failed: %v\n", err)
	}

outer:
	for {
		select {
		case <-Global.shutdown:
			break outer

		case update := <-updates:
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() {
				commandHandler(update.Message)
			} else if update.Message.Text == "hi" {
				handleHi(update.Message)
			} else {
				// Message is no command, handle it
				messages <- update.Message
			}
		}
	}
}
