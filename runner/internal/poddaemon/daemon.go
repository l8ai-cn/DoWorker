package poddaemon

import (
	"encoding/binary"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/safego"
)

// daemonServer manages the IPC server and PTY I/O forwarding.
type daemonServer struct {
	proc     daemonProcess
	listener net.Listener
	exitCode int           // set before exitDone is closed
	exitDone chan struct{} // closed when child process exits (broadcast)
	orphanCh chan struct{} // closed when state file is deleted (orphan protection)
	log      *slog.Logger
	state    *PodDaemonState

	// orphanCheckInterval controls how often orphanChecker polls.
	// Defaults to 60s in production; tests can inject a shorter value.
	orphanCheckInterval time.Duration

	// clientMu protects the client pointer AND the output history buffer. Hold
	// briefly to read/swap the pointer or append history — never hold while
	// doing network I/O.
	clientMu sync.Mutex
	client   net.Conn

	// history is a bounded ring of recent PTY output, replayed to a client on
	// attach. Without it, output the child produced before the runner attaches
	// (the runner connects a few hundred ms after the daemon spawns the child)
	// is dropped — invisible for agents that redraw continuously, fatal for an
	// agent that prints once then idles. Also lets a runner that restarted
	// (session recovery) repaint the terminal. Guarded by clientMu.
	history []byte

	// connWriteMu serializes writes to the IPC connection. This is separate
	// from clientMu so that ptyReader's potentially slow data writes don't
	// block control-plane operations (Pong, Exit notification) from acquiring
	// the client pointer.
	connWriteMu sync.Mutex
}

// outputHistoryLimit caps the replay buffer (bytes). Large enough to carry a
// startup banner plus a full-screen redraw; the runner-side VirtualTerminal
// owns the authoritative scrollback.
const outputHistoryLimit = 256 * 1024

// appendHistoryLocked records output for replay to a future attach. Caller
// must hold clientMu. Re-copies on trim so the backing array stays bounded.
func (d *daemonServer) appendHistoryLocked(data []byte) {
	d.history = append(d.history, data...)
	if len(d.history) > outputHistoryLimit {
		trimmed := make([]byte, outputHistoryLimit)
		copy(trimmed, d.history[len(d.history)-outputHistoryLimit:])
		d.history = trimmed
	}
}

func (d *daemonServer) run() {
	// PTY reader: must keep running, auto-restart on panic (otherwise terminal freezes)
	safego.GoLoop("daemon-pty-reader", d.ptyReader, 0)

	// Accept loop: must keep running, auto-restart on panic (otherwise Runner can't reconnect)
	safego.GoLoop("daemon-accept-loop", d.acceptLoop, 0)

	// Orphan protection: must keep running, auto-restart on panic (otherwise daemon leaks)
	safego.GoLoop("daemon-orphan-checker", d.orphanChecker, 0)

	// Wait for child exit or orphan signal
	select {
	case <-d.exitDone:
		d.log.Info("daemon shutting down (child exited)", "exit_code", d.exitCode)

		// Notify connected client about exit
		d.clientMu.Lock()
		client := d.client
		d.clientMu.Unlock()
		if client != nil {
			d.connWriteMu.Lock()
			payload := make([]byte, 4)
			binary.BigEndian.PutUint32(payload, uint32(int32(d.exitCode)))
			if err := WriteMessage(client, MsgExit, payload); err != nil {
				d.log.Debug("failed to send exit notification", "error", err)
			}
			d.connWriteMu.Unlock()
		}

	case <-d.orphanCh:
		d.log.Info("daemon shutting down (state file deleted, orphan protection)")
		// Kill the child process and exit
		d.proc.GracefulStop()
		select {
		case <-d.exitDone:
		case <-time.After(5 * time.Second):
			d.proc.Kill()
		}
	}
}

// orphanChecker periodically checks if the state file (pod_daemon.json) still
// exists. If the file has been deleted (e.g., by CleanupSession), the daemon
// is considered orphaned and shuts down gracefully.
//
// Behavior:
//   - Polls every 60 seconds (configurable via orphanCheckInterval for tests,
//     or _AGENTSMESH_ORPHAN_CHECK_INTERVAL_SEC env var).
//   - On detection: closes orphanCh → run() triggers GracefulStop on child
//     process, waits 5s, then kills if needed.
//   - Stops automatically when the child process exits (exitDone).
func (d *daemonServer) orphanChecker() {
	interval := d.orphanCheckInterval
	if interval <= 0 {
		interval = 60 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, err := os.Stat(StatePath(d.state.SandboxPath)); os.IsNotExist(err) {
				d.log.Info("state file deleted, triggering orphan protection")
				close(d.orphanCh)
				return
			}
		case <-d.exitDone:
			return
		}
	}
}
