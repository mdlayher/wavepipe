package core

import (
	"os"
	"runtime"
)

// osInfo represents information about the host operating system for this process
type osInfo struct {
	Architecture string
	Hostname     string
	NumCPU       int
	PID          int
	Platform     string
}

// OSInfo returns information about the host operating system for this process
func OSInfo() (*osInfo, error) {
	// Get system hostname
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	// Return osInfo
	return &osInfo{
		Architecture: runtime.GOARCH,
		Hostname:     hostname,
		NumCPU:       runtime.NumCPU(),
		PID:          os.Getpid(),
		Platform:     runtime.GOOS,
	}, nil
}
