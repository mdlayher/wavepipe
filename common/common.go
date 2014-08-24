package common

import (
	"os"
	"os/user"
)

// System is a grouped global which stores static information about the
// host operating system.
var System struct {
	Hostname string
	User     *user.User
}

// init fetches information from the operating system and stores it in System
// for common access from different components of the service.
func init() {
	// Fetch system hostname
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	System.Hostname = hostname

	// Fetch current user
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	System.User = user
}
