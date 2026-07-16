package poddaemon

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// PodDaemonManager manages the lifecycle of pod daemon sessions.
type PodDaemonManager struct {
	sandboxesDir  string // Base directory containing per-pod sandbox directories
	runnerBinPath string
}

// CreateOpts holds options for creating a new daemon session.
type CreateOpts struct {
	PodKey  string
	Agent   string
	Command string
	Args    []string
	WorkDir string
	Env     []string
	Cols    int
	Rows    int

	SandboxPath    string
	WorkspaceID    *WorkspaceIdentity
	RepositoryURL  string
	Branch         string
	TicketSlug     string
	VTHistoryLimit int
	Perpetual      bool
}

// authTokenBytes is the number of random bytes for IPC authentication tokens.
const authTokenBytes = 32

// NewPodDaemonManager creates a new manager.
// sandboxesDir is the base directory containing per-pod sandbox directories (each with pod_daemon.json).
func NewPodDaemonManager(sandboxesDir string) (*PodDaemonManager, error) {
	binPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("get executable path: %w", err)
	}

	return &PodDaemonManager{
		sandboxesDir:  sandboxesDir,
		runnerBinPath: binPath,
	}, nil
}

// generateAuthToken creates a cryptographically random hex-encoded token.
func generateAuthToken() (string, error) {
	b := make([]byte, authTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate auth token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// CreateSession spawns a new daemon process and returns a connected daemonPTY.
func (m *PodDaemonManager) CreateSession(opts CreateOpts) (*daemonPTY, *PodDaemonState, error) {
	log := slog.Default()

	if opts.SandboxPath == "" {
		return nil, nil, fmt.Errorf("sandbox path is required")
	}
	workspaceID, err := workspaceIdentityForSession(
		opts.WorkDir,
		opts.WorkspaceID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("capture workspace identity: %w", err)
	}

	authToken, err := generateAuthToken()
	if err != nil {
		return nil, nil, err
	}

	state := &PodDaemonState{
		PodKey:         opts.PodKey,
		Agent:          opts.Agent,
		AuthToken:      authToken,
		SandboxPath:    opts.SandboxPath,
		WorkDir:        opts.WorkDir,
		WorkspaceID:    workspaceID,
		RepositoryURL:  opts.RepositoryURL,
		Branch:         opts.Branch,
		TicketSlug:     opts.TicketSlug,
		Command:        opts.Command,
		Args:           opts.Args,
		Env:            opts.Env,
		Cols:           opts.Cols,
		Rows:           opts.Rows,
		StartedAt:      time.Now(),
		VTHistoryLimit: opts.VTHistoryLimit,
		Perpetual:      opts.Perpetual,
	}

	// Save state before starting daemon (daemon reads it on startup).
	// IPCAddr is empty — the daemon will fill it after binding a port.
	if err := SaveState(state); err != nil {
		return nil, nil, fmt.Errorf("save state: %w", err)
	}

	configPath := StatePath(opts.SandboxPath)
	pid, err := startDaemon(m.runnerBinPath, configPath, opts.SandboxPath, opts.Env)
	if err != nil {
		_ = DeleteState(opts.SandboxPath)
		return nil, nil, fmt.Errorf("start daemon: %w", err)
	}

	log.Info("daemon started, waiting for IPC", "pid", pid)

	// Wait for daemon to bind a port and write it to state file
	dpty, updatedState, err := m.waitForDaemon(opts.SandboxPath, authToken, pid)
	if err != nil {
		status := daemonProcessStatus(pid)
		log.Error("daemon failed to become ready",
			"pod_key", opts.PodKey, "pid", pid, "process_status", status, "error", err)
		captureDaemonLog(log, opts.SandboxPath, opts.PodKey)
		_ = DeleteState(opts.SandboxPath)
		return nil, nil, fmt.Errorf("connect to daemon (pid %d, %s): %w", pid, status, err)
	}

	return dpty, updatedState, nil
}
