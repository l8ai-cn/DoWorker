package config

import (
	"path/filepath"
	"testing"
)

func TestUserConfigDirUsesAgentCloud(t *testing.T) {
	home := t.TempDir()
	want := filepath.Join(home, userConfigDirName)
	if got := userConfigDirForHome(home); got != want {
		t.Fatalf("userConfigDirForHome() = %q, want %q", got, want)
	}
}
