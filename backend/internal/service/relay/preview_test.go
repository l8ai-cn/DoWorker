package relay

import (
	"errors"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
)

func TestResolvePreviewRoute(t *testing.T) {
	pod := &agentpod.Pod{RunnerID: 7, PreviewPort: 3000, PreviewPath: "/app/", Status: agentpod.StatusRunning}
	r, err := ResolvePreviewRoute(pod)
	if err != nil {
		t.Fatal(err)
	}
	if r.Target != "127.0.0.1:3000" || r.RunnerID != 7 || r.Path != "/app" {
		t.Fatalf("bad route %+v", r)
	}

	if _, err := ResolvePreviewRoute(&agentpod.Pod{RunnerID: 7, PreviewPort: 0, Status: agentpod.StatusRunning}); !errors.Is(err, ErrPreviewDisabled) {
		t.Fatalf("port 0 must error preview_disabled, got %v", err)
	}
	if _, err := ResolvePreviewRoute(&agentpod.Pod{RunnerID: 7, PreviewPort: -1, Status: agentpod.StatusRunning}); !errors.Is(err, ErrInvalidPreviewPort) {
		t.Fatalf("negative port must error invalid_preview_port, got %v", err)
	}

	if _, err := ResolvePreviewRoute(&agentpod.Pod{RunnerID: 7, PreviewPort: 3000, Status: agentpod.StatusCompleted}); !errors.Is(err, ErrPodNotActive) {
		t.Fatalf("inactive pod must error pod_not_active, got %v", err)
	}

	if _, err := ResolvePreviewRoute(nil); !errors.Is(err, ErrPodNotActive) {
		t.Fatalf("nil pod must error pod_not_active, got %v", err)
	}
}

func TestNormalizePreviewPath(t *testing.T) {
	t.Parallel()

	valid := map[string]string{
		"":                          "/",
		"/":                         "/",
		"/app/":                     "/app",
		"/app//api":                 "/app/api",
		"/app/./api":                "/app/api",
		"/files/%25":                "/files/%25",
		"/files/report%23draft.pdf": "/files/report%23draft.pdf",
		"/route/%3F":                "/route/%3F",
		"/app/%252e%252e":           "/app/%252e%252e",
		"/documents/%E4%B8%AD":      "/documents/%E4%B8%AD",
	}
	for input, want := range valid {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			got, err := NormalizePreviewPath(input)
			if err != nil {
				t.Fatalf("NormalizePreviewPath(%q): %v", input, err)
			}
			if got != want {
				t.Fatalf("NormalizePreviewPath(%q) = %q, want %q", input, got, want)
			}
			again, err := NormalizePreviewPath(got)
			if err != nil {
				t.Fatalf("NormalizePreviewPath(%q) second pass: %v", got, err)
			}
			if again != got {
				t.Fatalf("NormalizePreviewPath is not idempotent: first %q, second %q", got, again)
			}
		})
	}

	for _, input := range []string{
		"app",
		"/app/../admin",
		"/app/%2e%2e/admin",
		"/app?debug=true",
		"/app#fragment",
		"/bad%2",
	} {
		input := input
		t.Run("reject_"+input, func(t *testing.T) {
			if _, err := NormalizePreviewPath(input); !errors.Is(err, ErrInvalidPreviewPath) {
				t.Fatalf("NormalizePreviewPath(%q) error = %v, want ErrInvalidPreviewPath", input, err)
			}
		})
	}
}

func TestNormalizePreviewConfigValidatesPortRange(t *testing.T) {
	t.Parallel()

	for _, port := range []int{0, 1024, 65535} {
		path, err := NormalizePreviewConfig(port, "")
		if err != nil {
			t.Fatalf("NormalizePreviewConfig(%d): %v", port, err)
		}
		if path != "/" {
			t.Fatalf("NormalizePreviewConfig(%d) path = %q, want /", port, path)
		}
	}
	for _, port := range []int{-1, 1, 1023, 65536} {
		if _, err := NormalizePreviewConfig(port, "/"); !errors.Is(err, ErrInvalidPreviewPort) {
			t.Fatalf("NormalizePreviewConfig(%d) error = %v, want ErrInvalidPreviewPort", port, err)
		}
	}
}

func TestResolvePreviewRouteRejectsInvalidPath(t *testing.T) {
	pod := &agentpod.Pod{
		RunnerID:    7,
		PreviewPort: 3000,
		PreviewPath: "/app/%2e%2e/admin",
		Status:      agentpod.StatusRunning,
	}
	if _, err := ResolvePreviewRoute(pod); !errors.Is(err, ErrInvalidPreviewPath) {
		t.Fatalf("expected ErrInvalidPreviewPath, got %v", err)
	}
}
