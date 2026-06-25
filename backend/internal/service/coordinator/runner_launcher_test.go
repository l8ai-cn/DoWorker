package coordinator

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type recordingCommandRunner struct {
	calls           []string
	running         map[string]bool
	podPhase        map[string]string
	inspectNotFound bool
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

func TestDockerLauncherComposeMode(t *testing.T) {
	runner := &recordingCommandRunner{}
	l := NewDockerLauncher(dockerLauncherConfig{
		Binary:         "docker",
		ComposeDir:     t.TempDir(),
		ComposeService: "runner",
	}, nil)
	l.runner = runner
	if err := l.Launch(context.Background(), 1, "do-agent"); err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if len(runner.calls) != 1 || !strings.Contains(runner.calls[0], "compose") {
		t.Fatalf("calls = %#v, want docker compose", runner.calls)
	}
}

func TestDockerLauncherRunMode(t *testing.T) {
	runner := &recordingCommandRunner{inspectNotFound: true}
	l := NewDockerLauncher(dockerLauncherConfig{
		Binary: "docker",
		ContainerEnv: runnerContainerEnv{
			Image:        "agentsmesh/runner:dev",
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
}

func TestDockerLauncherSkipsRunningContainer(t *testing.T) {
	name := runnerInstanceID(1, "do-agent")
	runner := &recordingCommandRunner{
		running: map[string]bool{name: true},
	}
	l := NewDockerLauncher(dockerLauncherConfig{
		Binary: "docker",
		ContainerEnv: runnerContainerEnv{
			Image: "agentsmesh/runner:dev",
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

func TestK8sLauncherApplyPod(t *testing.T) {
	runner := &recordingCommandRunner{podPhase: map[string]string{}}
	l := NewK8sLauncher(k8sLauncherConfig{
		Kubectl:   "kubectl",
		Namespace: "agentsmesh",
		ContainerEnv: runnerContainerEnv{
			Image:        "agentsmesh/runner:prod",
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
	launcher, kind, err := NewRunnerLauncherFromEnv(nil)
	if err != nil || kind != "docker" {
		t.Fatalf("docker launcher: kind=%q err=%v", kind, err)
	}
	if _, ok := launcher.(*DockerLauncher); !ok {
		t.Fatalf("got %T, want *DockerLauncher", launcher)
	}

	t.Setenv("COORDINATOR_RUNNER_LAUNCHER", "k8s")
	launcher, kind, err = NewRunnerLauncherFromEnv(nil)
	if err != nil || kind != "k8s" {
		t.Fatalf("k8s launcher: kind=%q err=%v", kind, err)
	}
	if _, ok := launcher.(*K8sLauncher); !ok {
		t.Fatalf("got %T, want *K8sLauncher", launcher)
	}
}
