package runner

import "testing"

func TestPodStatusBroadcaster_MultipleListeners(t *testing.T) {
	b := newPodStatusBroadcaster()
	var calls int
	b.add(func(podKey, status, agentStatus string) { calls++ })
	b.add(func(podKey, status, agentStatus string) { calls++ })
	b.notify("pod-1", "running", "executing")
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestPodStatusBroadcaster_SetReplaces(t *testing.T) {
	b := newPodStatusBroadcaster()
	b.add(func(podKey, status, agentStatus string) {})
	b.set(func(podKey, status, agentStatus string) {})
	if b.len() != 1 {
		t.Fatalf("len = %d, want 1 after set", b.len())
	}
}
