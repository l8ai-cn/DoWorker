package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	userConfigDirName   = ".do-worker"
	legacyConfigDirName = ".agentsmesh"
)

// UserConfigDir returns the runner config directory under the user's home.
// Prefers ~/.do-worker; falls back to ~/.agentsmesh when only the legacy dir exists.
func UserConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	newDir := filepath.Join(home, userConfigDirName)
	legacyDir := filepath.Join(home, legacyConfigDirName)
	if info, err := os.Stat(newDir); err == nil && info.IsDir() {
		return newDir
	}
	if info, err := os.Stat(legacyDir); err == nil && info.IsDir() {
		return legacyDir
	}
	return newDir
}

// PreferredUserConfigDir always returns ~/.do-worker for new writes.
func PreferredUserConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, userConfigDirName)
}

func userConfigSearchPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(home, userConfigDirName),
		filepath.Join(home, legacyConfigDirName),
	}
}

func systemConfigSearchPaths() []string {
	if runtime.GOOS == "windows" {
		return nil
	}
	return []string{
		"/etc/do-worker",
		"/etc/agentsmesh",
	}
}
