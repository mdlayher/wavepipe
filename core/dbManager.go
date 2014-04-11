package core

import (
	"database/sql"
	"log"

	"github.com/astaxie/beedb"
	_ "github.com/mattn/go-sqlite3"
)

// dbLink is a database/sql link to the backend, used to work with the ORM
var dbLink *sql.DB

// orm() returns the instance of the beedb ORM using the open database link
func orm() beedb.Model {
	return beedb.New(dbLink)
}

// dbManager handles the database connection pool, and communicates back and forth with the manager goroutine
func dbManager(dbPath string, dbKillChan chan struct{}) {
	log.Println("db: starting...")

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Println(err)
		return
	}

	// Store database connection
	dbLink = db

	// Trigger events via channel
	for {
		select {
		// Stop database manager
		case <-dbKillChan:
			// Inform manager that shutdown is complete
			log.Println("db: stopped!")
			dbKillChan <- struct{}{}
			return
		}
	}
}
