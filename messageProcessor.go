package main

import "github.com/go-telegram-bot-api/telegram-bot-api"

type chatStats struct {
	name         string
	messageTotal int
	charTotal    int64
	people       map[int]person
}

type person struct {
	name      string
	msgcount  int
	charcount int64
}

var stats = make(map[int64]chatStats) // GLOBAL, I KNOW

func messageProcessor(g global) {
	defer g.wg.Done()

outer:
	for {
		select {
		case <-g.shutdown:
			break outer
		case msg := <-g.messages:
			processMessage(msg, stats)
			m := tgbotapi.NewMessage(msg.Chat.ID, "This is just a message")
			m.ReplyToMessageID = msg.MessageID
			g.bot.Send(m)
		}
	}

	logWarn.Println("Stopping message processor")
}

func processMessage(msg *tgbotapi.Message, stats map[int64]chatStats) {
	// Cherry pick the needed struct from the map
	i, ok := stats[msg.Chat.ID]
	if ok == false {
		i.name = msg.Chat.Title
		i.people = make(map[int]person)
	}

	i.messageTotal++
	i.charTotal += int64(len(msg.Text))

	// Cherry pick again
	p, ok := i.people[msg.From.ID]
	if ok == false {
		p.name = msg.From.String()
	}
	p.msgcount++
	p.charcount += int64(len(msg.Text))

	// Send the structs back to the maps
	i.people[msg.From.ID] = p
	stats[msg.Chat.ID] = i
}
