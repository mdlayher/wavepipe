package config

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/user"
	"path"
)

// JSONFileConfig represents configuration from a JSON configuration file
type JSONFileConfig struct {
	path string
}

// configCache is a cached configuration
var configCache *Config

// Load returns the configuration from a JSON configuration file
func (c *JSONFileConfig) Load() (*Config, error) {
	// Check for cached config
	if configCache != nil {
		return configCache, nil
	}

	// Attempt to load the configuration file from its path
	configFile, err := os.Open(c.path)
	if err != nil {
		return nil, err
	}

	// Decode JSON file
	var config *Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return nil, err
	}

	// Cache and return config
	configCache = config
	return config, nil
}

// Use sets the configuration file location for a JSON configuration file, attempting
// to create it if it does not exist
func (c *JSONFileConfig) Use(configPath string) error {
	// Check for configuration at this path
	if _, err := os.Stat(configPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		// Get current user
		user, err := user.Current()
		if err != nil {
			return err
		}

		// Only create file if it's in the default location
		if configPath != user.HomeDir+"/.config/wavepipe/wavepipe.json" {
			return errors.New("config: cannot create config file: " + configPath)
		}

		log.Println("config: creating new config file:", configPath)

		// Create a new config file in the default location
		dir := path.Dir(configPath) + "/"
		file := path.Base(configPath)

		// Make directory
		if err := os.MkdirAll(dir, 0775); err != nil {
			return err
		}

		// Attempt to open destination
		dest, err := os.Create(dir + file)
		if err != nil {
			return err
		}

		// Choose a config to copy, either regular or Travis build
		configBuf := DefaultConfig
		if os.Getenv("WAVEPIPE_TEST") == "1" {
			configBuf = TravisConfig
		}

		// Copy contents into destination
		if _, err := dest.Write(configBuf); err != nil {
			return err
		}

		// Close file
		if err := dest.Close(); err != nil {
			return err
		}
	}

	// Store path
	c.path = configPath
	return nil
}
