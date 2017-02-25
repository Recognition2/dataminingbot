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

func messageProcessor(id int, messages chan *tgbotapi.Message) {
	defer Global.wg.Done()
	logInfo.Printf("Starting message processor %d\n", id)
	defer logWarn.Printf("Stopping message processor %d\n", id)

outer:
	for {
		select {
		case <-Global.shutdown:
			break outer
		case msg := <-messages:
			processMessage(msg)
		}
	}

}

func processMessage(msg *tgbotapi.Message) {
	// Cherry pick the needed struct from the map
	Global.statsLock.Lock()
	i, ok := Global.stats[msg.Chat.ID] // USE LOCKS

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
	Global.stats[msg.Chat.ID] = i // USE LOCKS
	Global.statsLock.Unlock()
}
