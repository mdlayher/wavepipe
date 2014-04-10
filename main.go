package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mdlayher/wavepipe/core"
)

func main() {
	// Set up logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Application entry point
	log.Println(core.App, ": starting...")

	// Invoke the manager, with graceful termination and core.Application exit code channels
	killChan := make(chan struct{})
	exitChan := make(chan int)
	go core.Manager(killChan, exitChan)

	// Gracefully handle termination via UNIX signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)
	for sig := range sigChan {
		log.Println(core.App, ": caught signal:", sig)
		killChan <- struct{}{}
		break
	}

	// Force terminate if signaled twice
	go func() {
		for sig := range sigChan {
			log.Println(core.App, ": caught signal:", sig, ", force halting now!")
			os.Exit(1)
		}
	}()

	// Graceful exit
	code := <-exitChan
	log.Println(core.App, ": graceful shutdown complete")
	os.Exit(code)
}
