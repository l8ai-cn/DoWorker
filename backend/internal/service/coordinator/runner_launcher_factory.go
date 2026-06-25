package coordinator

import (
	"log/slog"
	"os"
	"strings"
)

// NewRunnerLauncherFromEnv builds a RunnerLauncher from COORDINATOR_RUNNER_LAUNCHER.
// Values: docker | k8s | kubernetes | path/to/script.sh
func NewRunnerLauncherFromEnv(logger *slog.Logger) (RunnerLauncher, string, error) {
	raw := strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_LAUNCHER"))
	if raw == "" {
		return nil, "", nil
	}
	if logger == nil {
		logger = slog.Default()
	}
	switch strings.ToLower(raw) {
	case "docker":
		return NewDockerLauncher(loadDockerLauncherConfig(), logger), "docker", nil
	case "k8s", "kubernetes":
		return NewK8sLauncher(loadK8sLauncherConfig(), logger), "k8s", nil
	default:
		return NewScriptLauncher(raw, logger), "script", nil
	}
}
