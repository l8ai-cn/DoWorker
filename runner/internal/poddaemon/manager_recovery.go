package poddaemon

import (
	"fmt"
	"os"
	"path/filepath"
)

func (m *PodDaemonManager) RecoverSessions() ([]*PodDaemonState, error) {
	entries, err := os.ReadDir(m.sandboxesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read sandboxes dir: %w", err)
	}

	var sessions []*PodDaemonState
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sandboxPath := filepath.Join(m.sandboxesDir, entry.Name())
		state, err := LoadState(sandboxPath)
		if err == nil {
			sessions = append(sessions, state)
		}
	}
	return sessions, nil
}

func (m *PodDaemonManager) CleanupSession(sandboxPath string) error {
	return DeleteState(sandboxPath)
}
