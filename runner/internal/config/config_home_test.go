package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserConfigDirPrefersAgentCloud(t *testing.T) {
	home := t.TempDir()

	for _, dir := range []string{userConfigDirName, ".do-worker", ".agentsmesh"} {
		if err := os.Mkdir(filepath.Join(home, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}

	want := filepath.Join(home, userConfigDirName)
	if got := userConfigDirForHome(home); got != want {
		t.Fatalf("userConfigDirForHome() = %q, want %q", got, want)
	}
}

func TestUserConfigDirUsesDoWorkerLegacyWhenPreferredMissing(t *testing.T) {
	home := t.TempDir()

	if err := os.Mkdir(filepath.Join(home, ".do-worker"), 0755); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(home, ".do-worker")
	if got := userConfigDirForHome(home); got != want {
		t.Fatalf("userConfigDirForHome() = %q, want %q", got, want)
	}
}

func TestUserConfigDirUsesAgentsmeshLegacyWhenOthersMissing(t *testing.T) {
	home := t.TempDir()

	if err := os.Mkdir(filepath.Join(home, ".agentsmesh"), 0755); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(home, ".agentsmesh")
	if got := userConfigDirForHome(home); got != want {
		t.Fatalf("userConfigDirForHome() = %q, want %q", got, want)
	}
}
