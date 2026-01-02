package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetDataDir(t *testing.T) {
	dir := GetDataDir()

	if dir == "" {
		t.Error("GetDataDir() returned empty string")
	}

	if !strings.Contains(dir, "guanaco") {
		t.Errorf("GetDataDir() = %q, want path containing 'guanaco'", dir)
	}
}

func TestGetConfigDir(t *testing.T) {
	dir := GetConfigDir()

	if dir == "" {
		t.Error("GetConfigDir() returned empty string")
	}

	if !strings.Contains(dir, "guanaco") {
		t.Errorf("GetConfigDir() = %q, want path containing 'guanaco'", dir)
	}
}

func TestGetDatabasePath(t *testing.T) {
	dbPath := GetDatabasePath()

	if dbPath == "" {
		t.Error("GetDatabasePath() returned empty string")
	}

	if !strings.HasSuffix(dbPath, "guanaco.db") {
		t.Errorf("GetDatabasePath() = %q, want path ending with 'guanaco.db'", dbPath)
	}
}

func TestGetDataDir_RespectsXDGDataHome(t *testing.T) {
	// Save original and restore after test
	original := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", original)

	testDir := "/tmp/test-xdg-data"
	os.Setenv("XDG_DATA_HOME", testDir)

	dir := GetDataDir()
	expected := filepath.Join(testDir, "guanaco")

	if dir != expected {
		t.Errorf("GetDataDir() = %q, want %q", dir, expected)
	}
}

func TestGetConfigDir_RespectsXDGConfigHome(t *testing.T) {
	// Save original and restore after test
	original := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", original)

	testDir := "/tmp/test-xdg-config"
	os.Setenv("XDG_CONFIG_HOME", testDir)

	dir := GetConfigDir()
	expected := filepath.Join(testDir, "guanaco")

	if dir != expected {
		t.Errorf("GetConfigDir() = %q, want %q", dir, expected)
	}
}

func TestEnsureDirectories(t *testing.T) {
	// Save originals and restore after test
	origData := os.Getenv("XDG_DATA_HOME")
	origConfig := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("XDG_DATA_HOME", origData)
		os.Setenv("XDG_CONFIG_HOME", origConfig)
	}()

	// Use temp directories for test
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))

	err := EnsureDirectories()
	if err != nil {
		t.Fatalf("EnsureDirectories() error = %v", err)
	}

	// Verify directories exist
	dataDir := GetDataDir()
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Errorf("Data directory was not created: %s", dataDir)
	}

	configDir := GetConfigDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created: %s", configDir)
	}
}
