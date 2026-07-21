package coordinator

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type K8sLauncher struct {
	cfg    k8sLauncherConfig
	runner commandRunner
	logger *slog.Logger
}

func NewK8sLauncher(cfg k8sLauncherConfig, logger *slog.Logger) *K8sLauncher {
	if logger == nil {
		logger = slog.Default()
	}
	return &K8sLauncher{cfg: cfg, runner: execCommandRunner{}, logger: logger}
}

func (l *K8sLauncher) Launch(ctx context.Context, orgID int64, agentSlug string) error {
	image, err := l.cfg.ContainerEnv.imageForAgent(agentSlug)
	if err != nil {
		return err
	}
	name := runnerInstanceID(orgID, agentSlug)
	phase, err := podPhase(ctx, l, name)
	if err != nil {
		return err
	}
	switch phase {
	case "Running", "Pending":
		l.logger.Debug("k8s runner pod already exists", "pod", name, "phase", phase)
		return nil
	case "Succeeded", "Failed", "Unknown":
		deleteArgs := append(kubectlBaseArgs(l.cfg), "delete", "pod", name, "--ignore-not-found", "--wait=false")
		_, _, _ = l.runner.Run(ctx, l.cfg.Kubectl, deleteArgs...)
	}

	manifest, err := renderRunnerPod(k8sPodSpec{
		Name:              name,
		Namespace:         l.cfg.Namespace,
		Image:             image,
		ImagePullPolicy:   l.cfg.ImagePullPolicy,
		BackendURL:        l.cfg.ContainerEnv.BackendURL,
		GRPCEndpoint:      l.cfg.ContainerEnv.GRPCEndpoint,
		RelayBaseURL:      l.cfg.ContainerEnv.RelayBaseURL,
		RunnerNodeID:      l.cfg.ContainerEnv.NodeIDPrefix + name,
		RunnerOrgSlug:     l.cfg.ContainerEnv.OrgSlug,
		MaxConcurrentPods: l.cfg.ContainerEnv.MaxConcurrentPods,
		SSLHostPath:       l.cfg.SSLHostPath,
		SSLSecretName:     l.cfg.SSLSecretName,
	})
	if err != nil {
		return fmt.Errorf("render runner pod: %w", err)
	}
	manifestPath, err := writeTempManifest(manifest)
	if err != nil {
		return err
	}
	defer os.Remove(manifestPath)

	applyArgs := append(kubectlBaseArgs(l.cfg), "apply", "-f", manifestPath)
	if _, stderr, err := l.runner.Run(ctx, l.cfg.Kubectl, applyArgs...); err != nil {
		return fmt.Errorf("kubectl apply: %w: %s", err, strings.TrimSpace(stderr))
	}
	waitArgs := append(
		kubectlBaseArgs(l.cfg),
		"wait", "--for=condition=Ready", "pod/"+name,
		fmt.Sprintf("--timeout=%ds", l.cfg.ReadyTimeoutSec),
	)
	if _, stderr, err := l.runner.Run(ctx, l.cfg.Kubectl, waitArgs...); err != nil {
		return fmt.Errorf("kubectl wait: %w: %s", err, strings.TrimSpace(stderr))
	}
	l.logger.Info("k8s runner pod ready", "pod", name, "namespace", l.cfg.Namespace)
	return nil
}

func writeTempManifest(data []byte) (string, error) {
	file, err := os.CreateTemp("", "agentcloud-runner-*.yaml")
	if err != nil {
		return "", err
	}
	path := file.Name()
	if _, err := file.Write(data); err != nil {
		file.Close()
		os.Remove(path)
		return "", err
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}
