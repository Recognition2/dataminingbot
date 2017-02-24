package main

import (
	"database/sql"
	"time"
)

func dbTimer(g *global, clearStats chan bool) {
	defer g.wg.Done()
	defer logWarn.Println("Database connection closed")

	user := g.c.Mysql_user
	passwd := g.c.Mysql_passwd
	dbname := g.c.Mysql_dbname

	db, err := sql.Open("mysql", user+":"+passwd+"@unix(/run/mysqld/mysqld.sock)/"+dbname) // DOES NOT open a connection
	if err != nil {
		logErr.Println(err)
	}

	err = db.Ping() // Validating DSN data
	if err != nil {
		logErr.Printf("Failed opening db connection: %v\n", err)
		logWarn.Println("Running in memory")
		g.useDB = false
		return
	}
	defer db.Close()

	g.useDB = true
	g.db = db

	shouldShutDown := false
	for !shouldShutDown {
		// Get time channel
		timeToSync := time.After(time.Minute * 2)

		select {
		case <-g.shutdown:
			shouldShutDown = true
			// If shutdown, sync stats then quit
			writeToDb(db)

		case <-timeToSync:
			writeToDb(db)
			clearStats <- true
		}
	}
}

func writeToDb(db *sql.DB) {
	// First, add the chat
	// Secondly, add all people (and their statistics) and link it to that chat

	beforeFlush := time.Now()

	// Iterate over all chats
	// ci = chatinfo, has to be used 8 times...
	for chatID, ci := range stats {
		stmt, err := db.Prepare(`
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
			stmt, err := db.Prepare(`
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
			delete(stats[chatID].people, personID)
		}
	}

	afterFlush := time.Now()
	logWarn.Printf("Flushing to database completed, took %.3f seconds\n", afterFlush.Sub(beforeFlush).Seconds())
}
