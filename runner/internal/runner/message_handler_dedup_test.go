package runner

import (
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/config"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func TestOnSendPrompt_DuplicateCommandID_Dropped(t *testing.T) {
	base := &sendPromptMockIO{mode: InteractionModePTY}
	io := &ptyTerminalMock{sendPromptMockIO: base}
	pod := &Pod{PodKey: "dedup-pod", InteractionMode: InteractionModePTY, IO: io}

	store := NewInMemoryPodStore()
	store.Put(pod.PodKey, pod)
	h := &RunnerMessageHandler{podStore: store}

	cmd := &runnerv1.SendPromptCommand{PodKey: pod.PodKey, Prompt: "hello", CommandId: "cmd-1"}
	if err := h.OnSendPrompt(cmd); err != nil {
		t.Fatalf("first OnSendPrompt error: %v", err)
	}
	if err := h.OnSendPrompt(cmd); err != nil {
		t.Fatalf("duplicate OnSendPrompt should be absorbed, got error: %v", err)
	}

	base.mu.Lock()
	defer base.mu.Unlock()
	if len(base.inputs) != 1 {
		t.Fatalf("duplicate command_id must write to PTY once; got %d writes", len(base.inputs))
	}
}

func TestOnSendPrompt_EmptyCommandID_NoDedup(t *testing.T) {
	base := &sendPromptMockIO{mode: InteractionModePTY}
	io := &ptyTerminalMock{sendPromptMockIO: base}
	pod := &Pod{PodKey: "nodedup-pod", InteractionMode: InteractionModePTY, IO: io}

	store := NewInMemoryPodStore()
	store.Put(pod.PodKey, pod)
	h := &RunnerMessageHandler{podStore: store}

	cmd := &runnerv1.SendPromptCommand{PodKey: pod.PodKey, Prompt: "hello"}
	for i := 0; i < 2; i++ {
		if err := h.OnSendPrompt(cmd); err != nil {
			t.Fatalf("OnSendPrompt #%d error: %v", i+1, err)
		}
	}

	base.mu.Lock()
	defer base.mu.Unlock()
	if len(base.inputs) != 2 {
		t.Fatalf("empty command_id must not dedup; got %d writes, want 2", len(base.inputs))
	}
}

func TestOnCreatePod_DuplicateRunningPod_AbsorbedWithAck(t *testing.T) {
	store := NewInMemoryPodStore()
	mockConn := client.NewMockConnection()
	r := &Runner{cfg: &config.Config{}}
	h := NewRunnerMessageHandler(r, store, mockConn)

	existing := &Pod{PodKey: "dup-pod", Status: PodStatusRunning, SandboxPath: "/tmp/sandbox", Branch: "feat"}
	store.Put("dup-pod", existing)

	if err := h.OnCreatePod(&runnerv1.CreatePodCommand{PodKey: "dup-pod"}); err != nil {
		t.Fatalf("duplicate create_pod should be absorbed, got error: %v", err)
	}

	if got := len(store.All()); got != 1 {
		t.Fatalf("pod must not be rebuilt; store has %d pods", got)
	}
	events := mockConn.GetEvents()
	hasCreated := false
	for _, e := range events {
		if e.Type == client.MsgTypePodCreated {
			hasCreated = true
		}
	}
	if !hasCreated {
		t.Error("absorbed duplicate for running pod should re-send pod_created ack")
	}
}

func TestCleanupPodExit_ClearsPromptDedup(t *testing.T) {
	store := NewInMemoryPodStore()
	mockConn := client.NewMockConnection()
	r := &Runner{cfg: &config.Config{}}
	h := NewRunnerMessageHandler(r, store, mockConn)

	base := &sendPromptMockIO{mode: InteractionModePTY}
	io := &ptyTerminalMock{sendPromptMockIO: base}
	pod := &Pod{PodKey: "gone-pod", InteractionMode: InteractionModePTY, IO: io, Status: PodStatusRunning}
	store.Put("gone-pod", pod)

	cmd := &runnerv1.SendPromptCommand{PodKey: "gone-pod", Prompt: "hi", CommandId: "cmd-z"}
	if err := h.OnSendPrompt(cmd); err != nil {
		t.Fatalf("OnSendPrompt error: %v", err)
	}

	h.cleanupPodExit("gone-pod", 0, false)

	h.promptDedupMu.Lock()
	_, exists := h.promptDedup["gone-pod"]
	h.promptDedupMu.Unlock()
	if exists {
		t.Error("prompt dedup entry should be cleared on pod exit")
	}
}
