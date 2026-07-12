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
		config, err := loadDockerLauncherConfig()
		if err != nil {
			return nil, "", err
		}
		return NewDockerLauncher(config, logger), "docker", nil
	case "k8s", "kubernetes":
		config, err := loadK8sLauncherConfig()
		if err != nil {
			return nil, "", err
		}
		return NewK8sLauncher(config, logger), "k8s", nil
	default:
		return NewScriptLauncher(raw, logger), "script", nil
	}
}
