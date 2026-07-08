package relay

import (
	"errors"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

func TestResolvePreviewRoute(t *testing.T) {
	pod := &agentpod.Pod{RunnerID: 7, PreviewPort: 3000, Status: agentpod.StatusRunning}
	r, err := ResolvePreviewRoute(pod)
	if err != nil {
		t.Fatal(err)
	}
	if r.Target != "127.0.0.1:3000" || r.RunnerID != 7 {
		t.Fatalf("bad route %+v", r)
	}

	if _, err := ResolvePreviewRoute(&agentpod.Pod{RunnerID: 7, PreviewPort: 0, Status: agentpod.StatusRunning}); !errors.Is(err, ErrPreviewDisabled) {
		t.Fatalf("port 0 must error preview_disabled, got %v", err)
	}

	if _, err := ResolvePreviewRoute(&agentpod.Pod{RunnerID: 7, PreviewPort: 3000, Status: agentpod.StatusCompleted}); !errors.Is(err, ErrPodNotActive) {
		t.Fatalf("inactive pod must error pod_not_active, got %v", err)
	}

	if _, err := ResolvePreviewRoute(nil); !errors.Is(err, ErrPodNotActive) {
		t.Fatalf("nil pod must error pod_not_active, got %v", err)
	}
}
