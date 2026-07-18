package coordinator

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	workerruntime "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
)

type recordingCommandRunner struct {
	calls            []string
	running          map[string]bool
	podPhase         map[string]string
	inspectNotFound  bool
	appliedManifests []string
}

func (r *recordingCommandRunner) Run(_ context.Context, name string, args ...string) (string, string, error) {
	call := name + " " + strings.Join(args, " ")
	r.calls = append(r.calls, call)

	if strings.Contains(call, "inspect -f {{.State.Running}}") {
		container := args[len(args)-1]
		if r.running != nil && r.running[container] {
			return "true\n", "", nil
		}
		return "false\n", "", nil
	}
	if strings.Contains(call, " inspect ") && !strings.Contains(call, "-f") {
		if r.inspectNotFound {
			return "", "Error: No such object", errors.New("exit 1")
		}
		return "{}", "", nil
	}
	if strings.Contains(call, "get pod") {
		pod := kubectlResourceName(args, "pod")
		if phase := r.podPhase[pod]; phase != "" {
			return phase, "", nil
		}
		return "", "NotFound", errors.New("exit 1")
	}
	if strings.Contains(call, " apply -f ") {
		if path := kubectlApplyPath(args); path != "" {
			if data, err := os.ReadFile(path); err == nil {
				r.appliedManifests = append(r.appliedManifests, string(data))
			}
		}
	}
	return "", "", nil
}

func kubectlResourceName(args []string, kind string) string {
	for i, arg := range args {
		if arg == kind && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func kubectlApplyPath(args []string) string {
	for i, arg := range args {
		if arg == "-f" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func TestDockerLauncherComposeModeUsesMultipleFiles(t *testing.T) {
	runner := &recordingCommandRunner{}
	dir := t.TempDir()
	l := NewDockerLauncher(dockerLauncherConfig{
		Binary:          "docker",
		ComposeDir:      dir,
		ComposeFiles:    []string{"docker-compose.yml", "docker-compose.runners.yml"},
		ComposeServices: map[string]string{"e2e-echo": "runner-e2e-echo"},
	}, nil)
	l.runner = runner
	if err := l.Launch(context.Background(), 1, "e2e-echo"); err != nil {
		t.Fatalf("Launch: %v", err)
	}
	call := runner.calls[0]
	if !strings.Contains(call, "docker-compose.runners.yml") {
		t.Fatalf("calls = %#v, want runners compose file", runner.calls)
	}
}

func TestDockerLauncherComposeMode(t *testing.T) {
	runner := &recordingCommandRunner{}
	l := NewDockerLauncher(dockerLauncherConfig{
		Binary:          "docker",
		ComposeDir:      t.TempDir(),
		ComposeServices: map[string]string{"do-agent": "runner-do-agent"},
	}, nil)
	l.runner = runner
	if err := l.Launch(context.Background(), 1, "do-agent"); err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if len(runner.calls) != 1 || !strings.Contains(runner.calls[0], "compose") {
		t.Fatalf("calls = %#v, want docker compose", runner.calls)
	}
	if !strings.Contains(runner.calls[0], "runner-do-agent") {
		t.Fatalf("calls = %#v, want agent-specific compose service", runner.calls)
	}
}

func TestDockerLauncherComposeModeRequiresAgentService(t *testing.T) {
	runner := &recordingCommandRunner{}
	l := NewDockerLauncher(dockerLauncherConfig{
		Binary:          "docker",
		ComposeDir:      t.TempDir(),
		ComposeServices: map[string]string{"codex-cli": "runner-codex"},
	}, nil)
	l.runner = runner
	err := l.Launch(context.Background(), 1, "claude-code")
	if err == nil {
		t.Fatal("expected missing compose service error")
	}
	if !strings.Contains(err.Error(), "COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES") {
		t.Fatalf("err = %v, want compose service mapping hint", err)
	}
}

func TestDockerLauncherRunMode(t *testing.T) {
	runner := &recordingCommandRunner{inspectNotFound: true}
	l := NewDockerLauncher(dockerLauncherConfig{
		Binary: "docker",
		ContainerEnv: runnerContainerEnv{
			AgentImages:  map[string]string{"do-agent": "agentsmesh/runner-do-agent:dev"},
			BackendURL:   "http://backend:8080",
			GRPCEndpoint: "backend:9443",
			OrgSlug:      "dev-org",
			NodeIDPrefix: "coord-",
		},
	}, nil)
	l.runner = runner
	if err := l.Launch(context.Background(), 42, "do-agent"); err != nil {
		t.Fatalf("Launch: %v", err)
	}
	joined := strings.Join(runner.calls, " | ")
	if !strings.Contains(joined, "docker run") {
		t.Fatalf("calls = %#v, want docker run", runner.calls)
	}
	if !strings.Contains(joined, "agentsmesh/runner-do-agent:dev") {
		t.Fatalf("calls = %#v, want agent-specific image", runner.calls)
	}
}

func TestDockerLauncherSkipsRunningContainer(t *testing.T) {
	name := runnerInstanceID(1, "do-agent")
	runner := &recordingCommandRunner{
		running: map[string]bool{name: true},
	}
	l := NewDockerLauncher(dockerLauncherConfig{
		Binary: "docker",
		ContainerEnv: runnerContainerEnv{
			AgentImages: map[string]string{"do-agent": "agentsmesh/runner-do-agent:dev"},
		},
	}, nil)
	l.runner = runner
	if err := l.Launch(context.Background(), 1, "do-agent"); err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if len(runner.calls) != 1 || !strings.Contains(runner.calls[0], "inspect -f") {
		t.Fatalf("calls = %#v, want single running inspect", runner.calls)
	}
}

func TestDockerLauncherRunModeRequiresAgentImage(t *testing.T) {
	runner := &recordingCommandRunner{inspectNotFound: true}
	l := NewDockerLauncher(dockerLauncherConfig{
		Binary: "docker",
		ContainerEnv: runnerContainerEnv{
			AgentImages: map[string]string{"codex-cli": "agentsmesh/runner-codex-cli:dev"},
		},
	}, nil)
	l.runner = runner
	err := l.Launch(context.Background(), 42, "claude-code")
	if err == nil {
		t.Fatal("expected missing agent image error")
	}
	if !strings.Contains(err.Error(), "COORDINATOR_RUNNER_IMAGES") {
		t.Fatalf("err = %v, want COORDINATOR_RUNNER_IMAGES hint", err)
	}
}

func TestK8sLauncherApplyPod(t *testing.T) {
	runner := &recordingCommandRunner{podPhase: map[string]string{}}
	l := NewK8sLauncher(k8sLauncherConfig{
		Kubectl:   "kubectl",
		Namespace: "agentsmesh",
		ContainerEnv: runnerContainerEnv{
			AgentImages:  map[string]string{"do-agent": "agentsmesh/runner-do-agent:prod"},
			BackendURL:   "http://backend",
			GRPCEndpoint: "backend:9443",
			OrgSlug:      "acme",
		},
		ReadyTimeoutSec: 30,
	}, nil)
	l.runner = runner
	if err := l.Launch(context.Background(), 7, "do-agent"); err != nil {
		t.Fatalf("Launch: %v", err)
	}
	joined := strings.Join(runner.calls, " | ")
	if !strings.Contains(joined, " apply -f ") {
		t.Fatalf("calls = %#v, want kubectl apply", runner.calls)
	}
	if !strings.Contains(joined, " wait --for=condition=Ready ") {
		t.Fatalf("calls = %#v, want kubectl wait", runner.calls)
	}
	if len(runner.appliedManifests) != 1 || !strings.Contains(runner.appliedManifests[0], "agentsmesh/runner-do-agent:prod") {
		t.Fatalf("applied manifests = %#v, want agent-specific image", runner.appliedManifests)
	}
}

func TestRenderRunnerPod(t *testing.T) {
	out, err := renderRunnerPod(k8sPodSpec{
		Name:              "amesh-runner-1-do-agent",
		Namespace:         "default",
		Image:             "agentsmesh/runner:latest",
		ImagePullPolicy:   "IfNotPresent",
		BackendURL:        "http://backend",
		GRPCEndpoint:      "backend:9443",
		RelayBaseURL:      "ws://relay",
		RunnerNodeID:      "coord-amesh-runner-1-do-agent",
		RunnerOrgSlug:     "dev-org",
		MaxConcurrentPods: 5,
		SSLHostPath:       "/data/ssl",
	})
	if err != nil {
		t.Fatalf("renderRunnerPod: %v", err)
	}
	text := string(out)
	for _, want := range []string{"kind: Pod", "agentsmesh/runner:latest", "/data/ssl"} {
		if !strings.Contains(text, want) {
			t.Fatalf("manifest missing %q:\n%s", want, text)
		}
	}
}

func TestSanitizeRunnerResourceName(t *testing.T) {
	got := sanitizeRunnerResourceName("amesh-runner-", "42/Do Agent!")
	if got != "amesh-runner-42-do-agent" {
		t.Fatalf("got %q", got)
	}
}

func TestNewRunnerLauncherFromEnv(t *testing.T) {
	t.Setenv("COORDINATOR_RUNNER_LAUNCHER", "docker")
	t.Setenv(
		"COORDINATOR_RUNNER_IMAGES",
		"claude-code="+testRunnerImage("claude-code")+
			",codex-cli="+testRunnerImage("codex-cli"),
	)
	launcher, kind, err := NewRunnerLauncherFromEnv(
		workerruntime.DefaultCatalog(),
		nil,
		nil,
	)
	if err != nil || kind != "docker" {
		t.Fatalf("docker launcher: kind=%q err=%v", kind, err)
	}
	if _, ok := launcher.(*DockerLauncher); !ok {
		t.Fatalf("got %T, want *DockerLauncher", launcher)
	}

	t.Setenv("COORDINATOR_RUNNER_LAUNCHER", "k8s")
	t.Setenv(
		"COORDINATOR_RUNNER_IMAGES",
		"e2e-echo=agentsmesh/runner-e2e-echo@sha256:077eb4511113ddb80dd8e09d7b46ffe3668d6b69d1840c1cbe849e97595087fa",
	)
	launcher, kind, err = NewRunnerLauncherFromEnv(
		workerruntime.DefaultCatalog(),
		nil,
		nil,
	)
	if err != nil || kind != "k8s" {
		t.Fatalf("k8s launcher: kind=%q err=%v", kind, err)
	}
	if _, ok := launcher.(*K8sLauncher); !ok {
		t.Fatalf("got %T, want *K8sLauncher", launcher)
	}
}

func TestLoadRunnerContainerEnvParsesAgentImages(t *testing.T) {
	t.Setenv(
		"COORDINATOR_RUNNER_IMAGES",
		"claude-code="+testRunnerImage("claude-code")+
			", codex-cli="+testRunnerImage("codex-cli"),
	)
	env, err := loadRunnerContainerEnv()
	if err != nil {
		t.Fatalf("loadRunnerContainerEnv: %v", err)
	}
	if got := env.AgentImages["claude-code"]; got != testRunnerImage("claude-code") {
		t.Fatalf("claude image = %q", got)
	}
	if got := env.AgentImages["codex-cli"]; got != testRunnerImage("codex-cli") {
		t.Fatalf("codex image = %q", got)
	}
}

func TestLoadDockerLauncherConfigParsesComposeServices(t *testing.T) {
	t.Setenv("COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES", "claude-code=runner-claude, codex-cli=runner-codex")
	cfg, err := loadDockerLauncherConfig()
	if err != nil {
		t.Fatalf("loadDockerLauncherConfig: %v", err)
	}
	if got := cfg.ComposeServices["claude-code"]; got != "runner-claude" {
		t.Fatalf("claude service = %q", got)
	}
	if got := cfg.ComposeServices["codex-cli"]; got != "runner-codex" {
		t.Fatalf("codex service = %q", got)
	}
}

func TestNewRunnerLauncherRejectsMutableImageReference(t *testing.T) {
	t.Setenv("COORDINATOR_RUNNER_LAUNCHER", "k8s")
	t.Setenv("COORDINATOR_RUNNER_IMAGES", "do-agent=repo.example/runner-do-agent:latest")

	_, _, err := NewRunnerLauncherFromEnv(
		workerruntime.DefaultCatalog(),
		nil,
		nil,
	)

	if err == nil || !strings.Contains(err.Error(), "immutable sha256") {
		t.Fatalf("err = %v, want immutable image validation", err)
	}
}

func testRunnerImage(runtime string) string {
	return "repo.example/runner-" + runtime + "@sha256:" + strings.Repeat("a", 64)
}
