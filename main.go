// main.go
// Author: gregory at 24-01-2017

package main

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"gopkg.in/telegram-bot-api.v4"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Apikey     string // Telegram API key
	Admins     []string
	UsePolling bool
	LogLevel   string // how much to log
}

type global struct {
	wg       *sync.WaitGroup  // For checking that everything has indeed shut down
	shutdown chan bool        // To make sure everything can shut down
	bot      *tgbotapi.BotAPI // The actual bot
	c        Config
}

// Global variables
// Loggers:
var logErr = log.New(os.Stderr, "[ERRO] ", log.Ldate+log.Ltime+log.Ltime)
var logWarn = log.New(os.Stdout, "[WARN] ", log.Ldate+log.Ltime)
var logInfo = log.New(os.Stdout, "[INFO] ", log.Ldate+log.Ltime)

func main() {
	os.Exit(mainExitCode())
}

func mainExitCode() int {
	// Create logging objects

	// Parse bot configuration
	var c Config
	_, err := toml.DecodeFile("settings.toml", &c)
	if err != nil {
		logErr.Println(err)
	}

	switch c.LogLevel {
	case "error":
		logWarn.SetFlags(0)
		logWarn.SetOutput(ioutil.Discard)
		fallthrough
	case "warn":
		logInfo.SetFlags(0)
		logInfo.SetOutput(ioutil.Discard)
		fallthrough
	case "info":
	default:
		logErr.Println("No valid logLevel directive in configuration file")
		return 1
	}
	logInfo.Println("Config file parsed")
	logWarn.Println("This is an example of a warning")
	logErr.Println("This is an example of an error")

	bot, err := tgbotapi.NewBotAPI(c.Apikey)
	if err != nil {
		logErr.Println(err)
	}

	shouldShutdown := make(chan bool)

	// Start the waitgroup
	var wg sync.WaitGroup
	g := global{
		wg:       &wg,
		shutdown: shouldShutdown,
		bot:      bot,
		c:        c,
	}

	// Start message monitor
	wg.Add(1)
	go messageMonitor(g, toSend) // Monitor messages

	// Wait for SIGINT or SIGTERM, then quit
	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT)
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigs // This goroutine will hang here until interrupt is sent
		println()
		logInfo.Println("Shutdown signal received, waiting for goroutines")
		close(shouldShutdown)
		// Await all other goroutines, then send Done
		done <- true
	}()

	logInfo.Println("All routines have been started, awaiting kill signal")

	// Program will hang here, probably forever
	<-done
	// Shutdown initiated, waiting for all goroutines to shut down
	wg.Wait()
	logInfo.Println("Shutting down")
	return 0
}

func messageMonitor(g global, toSend chan tgbotapi.Chattable) {
	defer g.wg.Done()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 15
	updates, err := g.bot.GetUpdatesChan(u)
	if err != nil {
		logErr.Printf("Update failed: %v\n", err)
	}

outer:
	for {
		select {
		case <-g.shutdown:
			break outer
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			logInfo.Printf("Message received from %s: '%s'", update.Message.From.UserName, update.Message.Text)

			if update.Message.IsCommand() {
				g.wg.Add(1)
				commandHandler(g, update.Message, toSend)
			}

		}
	}

	logWarn.Println("Stopping message monitor")
}

func commandHandler(g global, cmd *tgbotapi.Message, send chan tgbotapi.Chattable) {
	defer g.wg.Done()

	switch cmd.Command() {
	case "hi":
		handleHi(g.bot)
	default:
		if contains(string(cmd.From.ID), g.c.Admins) {
			adminCommandHandler(g, *cmd)
		}
	}
}

func adminCommandHandler(g global, cmd tgbotapi.Message) {
	// These commands are only available if:
	// - You're in a private chat
	// - You're in a group chat, but you specify "override"
	logInfo.Printf("Admin command '%s' requested\n", cmd.Command())

	switch cmd.Command() {
	case "load":
		handleLoad(g.bot)
	}
}

func contains(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
