package core

import (
	"log"
	"time"

	"github.com/mdlayher/wavepipe/common"
	"github.com/mdlayher/wavepipe/config"
)

// cronManager spawns and triggers events at regular intervals
func cronManager(cronKillChan chan struct{}) {
	log.Println("cron: starting...")

	// Retrieve configuration for use with crons
	conf, err := config.C.Load()
	if err != nil {
		log.Fatal("cron: could not load configuration")
	}

	// cronPrintCurrentStatus - run on startup, and every 5 minutes
	status := time.NewTicker(5 * time.Minute)
	go cronPrintCurrentStatus()

	// TODO: because of some erratic filesystem watcher behavior, we run full filesystem
	// TODO: media and orphan scans via cron at regular intervals.  The evented scans should work
	// TODO: in the vast majority of cases, but these will help ensure consistency until I have
	// TODO: done further research regarding the watchers

	// cronMediaScan - scan every 30 minutes
	mediaScan := time.NewTicker(30 * time.Minute)

	// cronOrphanScan - scan every 30 minutes
	orphanScan := time.NewTicker(30 * time.Minute)

	// Trigger events via ticker
	for {
		select {
		// Stop cron
		case <-cronKillChan:
			// Inform manager that shutdown is complete
			log.Println("cron: stopped!")
			cronKillChan <- struct{}{}
			return
		// Trigger status printing
		case <-status.C:
			go cronPrintCurrentStatus()
		// Trigger media scan
		case <-mediaScan.C:
			// Queue a new media scan
			m := new(fsMediaScan)
			m.SetFolders(conf.Media(), "")
			m.Verbose(true)
			fsQueue <- m
		// Trigger orphan scan
		case <-orphanScan.C:
			// Queue a new orphan scan
			o := new(fsOrphanScan)
			o.SetFolders(conf.Media(), "")
			o.Verbose(true)
			fsQueue <- o
		}
	}
}

// cronPrintCurrentStatus logs the regular status check banner
func cronPrintCurrentStatus() {
	// Regular status banner
	stat := common.ServerStatus()
	log.Printf("cron: status - [uptime: %d] [goroutines: %d] [memory: %02.3f MB]", stat.Uptime, stat.NumGoroutine, stat.MemoryMB)
}
