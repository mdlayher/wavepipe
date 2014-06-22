package core

import (
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/mdlayher/wavepipe/config"
	"github.com/mdlayher/wavepipe/data"
)

// dbManager manages database connections, and communicates back and forth with the manager goroutine
func dbManager(conf config.Config, dbLaunchChan chan struct{}, dbKillChan chan struct{}) {
	log.Println("db: starting...")

	// Attempt to open database connection, depending on configuration
	// sqlite
	if conf.Sqlite != nil {
		log.Println("db: sqlite:", conf.Sqlite.File)

		// Get current user
		user, err := user.Current()
		if err != nil {
			log.Fatalf("db: could not get current user: %s", err.Error())
		}

		// Replace the home character to set path
		path := strings.Replace(conf.Sqlite.File, "~", user.HomeDir, -1)

		// Set DSN
		data.DB = new(data.SqliteBackend)
		data.DB.DSN(path)

		// Set up the database
		if err := data.DB.Setup(); err != nil {
			log.Fatalf("db: could not set up database: %s", err.Error())
		}

		// Verify database file exists and is ready
		if _, err := os.Stat(path); err != nil {
			log.Fatalf("db: database file does not exist: %s", conf.Sqlite.File)
		}

		// Open the database connection
		if err := data.DB.Open(); err != nil {
			log.Fatalf("db: could not open database: %s", err)
		}

		// TODO: temporary, create a test user
		data.NewUser("test", "test", data.RoleAdmin)
	} else {
		// Invalid config
		log.Fatalf("db: invalid database selected")
	}

	// Database set up, trigger manager that it's ready
	close(dbLaunchChan)

	// Trigger events via channel
	for {
		select {
		// Stop database manager
		case <-dbKillChan:
			// Close the database connection pool
			if err := data.DB.Close(); err != nil {
				log.Fatalf("db: could not close connection")
			}

			// Inform manager that shutdown is complete
			log.Println("db: stopped!")
			dbKillChan <- struct{}{}
			return
		}
	}
}
