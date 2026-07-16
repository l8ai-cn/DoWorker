package poddaemon

import (
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/process"
)

func (m *PodDaemonManager) waitForDaemon(
	sandboxPath, authToken string,
	pid int,
) (*daemonPTY, *PodDaemonState, error) {
	const maxAttempts = 50
	const retryDelay = 100 * time.Millisecond

	var lastErr error
	for range maxAttempts {
		state, err := LoadState(sandboxPath)
		if err == nil && state.IPCAddr != "" {
			if state.AuthToken != authToken {
				return nil, nil, fmt.Errorf(
					"auth token mismatch in state file (possible tampering)",
				)
			}
			dpty, connectErr := connectDaemon(connectOpts{
				Addr: state.IPCAddr, AuthToken: authToken,
			})
			if connectErr == nil {
				return dpty, state, nil
			}
			lastErr = connectErr
		} else if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("daemon has not written IPC address yet")
		}
		if pid > 0 && process.IsAlive(pid) != nil {
			return nil, nil, fmt.Errorf(
				"daemon process (pid %d) exited before IPC ready: %w",
				pid,
				lastErr,
			)
		}
		time.Sleep(retryDelay)
	}
	return nil, nil, fmt.Errorf(
		"daemon did not become ready within %v: %w",
		time.Duration(maxAttempts)*retryDelay,
		lastErr,
	)
}

func (m *PodDaemonManager) AttachSession(
	state *PodDaemonState,
) (*daemonPTY, error) {
	return connectDaemon(connectOpts{
		Addr: state.IPCAddr, AuthToken: state.AuthToken,
	})
}
