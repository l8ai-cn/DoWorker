package coordinator

import (
	"os"
	"strconv"
	"strings"
)

type runnerContainerEnv struct {
	BackendURL        string
	GRPCEndpoint      string
	RelayBaseURL      string
	OrgSlug           string
	NodeIDPrefix      string
	MaxConcurrentPods int
	Image             string
}

type dockerLauncherConfig struct {
	Binary         string
	Network        string
	ComposeDir     string
	ComposeService string
	SSLHostPath    string
	EntrypointPath string
	ExtraVolumes   []string
	ContainerEnv   runnerContainerEnv
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

func loadRunnerContainerEnv() runnerContainerEnv {
	maxPods := 10
	if v := strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_MAX_CONCURRENT_PODS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxPods = n
		}
	}
	return runnerContainerEnv{
		BackendURL:        strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_BACKEND_URL")),
		GRPCEndpoint:      strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_GRPC_ENDPOINT")),
		RelayBaseURL:      strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_RELAY_BASE_URL")),
		OrgSlug:           strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_ORG_SLUG")),
		NodeIDPrefix:      defaultString(os.Getenv("COORDINATOR_RUNNER_NODE_ID_PREFIX"), "coord-runner-"),
		MaxConcurrentPods: maxPods,
		Image:             strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_IMAGE")),
	}
}

func loadDockerLauncherConfig() dockerLauncherConfig {
	extra := []string{}
	if raw := strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_DOCKER_EXTRA_VOLUMES")); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			if v := strings.TrimSpace(part); v != "" {
				extra = append(extra, v)
			}
		}
	}
	return dockerLauncherConfig{
		Binary:         defaultString(os.Getenv("COORDINATOR_RUNNER_DOCKER_BINARY"), "docker"),
		Network:        strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_DOCKER_NETWORK")),
		ComposeDir:     strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_DOCKER_COMPOSE_DIR")),
		ComposeService: defaultString(os.Getenv("COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICE"), "runner"),
		SSLHostPath:    strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_DOCKER_SSL_HOST_PATH")),
		EntrypointPath: strings.TrimSpace(os.Getenv("COORDINATOR_RUNNER_DOCKER_ENTRYPOINT_HOST_PATH")),
		ExtraVolumes:   extra,
		ContainerEnv:   loadRunnerContainerEnv(),
	}
}

func loadK8sLauncherConfig() k8sLauncherConfig {
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
		ContainerEnv:    loadRunnerContainerEnv(),
	}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}
