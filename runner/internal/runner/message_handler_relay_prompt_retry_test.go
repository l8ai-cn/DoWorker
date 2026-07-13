package runner

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

func TestHandleAcpRelayCommand_PromptWaitsForACPReady(t *testing.T) {
	h := newTestHandler()
	mock := &mockPodIO{sendErrors: []error{acp.ErrPromptNotReady, nil}}
	pod := &Pod{PodKey: "test-pod", IO: mock}

	payload, _ := json.Marshal(map[string]any{
		"type":   "prompt",
		"prompt": "hello world",
	})
	started := time.Now()
	h.handleAcpRelayCommand(pod, payload)

	mock.mu.Lock()
	defer mock.mu.Unlock()
	if len(mock.inputs) != 2 {
		t.Fatalf("inputs = %v, want retry after ACP becomes ready", mock.inputs)
	}
	if time.Since(started) < 100*time.Millisecond {
		t.Fatal("relay prompt did not wait for ACP readiness")
	}
}
