package core

import (
	"log"
	"os"
	"time"

	"github.com/mdlayher/wavepipe/config"
)

// App is the application's name
const App = "wavepipe"

// Version is the application's version
const Version = "git-master"

// ConfigPath is the application's configuration path
var ConfigPath string

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

	// Set configuration (if default path used, config will be created)
	config.C = new(config.JSONFileConfig)
	if err := config.C.Use(ConfigPath); err != nil {
		log.Fatalf("manager: could not use config: %s, %s", ConfigPath, err.Error())
	}

	// Load configuration from specified source
	conf, err := config.C.Load()
	if err != nil {
		log.Fatalf("manager: could not load config: %s, %s", ConfigPath, err.Error())
	}

	// Check valid media folder, unless in test mode
	folder := conf.Media()
	if os.Getenv("WAVEPIPE_TEST") != "1" {
		// Check empty folder
		if folder == "" {
			log.Fatalf("manager: no media folder set in config: %s", ConfigPath)
		} else if _, err := os.Stat(folder); err != nil {
			// Check file existence
			log.Fatalf("manager: invalid media folder set in config: %s", err.Error())
		}
	}

	// Launch database manager to handle database/ORM connections
	dbKillChan := make(chan struct{})
	go dbManager(*conf, dbKillChan)

	// Launch cron manager to handle timed events
	cronKillChan := make(chan struct{})
	go cronManager(cronKillChan)

	// Launch filesystem manager to handle file scanning
	fsKillChan := make(chan struct{})
	go fsManager(folder, fsKillChan)

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
