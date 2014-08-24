package config

import (
	"path"

	"github.com/mdlayher/wavepipe/common"
)

// C is the active configuration instance
var C ConfigSource

// Config represents the program configuration options
type Config struct {
	Host        string        `json:"host"`
	MediaFolder string        `json:"mediaFolder"`
	Sqlite      *SqliteConfig `json:"sqlite"`
}

// Media returns the media folder from config, but with special
// characters such as '~' replaced, and any trailing slashes trimmed.
func (c Config) Media() string {
	// Return path with strings replaced, trailing slash removed,
	// tilde replaced with current user's home directory
	return path.Clean(common.ExpandHomeDir(c.MediaFolder))
}

// SqliteConfig represents configuration for an sqlite backend
type SqliteConfig struct {
	File string `json:"file"`
}

// ConfigSource represents the configuration source for the program
type ConfigSource interface {
	Help() string
	Load() (*Config, error)
}
