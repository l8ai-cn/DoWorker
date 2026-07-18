package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListHostWorkspaceEntriesAcceptsAbsolutePathInsideRoot(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "project")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(child, "design.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, err := listHostWorkspaceEntries(root, child)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].GetName() != "design.txt" {
		t.Fatalf("unexpected entries: %#v", entries)
	}
}

func TestListHostWorkspaceEntriesRejectsAbsolutePathOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	if _, err := listHostWorkspaceEntries(root, outside); err == nil {
		t.Fatal("expected path outside workspace root to be rejected")
	}
}

func TestListHostWorkspaceEntriesRejectsSymlinkOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "outside")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}

	if _, err := listHostWorkspaceEntries(root, link); err == nil {
		t.Fatal("expected symlink outside workspace root to be rejected")
	}
}
