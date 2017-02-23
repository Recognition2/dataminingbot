// main.go
// Author: gregory at 24-01-2017

package main

import (
	"database/sql"
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Apikey       string // Telegram API key
	Admins       []string
	UsePolling   bool
	LogLevel     string // how much to log
	Mysql_user   string `json:"mysql_user"`
	Mysql_passwd string `json:"mysql_passwd"`
	Mysql_dbname string `json:"mysql_dbname"`
}

type global struct {
	wg       *sync.WaitGroup  // For checking that everything has indeed shut down
	shutdown chan bool        // To make sure everything can shut down
	bot      *tgbotapi.BotAPI // The actual bot
	c        Config
	messages chan *tgbotapi.Message
	db       *sql.DB
	useDB    bool
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

	// Create the waitgroup
	var wg sync.WaitGroup
	g := global{
		wg:       &wg,
		shutdown: shouldShutdown,
		bot:      bot,
		c:        c,
	}

	// Statistics object
	var stats = make(map[int64]*chatStats)

	// Start processing messages
	g.messages = make(chan *tgbotapi.Message, 100)
	clearStats := make(chan bool)
	wg.Add(1)
	go messageProcessor(&g, clearStats, stats)

	// Start message monitor
	wg.Add(1)
	go messageMonitor(&g, stats) // Monitor messages

	// Start the database connection
	wg.Add(1)
	go dbTimer(&g, clearStats, stats)

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

func messageMonitor(g *global, stats map[int64]*chatStats) {
	logInfo.Println("Starting message monitor")
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

			if update.Message.IsCommand() {
				commandHandler(g, update.Message, stats)
			} else {
				// Message is no command, handle it
				g.messages <- update.Message
			}
		}
	}

	logWarn.Println("Stopping message monitor")
}

func checkErr(e error) {
	if e != nil {
		logErr.Println(e)
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
