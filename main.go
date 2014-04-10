package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// app is the name of the application, as printed in logs
const app = "wavepipe"

func main() {
	// Set up logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Application entry point
	log.Println(app, ": hello, world!")

	// Gracefully handle termination via UNIX signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)
	for sig := range sigChan {
		log.Println(app, ": caught signal:", sig)
		break
	}

	// Force terminate if signaled twice
	go func() {
		for sig := range sigChan {
			log.Println(app, ": caught signal:", sig, ", force halting now!")
		}
	}()

	// Graceful exit
	log.Println(app, ": graceful shutdown complete")
	os.Exit(0)
}
