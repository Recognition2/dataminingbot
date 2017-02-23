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
	people       map[int]personStats
}

type personStats struct {
	name      string
	msgcount  int
	charcount int64
}

func messageProcessor(g *global, clearStats <-chan bool, stats map[int64]*chatStats) {
	defer g.wg.Done()
	logInfo.Println("Starting message processor")
	defer logWarn.Println("Stopping message processor")

outer:
	for {
		select {
		case <-clearStats:
			stats = make(map[int64]*chatStats)
		case <-g.shutdown:
			break outer
		case msg := <-g.messages:
			processMessage(msg, stats)
		}
	}

}

func processMessage(msg *tgbotapi.Message, stats map[int64]*chatStats) {
	// Cherry pick the needed struct from the map
	i, ok := stats[msg.Chat.ID]
	if ok == false || i.name != msg.Chat.Title {
		i.name = msg.Chat.Title
		i.people = make(map[int]personStats)
		i.Type = msg.Chat.Type
	}

	i.messageTotal++
	i.charTotal += int64(len(msg.Text))

	// Cherry pick again
	p, ok := i.people[msg.From.ID]
	personname := fmt.Sprintf("%s %s", msg.From.FirstName, msg.From.LastName)
	if ok == false || p.name != personname {
		p.name = personname
	}
	p.msgcount++
	p.charcount += int64(len(msg.Text))

	// Send the structs back to the maps
	i.people[msg.From.ID] = p
	stats[msg.Chat.ID] = i
}
