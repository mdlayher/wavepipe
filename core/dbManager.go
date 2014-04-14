package core

import (
	"log"

	"github.com/mdlayher/wavepipe/config"
	"github.com/mdlayher/wavepipe/data"
)

// dbManager manages database connections, and communicates back and forth with the manager goroutine
func dbManager(conf config.Config, dbKillChan chan struct{}) {
	log.Println("db: starting...")

	// Attempt to open database connection, depending on configuration
	// sqlite
	if conf.Sqlite != nil {
		log.Println("db: sqlite:", conf.Sqlite.File)
		data.DB = new(data.SqliteBackend)
		data.DB.DSN(conf.Sqlite.File)

		// Set up the database
		if err := data.DB.Setup(); err != nil {
			log.Fatalf("db: could not set up database: %s", err.Error())
		}
	} else {
		// Invalid config
		log.Fatalf("db: invalid database selected")
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
