package core

import (
	"log"
	"os"

	"github.com/mdlayher/wavepipe/common"
	"github.com/mdlayher/wavepipe/config"
	"github.com/mdlayher/wavepipe/env"
)

// App is the application's name
const App = "wavepipe"

// Version is the application's version
const Version = "git-master"

// Revision is the application's current git commit hash
var Revision string

// Manager is responsible for coordinating the application
func Manager(killChan chan struct{}, exitChan chan int) {
	// Check if a commit hash was injected
	if Revision == "" {
		log.Println("manager: empty git revision, please rebuild using 'make'")
	} else {
		log.Printf("manager: initializing %s %s [revision: %s]...", App, Version, Revision)
	}

	// Gather information about the operating system
	stat := common.OSInfo()
	log.Printf("manager: %s - %s_%s (%d CPU) [pid: %d]", stat.Hostname, stat.Platform, stat.Architecture, stat.NumCPU, stat.PID)

	// Set configuration source, load configuration
	config.C = new(config.CLIConfig)
	conf, err := config.C.Load()
	if err != nil {
		log.Fatalf("manager: could not load config: %s", err.Error())
	}

	// Check valid media folder, unless in test mode
	folder := conf.Media()
	if !env.IsTest() {
		// Check empty folder, provide help information if not set
		if folder == "" {
			log.Fatal("manager: no media folder set in config: ", config.C.Help())
		} else if _, err := os.Stat(folder); err != nil {
			// Check file existence
			log.Fatalf("manager: invalid media folder set in config: %s", err.Error())
		}
	}

	// Launch database manager to handle database/ORM connections
	dbLaunchChan := make(chan struct{})
	dbKillChan := make(chan struct{})
	go dbManager(*conf, dbLaunchChan, dbKillChan)

	// Wait for database to be fully ready before other operations start
	<-dbLaunchChan

	// Launch cron manager to handle timed events
	cronKillChan := make(chan struct{})
	go cronManager(cronKillChan)

	// Launch filesystem manager to handle file scanning
	fsKillChan := make(chan struct{})
	go fsManager(folder, fsKillChan)

	// Launch HTTP API server
	apiKillChan := make(chan struct{})
	go apiRouter(apiKillChan)

	// Launch transcode manager to handle ffmpeg and file transcoding
	transcodeKillChan := make(chan struct{})
	go transcodeManager(transcodeKillChan)

	// Wait for termination signal
	for {
		select {
		// Trigger a graceful shutdown
		case <-killChan:
			log.Println("manager: triggering graceful shutdown, press Ctrl+C again to force halt")

			// Stop transcodes, wait for confirmation
			transcodeKillChan <- struct{}{}
			<-transcodeKillChan
			close(transcodeKillChan)

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
