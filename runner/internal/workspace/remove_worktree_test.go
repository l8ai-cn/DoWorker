package workspace

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// --- Test RemoveWorktree ---

func TestRemoveWorktreeNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir, "")

	err := manager.RemoveWorktree(context.Background(), "/nonexistent/worktree")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveWorktreeWithoutGitMetadataRemovesCredential(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir, "")
	sandboxPath := filepath.Join(tmpDir, "sandboxes", "test")
	worktreePath := filepath.Join(sandboxPath, "workspace")
	credentialPath := filepath.Join(sandboxPath, gitCredentialFileName)
	os.MkdirAll(worktreePath, 0755)
	os.WriteFile(credentialPath, []byte("secret"), 0600)

	err := manager.RemoveWorktree(context.Background(), worktreePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("worktree should be removed")
	}
	if _, err := os.Stat(credentialPath); !os.IsNotExist(err) {
		t.Error("credential should be removed")
	}
}

func TestRemoveWorktreeWithGitFile(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir, "")

	worktreePath := filepath.Join(tmpDir, "worktrees", "test-wt")
	os.MkdirAll(worktreePath, 0755)

	gitFile := filepath.Join(worktreePath, ".git")
	os.WriteFile(gitFile, []byte("gitdir: /nonexistent/repo/.git/worktrees/test-wt"), 0644)
	credentialPath := filepath.Join(filepath.Dir(worktreePath), gitCredentialFileName)
	sshKeyPath := filepath.Join(filepath.Dir(worktreePath), ".ssh_key")
	os.WriteFile(credentialPath, []byte("secret"), 0600)
	os.WriteFile(sshKeyPath, []byte("private"), 0600)

	err := manager.RemoveWorktree(context.Background(), worktreePath)
	if err == nil {
		t.Fatal("expected prune error")
	}

	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("worktree should be removed")
	}
	for _, path := range []string{credentialPath, sshKeyPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("auth file should be removed after prune failure: %s", path)
		}
	}
}

func TestRemoveWorktreeInternalFallback(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir, "")

	worktreePath := filepath.Join(tmpDir, "worktree")
	repoPath := filepath.Join(tmpDir, "repo")
	os.MkdirAll(worktreePath, 0755)
	os.MkdirAll(repoPath, 0755)

	testFile := filepath.Join(worktreePath, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	err := manager.removeWorktreeInternal(context.Background(), repoPath, worktreePath)
	if err == nil {
		t.Fatal("expected prune error")
	}

	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("worktree should be removed")
	}
}

func TestRemoveWorktreeInternalWithPrune(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()

	repoPath := filepath.Join(tmpDir, "repo")
	os.MkdirAll(repoPath, 0755)

	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	cmd.Run()

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = repoPath
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = repoPath
	cmd.Run()

	testFile := filepath.Join(repoPath, "file.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = repoPath
	cmd.Run()

	worktreePath := filepath.Join(tmpDir, "worktree")
	cmd = exec.Command("git", "worktree", "add", worktreePath, "HEAD")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create git worktree: %v", err)
	}

	manager, _ := NewManager(tmpDir, "")

	ctx := context.Background()
	err := manager.removeWorktreeInternal(ctx, repoPath, worktreePath)
	if err != nil {
		t.Fatalf("removeWorktreeInternal failed: %v", err)
	}

	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("worktree should be removed")
	}
}

func TestRemoveWorktreeInternalIgnoresCallerCancellationDuringCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "repo")
	os.MkdirAll(repoPath, 0755)
	runGitTestCommand(t, repoPath, "init")
	runGitTestCommand(t, repoPath, "config", "user.email", "test@test.com")
	runGitTestCommand(t, repoPath, "config", "user.name", "Test")
	os.WriteFile(filepath.Join(repoPath, "file.txt"), []byte("content"), 0644)
	runGitTestCommand(t, repoPath, "add", ".")
	runGitTestCommand(t, repoPath, "commit", "-m", "init")

	worktreePath := filepath.Join(tmpDir, "worktree")
	runGitTestCommand(t, repoPath, "worktree", "add", worktreePath, "HEAD")
	manager, _ := NewManager(tmpDir, "")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := manager.removeWorktreeInternal(ctx, repoPath, worktreePath); err != nil {
		t.Fatalf("removeWorktreeInternal failed: %v", err)
	}
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("worktree should be removed")
	}
}

func runGitTestCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v: %s", args, err, output)
	}
}
