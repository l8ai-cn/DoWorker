package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserConfigDirPrefersDoWorker(t *testing.T) {
	home := t.TempDir()

	for _, dir := range []string{userConfigDirName, legacyConfigDirName} {
		if err := os.Mkdir(filepath.Join(home, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}

	want := filepath.Join(home, userConfigDirName)
	if got := userConfigDirForHome(home); got != want {
		t.Fatalf("userConfigDirForHome() = %q, want %q", got, want)
	}
}

func TestUserConfigDirUsesLegacyDirectoryWhenPreferredMissing(t *testing.T) {
	home := t.TempDir()

	if err := os.Mkdir(filepath.Join(home, legacyConfigDirName), 0755); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(home, legacyConfigDirName)
	if got := userConfigDirForHome(home); got != want {
		t.Fatalf("userConfigDirForHome() = %q, want %q", got, want)
	}
}
