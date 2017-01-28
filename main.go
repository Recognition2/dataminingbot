// main.go
// Author: gregory at 24-01-2017

package main

import (
	"bytes"
	"encoding/json"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
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
		return 2
	}

	shouldShutdown := make(chan bool)

	// Start the waitgroup
	var wg sync.WaitGroup

	// Start all async operations
	wg.Add(1)
	go messageMonitor(&wg, bot, shouldShutdown)

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
	logInfo.Println("Shutdown successfull")
	return 0
}

func messageMonitor(wg *sync.WaitGroup, bot Tbot, shutdown chan bool) {
	defer wg.Done()
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
		logInfo.Printf("Updates received: %+v\n", updates)
		for _, k := range updates {
			if k.Update_id > offset {
				offset = k.Update_id
			}

			if k.Update_id != 0 {
				wg.Add(1)
				go handleMessage(k.Message, bot, wg)
			}
		}
		time.Sleep(time.Second * 5) // REMOVE THIS
	}
	logInfo.Println("Stopping message monitor")
}

func handleMessage(m Message, bot Tbot, wg *sync.WaitGroup) {
	defer wg.Done()
	logInfo.Println("Handling message")

	s := sendMessage{
		Chat_id: m.Chat.Id,
		Text:    "Hoi. You sent: " + m.Text,
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(s)
	r, err := http.Post(turl+bot.apikey+"/sendMessage", "application/json; charset=utf-8", b)
	if err != nil {
		logErr.Println(err)
		return
	}

	defer r.Body.Close()
	rdata, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logErr.Println(err)
	}

	var apierr APIError
	err = json.Unmarshal(rdata, &apierr)
	if err != nil {
		logWarn.Printf("Error decoding sent response: %v", err)
	}

	if apierr.Ok == false {
		logWarn.Printf("Error from Telegram API. Error code: %d. Description: %s", apierr.Error_code, apierr.Description)
		return
	}

	logInfo.Println(b.String())

}
