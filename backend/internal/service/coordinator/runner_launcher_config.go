package coordinator

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	runtimedomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerruntime"
)

type runnerContainerEnv struct {
	BackendURL        string
	GRPCEndpoint      string
	RelayBaseURL      string
	OrgSlug           string
	NodeIDPrefix      string
	MaxConcurrentPods int
	AgentImages       map[string]string
}

type dockerLauncherConfig struct {
	Binary          string
	Network         string
	ComposeDir      string
	ComposeFiles    []string
	ComposeServices map[string]string
	SSLHostPath     string
	EntrypointPath  string
	ExtraVolumes    []string
	ContainerEnv    runnerContainerEnv
}

type k8sLauncherConfig struct {
	Kubectl         string
	Kubeconfig      string
	Namespace       string
	ImagePullPolicy string
	ReadyTimeoutSec int
	SSLHostPath     string
	SSLSecretName   string
	ContainerEnv    runnerContainerEnv
}

func loadRunnerContainerEnv() (runnerContainerEnv, error) {
	maxPods := 10
	if v := strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_MAX_CONCURRENT_PODS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxPods = n
		}
	}
	_, images, err := runtimedomain.ParseRuntimeImageReferences(
		os.Getenv(runtimedomain.RuntimeImageReferencesEnv),
	)
	if err != nil {
		return runnerContainerEnv{}, fmt.Errorf("coordinator runtime images: %w", err)
	}
	return runnerContainerEnv{
		BackendURL:        strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_BACKEND_URL")),
		GRPCEndpoint:      strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_GRPC_ENDPOINT")),
		RelayBaseURL:      strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_RELAY_BASE_URL")),
		OrgSlug:           strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_ORG_SLUG")),
		NodeIDPrefix:      defaultString(os.Getenv("COORDINATOR_RUNNER_NODE_ID_PREFIX"), "coord-runner-"),
		MaxConcurrentPods: maxPods,
		AgentImages:       images,
	}, nil
}

func loadDockerLauncherConfig() (dockerLauncherConfig, error) {
	containerEnv, err := loadRunnerContainerEnv()
	if err != nil {
		return dockerLauncherConfig{}, err
	}
	extra := []string{}
	if raw := strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_DOCKER_EXTRA_VOLUMES")); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			if v := strings.TrimSpace(part); v != "" {
				extra = append(extra, v)
			}
		}
	}
	return dockerLauncherConfig{
		Binary:          defaultString(os.Getenv("COORDINATOR_RUNNER_DOCKER_BINARY"), "docker"),
		Network:         strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_DOCKER_NETWORK")),
		ComposeDir:      strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_DOCKER_COMPOSE_DIR")),
		ComposeFiles:    parseComposeFiles(os.Getenv("COORDINATOR_RUNNER_DOCKER_COMPOSE_FILES")),
		ComposeServices: parseLauncherMap(os.Getenv("COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES")),
		SSLHostPath:     strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_DOCKER_SSL_HOST_PATH")),
		EntrypointPath:  strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_DOCKER_ENTRYPOINT_HOST_PATH")),
		ExtraVolumes:    extra,
		ContainerEnv:    containerEnv,
	}, nil
}

func parseComposeFiles(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	out := make([]string, 0, 2)
	for _, part := range strings.Split(raw, ",") {
		if f := strings.TrimSpace(part); f != "" {
			out = append(out, f)
		}
	}
	return out
}

func loadK8sLauncherConfig() (k8sLauncherConfig, error) {
	containerEnv, err := loadRunnerContainerEnv()
	if err != nil {
		return k8sLauncherConfig{}, err
	}
	timeout := 120
	if v := strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_K8S_READY_TIMEOUT_SECONDS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeout = n
		}
	}
	return k8sLauncherConfig{
		Kubectl:         defaultString(os.Getenv("COORDINATOR_RUNNER_K8S_KUBECTL"), "kubectl"),
		Kubeconfig:      strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_K8S_KUBECONFIG")),
		Namespace:       defaultString(os.Getenv("COORDINATOR_RUNNER_K8S_NAMESPACE"), "default"),
		ImagePullPolicy: defaultString(os.Getenv("COORDINATOR_RUNNER_K8S_IMAGE_PULL_POLICY"), "IfNotPresent"),
		ReadyTimeoutSec: timeout,
		SSLHostPath:     strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_K8S_SSL_HOST_PATH")),
		SSLSecretName:   strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_K8S_SSL_SECRET")),
		ContainerEnv:    containerEnv,
	}, nil
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func parseLauncherMap(raw string) map[string]string {
	out := map[string]string{}
	for _, part := range strings.Split(raw, ",") {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			out[key] = value
		}
	}
	return out
}
