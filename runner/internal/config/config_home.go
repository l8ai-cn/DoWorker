package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const userConfigDirName = ".agent-cloud"

// Keep literal legacy brand directory names for installed runners.
var legacyUserConfigDirNames = []string{
	".do-worker",
	".agentsmesh",
}

// UserConfigDir returns the runner config directory under the user's home.
// Prefers ~/.agent-cloud; falls back to known legacy dirs when only those exist.
func UserConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return userConfigDirForHome(home)
}

func userConfigDirForHome(home string) string {
	preferred := filepath.Join(home, userConfigDirName)
	if info, err := os.Stat(preferred); err == nil && info.IsDir() {
		return preferred
	}
	for _, name := range legacyUserConfigDirNames {
		dir := filepath.Join(home, name)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return preferred
}

// PreferredUserConfigDir always returns ~/.agent-cloud for new writes.
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
	paths := []string{filepath.Join(home, userConfigDirName)}
	for _, name := range legacyUserConfigDirNames {
		paths = append(paths, filepath.Join(home, name))
	}
	return paths
}

func systemConfigSearchPaths() []string {
	if runtime.GOOS == "windows" {
		return nil
	}
	return []string{
		"/etc/agent-cloud",
		"/etc/do-worker",
		"/etc/agentsmesh",
	}
}
