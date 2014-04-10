package core

import (
	"log"
	"time"
)

// App is the application's name
const App = "wavepipe"

// Version is the application's version
const Version = "git-master"

// MediaFolder is the folder which we will recursively scan for media
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

	// Launch cron manager to handle timed events
	cronKillChan := make(chan struct{})
	go cronManager(cronKillChan)

	// Launch filesystem manager to handle file scanning
	// TODO: make this an actual path later on via configuration
	fsKillChan := make(chan struct{})
	go fsManager(MediaFolder, fsKillChan)

	// Launch HTTP server
	log.Println("manager: starting HTTP server")

	// Wait for termination signal
	for {
		select {
		// Trigger a graceful shutdown
		case <-killChan:
			log.Println("manager: triggering graceful shutdown, press Ctrl+C again to force halt")

			// Stop cron, wait for confirmation
			cronKillChan <- struct{}{}
			<-cronKillChan
			close(cronKillChan)

			// Stop filesystem, wait for confirmation
			fsKillChan <- struct{}{}
			<-fsKillChan
			close(fsKillChan)

			// Exit gracefully
			log.Println("manager: stopped!")
			exitChan <- 0
		}
	}
}
