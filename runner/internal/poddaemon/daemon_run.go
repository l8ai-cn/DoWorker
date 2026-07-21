package poddaemon

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/l8ai-cn/agentcloud/runner/internal/envfilter"
	"github.com/l8ai-cn/agentcloud/runner/internal/safego"
)

func RunDaemon(configPath string) {
	log := slog.Default()
	log.Info("pod daemon starting", "config", configPath)
	if message := os.Getenv("_AGENTCLOUD_DAEMON_TEST_PANIC"); message != "" {
		panic(message)
	}

	state, err := LoadState(filepath.Dir(configPath))
	if err != nil {
		log.Error("failed to load state", "error", err)
		os.Exit(1)
	}
	env := state.Env
	if len(env) == 0 {
		env = os.Environ()
	}
	workspace, err := OpenWorkspaceLaunchGuard(
		state.WorkDir,
		state.WorkspaceID,
	)
	if err != nil {
		log.Error("workspace identity rejected", "error", err)
		os.Exit(1)
	}
	proc, err := startDaemonProcessInWorkspace(
		state.Command,
		state.Args,
		state.WorkDir,
		workspace,
		envfilter.FilterEnv(env),
		state.Cols,
		state.Rows,
	)
	workspace.Close()
	if err != nil {
		log.Error("failed to start process", "error", err)
		os.Exit(1)
	}
	defer proc.Close()

	listener, err := Listen()
	if err != nil {
		log.Error("failed to listen on IPC", "error", err)
		os.Exit(1)
	}
	defer listener.Close()
	state.IPCAddr = listener.Addr().String()
	state.DaemonPID = os.Getpid()
	if err := SaveState(state); err != nil {
		log.Error("failed to save state with IPC addr", "error", err)
		listener.Close()
		os.Exit(1)
	}

	daemon := &daemonServer{
		proc:     proc,
		listener: listener,
		exitDone: make(chan struct{}),
		orphanCh: make(chan struct{}),
		log:      log,
		state:    state,
	}
	if value := os.Getenv("_AGENTCLOUD_ORPHAN_CHECK_INTERVAL_SEC"); value != "" {
		if seconds, parseErr := strconv.Atoi(value); parseErr == nil && seconds > 0 {
			daemon.orphanCheckInterval = time.Duration(seconds) * time.Second
		}
	}
	safego.Go("daemon-proc-wait", func() {
		code, waitErr := proc.Wait()
		if waitErr != nil {
			log.Error("process wait error", "error", waitErr)
		}
		log.Info("child process exited", "exit_code", code)
		daemon.exitCode = code
		close(daemon.exitDone)
	})
	daemon.run()
}
