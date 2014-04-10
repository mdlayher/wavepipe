package core

import (
	"log"
	"time"
)

// cronManager spawns and triggers events at regular intervals
func cronManager(cronKillChan chan struct{}) {
	log.Println("cron: starting...")

	// cronPrintCurrentStatus - run on startup, and every 5 minutes
	status := time.NewTicker(5 * time.Minute)
	go cronPrintCurrentStatus()

	// Trigger events via ticker
	for {
		select {
		// Stop cron
		case <-cronKillChan:
			// Inform manager that shutdown is complete
			log.Println("cron: stopped!")
			cronKillChan <- struct{}{}
		case <-status.C:
			go cronPrintCurrentStatus()
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

