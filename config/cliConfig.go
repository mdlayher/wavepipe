package config

import (
	"flag"
)

var (
	// hostFlag is a flag which defines the host which wavepipe will bind to
	hostFlag = flag.String("host", ":8080", "The host which wavepipe will bind to.")
	// mediaFlag is a flag which defines the media folder wavepipe will scan
	mediaFlag = flag.String("media", "", "The media folder which wavepipe will scan and watch.")
	// sqliteFlag is a flag which defines the location of the wavepipe sqlite database
	sqliteFlag = flag.String("sqlite", "~/.config/wavepipe/wavepipe.db", "The sqlite database which wavepipe will use.")
)

// CLIConfig represents configuration from command-line flags
type CLIConfig struct{}

// Help returns a string containing help information about command-line flags
func (CLIConfig) Help() string {
	return "use the '-media' flag to specify a folder"
}

// Load returns the configuration from command-line flags
func (c *CLIConfig) Load() (*Config, error) {
	flag.Parse()

	return &Config{
		Host:        *hostFlag,
		MediaFolder: *mediaFlag,
		Sqlite: &SqliteConfig{
			File: *sqliteFlag,
		},
	}, nil
}
