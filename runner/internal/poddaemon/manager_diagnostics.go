package poddaemon

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/l8ai-cn/agentcloud/runner/internal/process"
)

const daemonLogFile = "pod_daemon.log"

func captureDaemonLog(log *slog.Logger, sandboxPath, podKey string) {
	logPath := filepath.Join(sandboxPath, daemonLogFile)
	data, err := os.ReadFile(logPath)
	if err != nil {
		log.Error(
			"pod daemon log unavailable",
			"pod_key", podKey,
			"path", logPath,
			"error", err,
		)
		return
	}
	if len(data) == 0 {
		log.Error(
			"pod daemon log is empty (daemon likely crashed before any Go code executed)",
			"pod_key", podKey,
			"path", logPath,
		)
		return
	}
	const maxLen = 2048
	if len(data) > maxLen {
		data = data[len(data)-maxLen:]
	}
	log.Error(
		"pod daemon log (process exited before IPC ready)",
		"pod_key", podKey,
		"log", strings.TrimSpace(string(data)),
	)
}

func daemonProcessStatus(pid int) string {
	if pid <= 0 {
		return "unknown"
	}
	if process.IsAlive(pid) == nil {
		return "alive"
	}
	return "dead"
}
