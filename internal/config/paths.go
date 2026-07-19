package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func DefaultConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "cogniq"), nil
}

func DefaultDataDir() (string, error) {
	switch runtime.GOOS {
	case "linux":
		base := os.Getenv("XDG_DATA_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(base, "cogniq"), nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "cogniq"), nil
	case "windows":
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(base, "CogniQ"), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share", "cogniq"), nil
	}
}

func DefaultSocketPath() (string, error) {
	dataDir, err := DefaultDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "cogniq.sock"), nil
}
