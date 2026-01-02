// Package config provides application configuration and path management.
// It follows XDG Base Directory Specification for Linux.
package config

import (
	"os"
	"path/filepath"
)

const (
	// AppName is the application identifier used in paths
	AppName = "guanaco"

	// DatabaseName is the SQLite database filename
	DatabaseName = "guanaco.db"
)

// GetDataDir returns the path to the application data directory.
// Respects XDG_DATA_HOME, defaults to ~/.local/share/guanaco
func GetDataDir() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, _ := os.UserHomeDir()
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, AppName)
}

// GetConfigDir returns the path to the application config directory.
// Respects XDG_CONFIG_HOME, defaults to ~/.config/guanaco
func GetConfigDir() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, _ := os.UserHomeDir()
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, AppName)
}

// GetDatabasePath returns the full path to the SQLite database file.
func GetDatabasePath() string {
	return filepath.Join(GetDataDir(), DatabaseName)
}

// EnsureDirectories creates the necessary application directories if they don't exist.
func EnsureDirectories() error {
	dirs := []string{
		GetDataDir(),
		GetConfigDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
