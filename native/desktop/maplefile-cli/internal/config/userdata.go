// internal/config/userdata.go
package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetUserDataDir returns the appropriate directory for storing application data
// following platform-specific conventions:
// - Windows: %LOCALAPPDATA%\{appName}
// - macOS: ~/Library/Application Support/{appName}
// - Linux: ~/.local/share/{appName} (or $XDG_DATA_HOME/{appName})
func GetUserDataDir(appName string) (string, error) {
	var baseDir string
	var err error

	switch runtime.GOOS {
	case "windows":
		// Use LOCALAPPDATA for application data on Windows
		baseDir = os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			// Fallback to APPDATA if LOCALAPPDATA is not set
			baseDir = os.Getenv("APPDATA")
			if baseDir == "" {
				// Last resort: use UserConfigDir
				baseDir, err = os.UserConfigDir()
				if err != nil {
					return "", err
				}
			}
		}

	case "darwin":
		// Use ~/Library/Application Support on macOS
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		baseDir = filepath.Join(home, "Library", "Application Support")

	default:
		// Linux and other Unix-like systems
		// Follow XDG Base Directory Specification
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			baseDir = xdgData
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			baseDir = filepath.Join(home, ".local", "share")
		}
	}

	// Combine with app name
	appDataDir := filepath.Join(baseDir, appName)

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		return "", err
	}

	return appDataDir, nil
}
