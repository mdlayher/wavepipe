package core

import (
	"log"

	"github.com/mdlayher/wavepipe/data"
)

// dbManager manages database connections, and communicates back and forth with the manager goroutine
func dbManager(dbPath string, dbKillChan chan struct{}) {
	log.Println("db: starting...")

	// Attempt to open database connection
	if dbPath != "" {
		data.DB = new(data.SqliteBackend)
		data.DB.DSN(dbPath)
	}

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
