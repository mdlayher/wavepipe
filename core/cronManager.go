package core

import (
	"log"
	"time"

	"github.com/mdlayher/wavepipe/api/auth"
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

	// cronOrphanScan - scan every 30 minutes
	orphanScan := time.NewTicker(30 * time.Minute)

	// cronClearNonceFilter - clear the API nonce bloom filter every day
	clearNonceFilter := time.NewTicker(24 * time.Hour)

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
		// Trigger orphan scan
		case <-orphanScan.C:
			// Queue a new orphan scan
			o := new(fsOrphanScan)
			o.SetFolders(conf.Media(), "")
			o.Verbose(true)
			fsQueue <- o
		// Trigger a nonce filter reset
		case <-clearNonceFilter.C:
			log.Println("cron: clearing API nonce bloom filter")
			auth.NonceFilter.ClearAll()
		}
	}
}

// cronPrintCurrentStatus logs the regular status check banner
func cronPrintCurrentStatus() {
	// Get server status
	stat, err := Status()
	if err != nil {
		log.Printf("cron: could not get current status: %s", err.Error())
		return
	}

	// Regular status banner
	log.Printf("cron: status - [uptime: %d] [goroutines: %d] [memory: %02.3f MB]", stat.Uptime, stat.NumGoroutine, stat.MemoryMB)
}
