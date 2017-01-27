// main.go
// Author: gregory at 24-01-2017

package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Apikey     string // Telegram API key
	Admins     []string
	UsePolling bool
	LogLevel   string // how much to log
}

// Global variables
// Loggers:
var logErr = log.New(os.Stderr, "[ERR] ", log.Ldate+log.Ltime+log.Ltime)
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
	bot := Tbot{apikey: c.Apikey}
	bot.Username, err = bot.getUsername()
	if err != nil {
		logErr.Print(err)
	}

	shouldShutdown := make(chan bool)

	// Start the waitgroup
	var wg sync.WaitGroup

	// Start all async operations
	wg.Add(1)
	go messageMonitor(wg, bot, shouldShutdown)

	// Wait for SIGINT or SIGTERM, then quit
	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT)
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigs // This goroutine will hang here until interrupt is sent
		logInfo.Println("Shutdown signal received, waiting for goroutines")
		close(shouldShutdown)
		// Await all other goroutines, then send Done
		done <- true
	}()

	logInfo.Println("All routines have been started, awaiting kill signal")

	// Program will hang here, probably forever
	<-done
	// Shutdown initiated, waiting for all goroutines to shut down
	logInfo.Println("Waiting for async operations to stop...")
	wg.Wait()
	logInfo.Println("Everything has shut down, stopping now")
	return 0
}

func messageMonitor(wg sync.WaitGroup, bot Tbot, shutdown chan bool) {
	var offset int64 = 0

outer:
	for {
		select {
		case <-shutdown:
			break outer
		default:
		}
		updates, err := bot.getUpdates(offset, shutdown)
		if err != nil {
			logWarn.Print(err)
		}
		for _, k := range updates {
			fmt.Printf("%#v\n", k)
			if k.Update_id > offset {
				offset = k.Update_id
				go handleMessage(k.Message, bot)
			}
		}
		time.Sleep(time.Second * 10)
	}
	logInfo.Println("Stopping message monitor")
	wg.Done()
}

func handleMessage(m Message, bot Tbot) {
	logInfo.Println("Handling message")
	raw := url.Values{}
	raw.Set("chat_id", strconv.FormatInt(m.Chat.Id, 10))
	raw.Add("text", "Hoi. You sent: "+m.Text)

	http.Post(turl+bot.apikey+"/sendMessage", raw.Encode(), nil)
}
