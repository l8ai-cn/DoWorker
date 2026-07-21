package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const userConfigDirName = ".agent-cloud"

// UserConfigDir returns the runner config directory under the user's home.
func UserConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, userConfigDirName)
}

func userConfigDirForHome(home string) string {
	return filepath.Join(home, userConfigDirName)
}

// PreferredUserConfigDir always returns ~/.agent-cloud for new writes.
func PreferredUserConfigDir() string {
	return UserConfigDir()
}

func userConfigSearchPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{filepath.Join(home, userConfigDirName)}
}

func systemConfigSearchPaths() []string {
	if runtime.GOOS == "windows" {
		return nil
	}
	return []string{"/etc/agent-cloud"}
}
