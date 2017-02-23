package main

import (
	"database/sql"
	"time"
)

func dbTimer(g *global, clearStats chan bool, stats map[int64]*chatStats) {
	defer g.wg.Done()
	user := g.c.Mysql_user
	passwd := g.c.Mysql_passwd
	dbname := g.c.Mysql_dbname

	db, err := sql.Open("mysql", user+":"+passwd+"@/"+dbname) // DOES NOT open a connection
	if err != nil {
		logErr.Printf("Failed opening db connection: %v\n", err)
		g.useDB = false
		return
	}
	g.useDB = true
	defer db.Close()

	err = db.Ping() // Validating DSN data
	checkErr(err)

	g.db = db

	shouldShutDown := false
	for !shouldShutDown {
		// Get time channel
		timeToSync := time.After(time.Second * 20)

		select {
		case <-g.shutdown:
			shouldShutDown = true
			// If shutdown, sync stats then quit
			syncStats(db, stats, clearStats)
		case <-timeToSync:
			syncStats(db, stats, clearStats)
		}
	}

	logWarn.Println("Stopping database connection")
}

func syncStats(db *sql.DB, stats map[int64]*chatStats, clearStats chan bool) {
	writeToDb(db, stats)
	clearStats <- true
}

func writeToDb(db *sql.DB, stats map[int64]*chatStats) {
	// First, add the chat
	// Secondly, add all people (and their statistics) and link it to that chat

	logInfo.Println("Flushing stats...")
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
		checkErr(err)

		res, err := stmt.Exec(chatID, ci.name, ci.messageTotal, ci.charTotal, ci.Type /*Again*/, ci.name, ci.messageTotal, ci.charTotal, ci.Type)
		checkErr(err)
		logInfo.Println(res)

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
			checkErr(err)

			_, err = stmt.Exec(chatID, personID, pi.msgcount, pi.charcount, pi.name /*AGAIN*/, pi.msgcount, pi.charcount, pi.name)
			checkErr(err)
			delete(stats[chatID].people, personID)
		}
	}

	afterFlush := time.Now()
	logWarn.Printf("Flushing to database completed, took %f seconds\n", afterFlush.Sub(beforeFlush).Seconds())
}
