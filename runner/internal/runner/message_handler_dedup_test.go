package runner

import (
	"errors"
	"testing"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/config"
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

func TestOnSendPrompt_FailedCommandIDCanRetry(t *testing.T) {
	base := &sendPromptMockIO{mode: InteractionModePTY, inputErr: errors.New("write failed")}
	io := &ptyTerminalMock{sendPromptMockIO: base}
	pod := &Pod{PodKey: "retry-pod", InteractionMode: InteractionModePTY, IO: io}

	store := NewInMemoryPodStore()
	store.Put(pod.PodKey, pod)
	h := &RunnerMessageHandler{podStore: store}
	cmd := &runnerv1.SendPromptCommand{
		PodKey: pod.PodKey, Prompt: "retry me", CommandId: "cmd-retry",
	}

	if err := h.OnSendPrompt(cmd); err == nil {
		t.Fatal("first OnSendPrompt should fail")
	}
	base.mu.Lock()
	base.inputErr = nil
	base.mu.Unlock()
	if err := h.OnSendPrompt(cmd); err != nil {
		t.Fatalf("retry OnSendPrompt error: %v", err)
	}
	if err := h.OnSendPrompt(cmd); err != nil {
		t.Fatalf("duplicate successful retry error: %v", err)
	}

	base.mu.Lock()
	defer base.mu.Unlock()
	if len(base.inputs) != 2 {
		t.Fatalf("failed command_id must remain retryable; got %d writes", len(base.inputs))
	}
	if len(base.keys) != 1 || base.keys[0].payload != "enter" {
		t.Fatalf("successful retry must submit once; got keys %v", base.keys)
	}
}

func TestOnSendPrompt_EnterFailureDoesNotResendBody(t *testing.T) {
	base := &sendPromptMockIO{mode: InteractionModePTY, keysErr: errors.New("enter failed")}
	io := &ptyTerminalMock{sendPromptMockIO: base}
	pod := &Pod{PodKey: "partial-pod", InteractionMode: InteractionModePTY, IO: io}

	store := NewInMemoryPodStore()
	store.Put(pod.PodKey, pod)
	h := &RunnerMessageHandler{podStore: store}
	cmd := &runnerv1.SendPromptCommand{
		PodKey: pod.PodKey, Prompt: "run once", CommandId: "cmd-partial",
	}

	requirePromptError(t, h.OnSendPrompt(cmd))
	if err := h.OnSendPrompt(cmd); err != nil {
		t.Fatalf("uncertain duplicate should be absorbed: %v", err)
	}

	base.mu.Lock()
	defer base.mu.Unlock()
	if len(base.inputs) != 1 {
		t.Fatalf("body must not be resent after Enter failure; got %d writes", len(base.inputs))
	}
	if len(base.keys) != 1 {
		t.Fatalf("Enter must only be attempted once; got %d attempts", len(base.keys))
	}
}

func TestOnSendPrompt_DedupSurvivesHandlerRestart(t *testing.T) {
	receiptRoot := t.TempDir()
	firstIO := &ptyTerminalMock{sendPromptMockIO: &sendPromptMockIO{mode: InteractionModePTY}}
	firstStore := NewInMemoryPodStore()
	firstStore.Put("durable-pod", &Pod{
		PodKey: "durable-pod", InteractionMode: InteractionModePTY, IO: firstIO,
	})
	first := &RunnerMessageHandler{
		podStore: firstStore,
		receipts: newCommandReceiptStore(receiptRoot),
	}
	command := &runnerv1.SendPromptCommand{
		PodKey: "durable-pod", Prompt: "run once", CommandId: "cmd-durable",
	}
	if err := first.OnSendPrompt(command); err != nil {
		t.Fatalf("first OnSendPrompt error: %v", err)
	}

	secondIO := &ptyTerminalMock{sendPromptMockIO: &sendPromptMockIO{mode: InteractionModePTY}}
	secondStore := NewInMemoryPodStore()
	secondStore.Put("durable-pod", &Pod{
		PodKey: "durable-pod", InteractionMode: InteractionModePTY, IO: secondIO,
	})
	second := &RunnerMessageHandler{
		podStore: secondStore,
		receipts: newCommandReceiptStore(receiptRoot),
	}

	if err := second.OnSendPrompt(command); err != nil {
		t.Fatalf("replayed OnSendPrompt error: %v", err)
	}
	secondIO.mu.Lock()
	defer secondIO.mu.Unlock()
	if len(secondIO.inputs) != 0 {
		t.Fatalf("durable duplicate must not rewrite prompt; got %d writes", len(secondIO.inputs))
	}
}

func requirePromptError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("OnSendPrompt should fail")
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
