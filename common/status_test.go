package common

import (
	"os"
	"runtime"
	"testing"
)

// TestOSInfo verifies that correctness of the OSInfo() function
// NOTE: This also covers the Status() function, since all information generated there
// is dynamic except for what is provided by OSInfo()
func TestOSInfo(t *testing.T) {
	// Retrieve information about the operating system
	osStat := OSInfo()

	// Verify correctness of all fields

	// Architecture
	if osStat.Architecture != runtime.GOARCH {
		t.Fatalf("Mismatched architecture: %s != %s", osStat.Architecture, runtime.GOARCH)
	}

	// Get the current hostname
	hostname, err := os.Hostname()
	if err != nil {
		t.Fatalf("Failed to retrieve hostname: %s", err.Error())
	}

	// Hostname
	if osStat.Hostname != hostname {
		t.Fatalf("Mismatched hostname: %s != %s", osStat.Hostname, hostname)
	}

	// NumCPU
	if osStat.NumCPU != runtime.NumCPU() {
		t.Fatalf("Mismatched NumCPU: %d != %d", osStat.NumCPU, runtime.NumCPU())
	}

	// PID
	pid := os.Getpid()
	if osStat.PID != pid {
		t.Fatalf("Mismatched PID: %d != %d", osStat.PID, pid)
	}

	// Platform
	if osStat.Platform != runtime.GOOS {
		t.Fatalf("Mismatched platform: %s != %s", osStat.Platform, runtime.GOOS)
	}
}
