package coordinator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type DockerLauncher struct {
	cfg    dockerLauncherConfig
	runner commandRunner
	logger *slog.Logger
}

func NewDockerLauncher(cfg dockerLauncherConfig, logger *slog.Logger) *DockerLauncher {
	if logger == nil {
		logger = slog.Default()
	}
	return &DockerLauncher{cfg: cfg, runner: execCommandRunner{}, logger: logger}
}

func (l *DockerLauncher) Launch(ctx context.Context, orgID int64, agentSlug string) error {
	if strings.TrimSpace(l.cfg.ComposeDir) != "" {
		return l.launchCompose(ctx)
	}
	return l.launchRun(ctx, orgID, agentSlug)
}

func (l *DockerLauncher) launchCompose(ctx context.Context) error {
	dir, err := filepath.Abs(l.cfg.ComposeDir)
	if err != nil {
		return fmt.Errorf("docker compose dir: %w", err)
	}
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("docker compose dir %q: %w", dir, err)
	}
	args := []string{"compose", "-f", filepath.Join(dir, "docker-compose.yml"), "up", "-d", l.cfg.ComposeService}
	if _, stderr, err := l.runner.Run(ctx, l.cfg.Binary, args...); err != nil {
		return fmt.Errorf("docker compose up: %w: %s", err, strings.TrimSpace(stderr))
	}
	l.logger.Info("docker compose runner started", "dir", dir, "service", l.cfg.ComposeService)
	return nil
}

func (l *DockerLauncher) launchRun(ctx context.Context, orgID int64, agentSlug string) error {
	if l.cfg.ContainerEnv.Image == "" {
		return errors.New("coordinator: COORDINATOR_RUNNER_IMAGE is required for docker run")
	}
	name := runnerInstanceID(orgID, agentSlug)
	if running, err := l.containerRunning(ctx, name); err != nil {
		return err
	} else if running {
		return nil
	}
	if exists, err := l.containerExists(ctx, name); err != nil {
		return err
	} else if exists {
		if _, stderr, err := l.runner.Run(ctx, l.cfg.Binary, "start", name); err != nil {
			return fmt.Errorf("docker start %q: %w: %s", name, err, strings.TrimSpace(stderr))
		}
		l.logger.Info("docker runner restarted", "container", name)
		return nil
	}

	nodeID := l.cfg.ContainerEnv.NodeIDPrefix + name
	args := []string{
		"run", "-d",
		"--name", name,
		"--restart", "unless-stopped",
		"-e", "BACKEND_URL=" + l.cfg.ContainerEnv.BackendURL,
		"-e", "GRPC_ENDPOINT=" + l.cfg.ContainerEnv.GRPCEndpoint,
		"-e", "RELAY_BASE_URL=" + l.cfg.ContainerEnv.RelayBaseURL,
		"-e", "RUNNER_NODE_ID=" + nodeID,
		"-e", "RUNNER_ORG_SLUG=" + l.cfg.ContainerEnv.OrgSlug,
		"-e", "MAX_CONCURRENT_PODS=" + strconv.Itoa(l.cfg.ContainerEnv.MaxConcurrentPods),
		"-e", "AGENTSMESH_MCP_BIND=0.0.0.0",
	}
	if network := strings.TrimSpace(l.cfg.Network); network != "" {
		args = append(args, "--network", network)
	}
	if ssl := strings.TrimSpace(l.cfg.SSLHostPath); ssl != "" {
		args = append(args, "-v", ssl+":/app/ssl:ro")
	}
	if entrypoint := strings.TrimSpace(l.cfg.EntrypointPath); entrypoint != "" {
		args = append(args, "-v", entrypoint+":/usr/local/bin/runner-entrypoint.sh:ro")
	}
	for _, vol := range l.cfg.ExtraVolumes {
		args = append(args, "-v", vol)
	}
	args = append(args, l.cfg.ContainerEnv.Image)
	if entrypoint := strings.TrimSpace(l.cfg.EntrypointPath); entrypoint != "" {
		args = append(args, "/usr/local/bin/runner-entrypoint.sh")
	}
	if _, stderr, err := l.runner.Run(ctx, l.cfg.Binary, args...); err != nil {
		return fmt.Errorf("docker run: %w: %s", err, strings.TrimSpace(stderr))
	}
	l.logger.Info("docker runner created", "container", name, "image", l.cfg.ContainerEnv.Image)
	return nil
}

func (l *DockerLauncher) containerExists(ctx context.Context, name string) (bool, error) {
	_, stderr, err := l.runner.Run(ctx, l.cfg.Binary, "inspect", name)
	if err == nil {
		return true, nil
	}
	if strings.Contains(stderr, "No such object") || strings.Contains(stderr, "Error: No such") {
		return false, nil
	}
	return false, fmt.Errorf("docker inspect: %w: %s", err, strings.TrimSpace(stderr))
}

func (l *DockerLauncher) containerRunning(ctx context.Context, name string) (bool, error) {
	stdout, stderr, err := l.runner.Run(ctx, l.cfg.Binary, "inspect", "-f", "{{.State.Running}}", name)
	if err != nil {
		if strings.Contains(stderr, "No such object") {
			return false, nil
		}
		return false, fmt.Errorf("docker inspect running: %w: %s", err, strings.TrimSpace(stderr))
	}
	return strings.TrimSpace(stdout) == "true", nil
}
