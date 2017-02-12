package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type chatStats struct {
	name         string
	messageTotal int
	charTotal    int64
	Type         string
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
		i.Type = msg.Chat.Type
	}

	i.messageTotal++
	i.charTotal += int64(len(msg.Text))

	// Cherry pick again
	p, ok := i.people[msg.From.ID]
	if ok == false {
		p.name = fmt.Sprintf("%s %s", msg.From.FirstName, msg.From.LastName)
	}
	p.msgcount++
	p.charcount += int64(len(msg.Text))

	// Send the structs back to the maps
	i.people[msg.From.ID] = p
	stats[msg.Chat.ID] = i
}
