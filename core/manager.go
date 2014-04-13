package core

import (
	"log"
	"time"
)

// App is the application's name
const App = "wavepipe"

// Version is the application's version
const Version = "git-master"

// DBPath is the path to the sqlite3 database
// TODO: remove this for config
var DBPath string

// MediaFolder is the folder which we will recursively scan for media
// TODO: remove this for config
var MediaFolder string

// StartTime is the application's starting UNIX timestamp
var StartTime = time.Now().Unix()

// Manager is responsible for coordinating the application
func Manager(killChan chan struct{}, exitChan chan int) {
	log.Printf("manager: initializing %s %s...", App, Version)

	// Gather information about the operating system
	stat, err := OSInfo()
	if err != nil {
		log.Println("manager: could not get operating system info:", err)
	} else {
		log.Printf("manager: %s - %s_%s (%d CPU) [pid: %d]", stat.Hostname, stat.Platform, stat.Architecture, stat.NumCPU, stat.PID)
	}

	// Launch database manager to handle database/ORM connections
	dbKillChan := make(chan struct{})
	go dbManager(DBPath, dbKillChan)

	// Launch cron manager to handle timed events
	cronKillChan := make(chan struct{})
	go cronManager(cronKillChan)

	// Launch filesystem manager to handle file scanning
	// TODO: make this an actual path later on via configuration
	fsKillChan := make(chan struct{})
	go fsManager(MediaFolder, fsKillChan)

	// Launch HTTP API server
	apiKillChan := make(chan struct{})
	go apiRouter(apiKillChan)

	// Wait for termination signal
	for {
		select {
		// Trigger a graceful shutdown
		case <-killChan:
			log.Println("manager: triggering graceful shutdown, press Ctrl+C again to force halt")

			// Stop API, wait for confirmation
			apiKillChan <- struct{}{}
			<-apiKillChan
			close(apiKillChan)

			// Stop filesystem, wait for confirmation
			fsKillChan <- struct{}{}
			<-fsKillChan
			close(fsKillChan)

			// Stop database, wait for confirmation
			dbKillChan <- struct{}{}
			<-dbKillChan
			close(dbKillChan)

			// Stop cron, wait for confirmation
			cronKillChan <- struct{}{}
			<-cronKillChan
			close(cronKillChan)

			// Exit gracefully
			log.Println("manager: stopped!")
			exitChan <- 0
		}
	}
}
