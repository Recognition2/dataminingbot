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
	wg        *sync.WaitGroup     // For checking that everything has indeed shut down
	shutdown  chan bool           // To make sure everything can shut down
	bot       *tgbotapi.BotAPI    // The actual bot
	config    Config              // Configuration file
	db        *sql.DB             // Database connection
	useDB     bool                // Use a database connection, or just run in memory
	stats     map[int64]chatStats // statistics global object
	statsLock *sync.RWMutex       // Lock for stats object
}

// Global variables
// Loggers:
var logErr = log.New(os.Stderr, "[ERRO] ", log.Ldate+log.Ltime+log.Ltime+log.Lshortfile)
var logWarn = log.New(os.Stdout, "[WARN] ", log.Ldate+log.Ltime)
var logInfo = log.New(os.Stdout, "[INFO] ", log.Ldate+log.Ltime)

var Global = global{
	shutdown: make(chan bool),
	stats:    make(map[int64]chatStats),
}

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
	Global.config = c

	Global.bot, err = tgbotapi.NewBotAPI(c.Apikey)
	if err != nil {
		logErr.Println(err)
	}

	// Create the waitgroup
	var wg sync.WaitGroup
	Global.wg = &wg

	// Create the stats lock
	var statsLock sync.RWMutex
	Global.statsLock = &statsLock

	// Start processing messages
	messages := make(chan *tgbotapi.Message, 100)
	var n int = 4
	Global.wg.Add(n - 1)
	for i := 0; i < n; i++ {
		go messageProcessor(i, messages) // Start multiple message processors
	}

	// Start message monitor
	Global.wg.Add(1)
	go messageMonitor(messages) // Monitor messages

	// Start the database connection
	Global.wg.Add(1)
	go dbTimer()

	// Wait for SIGINT or SIGTERM, then quit
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT)

	logInfo.Println("All routines have been started, awaiting kill signal")

	// Program will hang here, probably forever
	<-sigs
	println()
	logInfo.Println("Shutdown signal received, waiting for goroutines")
	close(Global.shutdown)
	// Shutdown initiated, waiting for all goroutines to shut down
	Global.wg.Wait()
	logWarn.Println("Shutting down")
	return 0
}

func contains(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
