package main

import (
	"database/sql"
	"time"
)

func dbTimer() {
	defer Global.wg.Done()

	user := Global.config.Mysql_user
	passwd := Global.config.Mysql_passwd
	dbname := Global.config.Mysql_dbname

	db, err := sql.Open("mysql", user+":"+passwd+"@/"+dbname) // DOES NOT open a connection
	if err != nil {
		logErr.Println(err)
	}

	err = db.Ping() // Validating DSN data
	if err != nil {
		logErr.Printf("Failed opening db connection: %v\n", err)
		logWarn.Println("Running in memory")
		Global.useDB = false
		return
	}
	defer db.Close()

	Global.useDB = true
	Global.db = db
	logInfo.Println("Database connection opened!")
	defer logWarn.Println("Database connection closed")

outer:
	for {
		// Get time channel
		timeToSync := time.After(time.Second * 100)

		select {
		case <-Global.shutdown:
			// If shutdown, sync stats then quit
			writeToDb()
			break outer

		case <-timeToSync:
			writeToDb()
		}
	}
}

func writeToDb() {
	beforeFlush := time.Now()

	// First, add the chat
	// Secondly, add all people (and their statistics) and link it to that chat

	// Iterate over all chats
	// ci = chatinfo, has to be used 8 times...
	for chatID, ci := range Global.stats {
		stmt, err := Global.db.Prepare(`
			INSERT INTO
				chats (chatid, name, messageTotal, charTotal, Type)
				VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				name = ?,
				messageTotal = messageTotal + ?,
				charTotal = charTotal + ?,
				Type = ?`)
		if err != nil {
			logErr.Println(err)
		}

		_, err = stmt.Exec(chatID, ci.name, ci.messageTotal, ci.charTotal, ci.Type /*Again*/, ci.name, ci.messageTotal, ci.charTotal, ci.Type)
		if err != nil {
			logErr.Println(err)
		}

		for personID, pi := range ci.people {
			stmt, err := Global.db.Prepare(`
			INSERT INTO
				personstats (chatfk, personid, msgcount, charcount, name)
				VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				msgcount = msgcount + ?,
				charcount = charcount + ?,
				name = ?
				`)
			if err != nil {
				logErr.Println(err)
			}

			_, err = stmt.Exec(chatID, personID, pi.msgcount, pi.charcount, pi.name /*AGAIN*/, pi.msgcount, pi.charcount, pi.name)
			if err != nil {
				logErr.Println(err)
			}
		}
	}
	// Clear stats object
	Global.statsLock.Lock()
	Global.stats = make(map[int64]chatStats) // LOCK
	Global.statsLock.Unlock()

	timeAfter := time.Now()
	logWarn.Printf("Flushing to database completed, took %.5f seconds\n", timeAfter.Sub(beforeFlush).Seconds())
}
