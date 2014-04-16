package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"github.com/mdlayher/wavepipe/core"
)

// testFlag invokes wavepipe in "test" mode, where it will start and exit shortly after.  Used for testing.
var testFlag = flag.Bool("test", false, "Starts "+core.App+" in test mode, causing it to exit shortly after starting.")

func main() {
	// Set up logging, parse flags
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()

	// Check if wavepipe was invoked as root (which is a really bad idea)
	currUser, err := user.Current()
	if err != nil {
		log.Fatal(core.App, ": could not determine current user, exiting")
	}

	// Check for root, notify user if so
	if currUser.Uid == "0" || currUser.Gid == "0" || currUser.Username == "root" {
		log.Println(core.App, ": WARNING, it is NOT advisable to run wavepipe as root!")
	}

	// Set configuration path
	core.ConfigPath = currUser.HomeDir + "/.config/wavepipe/wavepipe.json"

	// Application entry point
	log.Println(core.App, ": starting...")

	// Gracefully handle termination via UNIX signal
	sigChan := make(chan os.Signal, 1)

	// In test mode, wait for a short time, then invoke a signal shutdown
	if *testFlag {
		// Set an environment variable to enable mocking in other areas of the program
		if err := os.Setenv("WAVEPIPE_TEST", "1"); err != nil {
			log.Println(err)
		}

		go func() {
			// Wait 30 seconds, to allow reasonable startup time
			seconds := 30
			log.Println(core.App, ": started in test mode, stopping in", seconds, "seconds.")
			<-time.After(time.Duration(seconds) * time.Second)

			// Send interrupt
			sigChan <- os.Interrupt
		}()
	}

	// Invoke the manager, with graceful termination and core.Application exit code channels
	killChan := make(chan struct{})
	exitChan := make(chan int)
	go core.Manager(killChan, exitChan)

	// Trigger a shutdown if SIGINT or SIGTERM received
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
