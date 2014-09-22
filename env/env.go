package env

import (
	"os"
)

const (
	// envDebug is the name of the environment variable which enables wavepipe's
	// "debug" mode, used to enable debugging mode and disable authentication
	envDebug = "WAVEPIPE_DEBUG"

	// envTest is the name of the environment variable which enables wavepipe's
	// "test" mode, used for CI and other automated tests
	envTest = "WAVEPIPE_TEST"

	// disabled is the constant used to set a mode as disabled.
	disabled = "0"

	// enabled is the constant used to set a mode as enabled.
	enabled = "1"
)

// IsDebug checks if wavepipe debug mode is enabled.
func IsDebug() bool {
	return os.Getenv(envDebug) == enabled
}

// IsTest checks if wavepipe test mode is enabled.
func IsTest() bool {
	return os.Getenv(envTest) == enabled
}

// SetDebug enables or disables wavepipe debug mode.
func SetDebug(value bool) error {
	var enable string
	if value {
		enable = enabled
	} else {
		enable = disabled
	}

	return os.Setenv(envTest, enable)
}

// SetTest enables or disables wavepipe test mode.
func SetTest(value bool) error {
	var enable string
	if value {
		enable = enabled
	} else {
		enable = disabled
	}

	return os.Setenv(envTest, enable)
}
