//go:build integration

package poddaemon

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/runner/internal/processmgr"
)

// TestMain wires the test process so it can both initialize processmgr (every
// integration test that drives CreateSession lands in startDaemon, which
// requires processmgr.Global()) and act as the launcher subprocess when
// startDaemon re-execs the test binary itself via os.Executable(). Without
// the LauncherSubcommand fork the first test that calls CreateSession would
// fail to spawn a daemon.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == processmgr.LauncherSubcommand {
		processmgr.RunLauncher()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	processmgr.Init(ctx, processmgr.Options{})
	os.Exit(m.Run())
}

// findModuleRoot walks up from the current directory to find go.mod.
func findModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find module root (go.mod)")
		}
		dir = parent
	}
}

// buildTestRunner returns a usable path to the runner binary for
// `go test -tags=integration`. Prefers RUNNER_BIN when set; otherwise
// builds ./runner/cmd/runner into a temp dir (proto stubs via
// scripts/proto-gen-go.sh when missing).
func buildTestRunner(t *testing.T) string {
	t.Helper()

	if env := os.Getenv("RUNNER_BIN"); env != "" {
		if _, err := os.Stat(env); err == nil {
			return env
		}
		t.Fatalf("RUNNER_BIN=%s not found", env)
	}

	modRoot := findModuleRoot(t)
	if _, err := os.Stat(filepath.Join(modRoot, "proto", "gen", "go", "runner", "v1", "runner.pb.go")); err != nil {
		cmd := exec.Command("bash", "scripts/proto-gen-go.sh", "--force")
		cmd.Dir = modRoot
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("cannot generate proto stubs for runner build: %v", err)
		}
	}

	bin := "runner"
	if runtime.GOOS == "windows" {
		bin = "runner.exe"
	}
	out := filepath.Join(t.TempDir(), bin)
	cmd := exec.Command("go", "build", "-o", out, "./runner/cmd/runner")
	cmd.Dir = modRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("cannot build runner binary: %v", err)
	}
	return out
}

// shortWorkspace creates a short temp dir for integration tests.
// Returns (workspace, sandbox) paths.
func shortWorkspace(t *testing.T, name string) (string, string) {
	t.Helper()

	workspace := t.TempDir()
	sandbox := filepath.Join(workspace, name)
	require.NoError(t, os.MkdirAll(sandbox, 0755))
	return workspace, sandbox
}
