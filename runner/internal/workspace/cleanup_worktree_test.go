package workspace

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/runner/internal/testutil"
)

// --- Test CleanupOldWorktrees ---

func TestCleanupOldWorktreesEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir, "")

	err := manager.CleanupOldWorktrees(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCleanupOldWorktreesPreservesNonGitWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir, "")

	sandboxesDir := filepath.Join(tmpDir, "sandboxes")
	workspacePath := filepath.Join(sandboxesDir, "empty-pod", "workspace")
	credentialPath := filepath.Join(sandboxesDir, "empty-pod", gitCredentialFileName)
	sshKeyPath := filepath.Join(sandboxesDir, "empty-pod", ".ssh_key")
	os.MkdirAll(workspacePath, 0755)
	os.WriteFile(credentialPath, []byte("secret"), 0600)
	os.WriteFile(sshKeyPath, []byte("private"), 0600)

	err := manager.CleanupOldWorktrees(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(workspacePath); err != nil {
		t.Fatalf("non-Git workspace should be preserved: %v", err)
	}
	if _, err := os.Stat(credentialPath); err != nil {
		t.Fatalf("live sandbox credential should be preserved: %v", err)
	}
	if _, err := os.Stat(sshKeyPath); err != nil {
		t.Fatalf("live sandbox SSH key should be preserved: %v", err)
	}
}

func TestCleanupOldWorktreesRemovesAuthWhenWorkspaceIsMissing(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir, "")
	sandboxPath := filepath.Join(tmpDir, "sandboxes", "missing-workspace")
	credentialPath := filepath.Join(sandboxPath, gitCredentialFileName)
	sshKeyPath := filepath.Join(sandboxPath, ".ssh_key")
	os.MkdirAll(sandboxPath, 0755)
	os.WriteFile(credentialPath, []byte("secret"), 0600)
	os.WriteFile(sshKeyPath, []byte("private"), 0600)

	if err := manager.CleanupOldWorktrees(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, path := range []string{credentialPath, sshKeyPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("orphaned auth file should be removed: %s", path)
		}
	}
}

func TestCleanupOldWorktreesWithValidWorktree(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir, "")

	validWT := filepath.Join(tmpDir, "sandboxes", "valid", "workspace")
	os.MkdirAll(validWT, 0755)

	gitFile := filepath.Join(validWT, ".git")
	os.WriteFile(gitFile, []byte("gitdir: /some/path"), 0644)

	err := manager.CleanupOldWorktrees(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(validWT); os.IsNotExist(err) {
		t.Error("valid worktree should not be removed")
	}
}

func TestCleanupOldWorktreesWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir, "")

	sandboxesDir := filepath.Join(tmpDir, "sandboxes")
	os.MkdirAll(sandboxesDir, 0755)

	testFile := filepath.Join(sandboxesDir, "testfile")
	os.WriteFile(testFile, []byte("test"), 0644)

	err := manager.CleanupOldWorktrees(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("file should not be removed")
	}
}

func TestCleanupOldWorktreesPrunesMissingWorkspaceMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir, "")
	origin, clone := createPinnedOrigin(t)
	commit := commitPinnedFile(t, clone, "stale")
	pushPinnedBranch(t, clone)
	repoPath := filepath.Join(tmpDir, "repos", "repo.git")
	os.MkdirAll(filepath.Dir(repoPath), 0755)
	cmd := exec.Command("git", "clone", "--bare", origin, repoPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone --bare failed: %v: %s", err, output)
	}

	sandboxPath := filepath.Join(tmpDir, "sandboxes", "stale-pod")
	worktreePath := filepath.Join(sandboxPath, "workspace")
	runGitTestCommand(t, repoPath, "worktree", "add", "--detach", worktreePath, commit)
	credentialPath := filepath.Join(sandboxPath, gitCredentialFileName)
	sshKeyPath := filepath.Join(sandboxPath, ".ssh_key")
	os.WriteFile(credentialPath, []byte("secret"), 0600)
	os.WriteFile(sshKeyPath, []byte("private"), 0600)
	os.RemoveAll(worktreePath)

	err := manager.CleanupOldWorktrees(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(credentialPath); !os.IsNotExist(err) {
		t.Error("stale credential should be removed")
	}
	if _, err := os.Stat(sshKeyPath); !os.IsNotExist(err) {
		t.Error("stale SSH key should be removed")
	}
	cmd = exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git worktree list failed: %v: %s", err, output)
	}
	if strings.Contains(string(output), worktreePath) {
		t.Fatalf("stale worktree metadata remains: %s", output)
	}
}

func TestCleanupOldWorktreesReadDirError(t *testing.T) {
	testutil.SkipIfRoot(t)
	testutil.SkipIfNoChmodSupport(t)

	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir, "")

	sandboxesDir := filepath.Join(tmpDir, "sandboxes")
	os.MkdirAll(sandboxesDir, 0755)

	subDir := filepath.Join(sandboxesDir, "subdir")
	os.MkdirAll(subDir, 0755)
	os.Chmod(sandboxesDir, 0000)
	defer os.Chmod(sandboxesDir, 0755)

	err := manager.CleanupOldWorktrees(context.Background())
	if err == nil {
		t.Error("expected error for unreadable directory")
	}
}
