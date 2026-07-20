package runner

import (
	"testing"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/config"
	"github.com/anthropics/agentsmesh/runner/internal/testutil"
	"github.com/anthropics/agentsmesh/runner/internal/workspace"
)

// Tests for OnCreatePod operations

func TestOnCreatePodSuccess(t *testing.T) {
	tempDir := t.TempDir()
	store := NewInMemoryPodStore()
	mockConn := client.NewMockConnection()

	ws, err := workspace.NewManager(tempDir, "")
	if err != nil {
		t.Skipf("Could not create workspace manager: %v", err)
	}

	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:     tempDir,
		},
		workspace: ws,
	}

	handler := NewRunnerMessageHandler(runner, store, mockConn)

	sleepCmd, sleepArgs := testutil.SleepCommand(10)
	cmd := &runnerv1.CreatePodCommand{
		PodKey:          "test-pod-1",
		LaunchCommand:   sleepCmd,
		LaunchArgs:      sleepArgs,
		AgentfileSource: "AGENT " + sleepCmd + "\nPROMPT_POSITION prepend\n",
	}

	err = handler.OnCreatePod(cmd)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify pod was created
	pod, ok := store.Get("test-pod-1")
	if !ok {
		t.Error("pod should be stored")
	} else {
		if pod.GetStatus() != PodStatusRunning {
			t.Errorf("pod status = %s, want running", pod.GetStatus())
		}
		// Clean up terminal
		if comps := testPTYComponents(pod); comps != nil && comps.Terminal != nil {
			comps.Terminal.Stop()
		}
	}

	// Verify pod_created event was sent
	events := mockConn.GetEvents()
	hasCreated := false
	for _, e := range events {
		if e.Type == client.MsgTypePodCreated {
			hasCreated = true
			break
		}
	}
	if !hasCreated {
		t.Error("should have sent pod_created event")
	}
}

func TestOnCreatePodInvalidCommand(t *testing.T) {
	tempDir := t.TempDir()
	store := NewInMemoryPodStore()
	mockConn := client.NewMockConnection()

	ws, err := workspace.NewManager(tempDir, "")
	if err != nil {
		t.Skipf("Could not create workspace manager: %v", err)
	}

	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:     tempDir,
		},
		workspace: ws,
	}

	handler := NewRunnerMessageHandler(runner, store, mockConn)

	cmd := &runnerv1.CreatePodCommand{
		PodKey:          "invalid-cmd-pod",
		LaunchCommand:   "/nonexistent/command/path",
		AgentfileSource: "AGENT test\nPROMPT_POSITION prepend\n",
	}

	err = handler.OnCreatePod(cmd)
	// Command may or may not fail depending on OS
	t.Logf("OnCreatePod with invalid command: %v", err)
}

func TestOnCreatePodPreservesJoinedPodErrorCode(t *testing.T) {
	tempDir := t.TempDir()
	store := NewInMemoryPodStore()
	mockConn := client.NewMockConnection()
	runner := &Runner{
		cfg: &config.Config{
			MaxConcurrentPods: 10,
			WorkspaceRoot:     tempDir,
		},
		workspace: &mockWorkspace{},
	}
	handler := NewRunnerMessageHandler(runner, store, mockConn)
	cmd := &runnerv1.CreatePodCommand{
		PodKey:        "joined-pod-error",
		LaunchCommand: "echo",
		SandboxConfig: &runnerv1.SandboxConfig{
			HttpCloneUrl:    "https://github.com/org/repo.git",
			SourceCommitSha: testCommitSHA,
			CredentialType:  "unsupported",
		},
	}

	if err := handler.OnCreatePod(cmd); err == nil {
		t.Fatal("expected create failure")
	}
	for _, event := range mockConn.GetEvents() {
		if event.Type != client.MessageType("error") {
			continue
		}
		data, ok := event.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("unexpected error event: %#v", event.Data)
		}
		if data["code"] != client.ErrCodeGitAuth {
			t.Fatalf("error code = %v, want %s", data["code"], client.ErrCodeGitAuth)
		}
		return
	}
	t.Fatal("expected structured error event")
}
