package relay

import (
	"errors"
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

// PreviewRoute is the resolved routing target for a pod preview request. Target
// is always a pod-local loopback address; the runner re-validates it.
type PreviewRoute struct {
	RunnerID int64
	Target   string
	Path     string
}

var (
	// ErrPreviewDisabled indicates the pod has no preview port configured.
	ErrPreviewDisabled = errors.New("preview_disabled")
	// ErrPodNotActive indicates the pod is not in a routable/active state.
	ErrPodNotActive = errors.New("pod_not_active")
)

// ResolvePreviewRoute derives the preview routing target from pod metadata
// without any table lookup at request time (routing later relies on the JWT
// claim). It enforces that the pod is active, has a runner, and has preview on.
func ResolvePreviewRoute(pod *agentpod.Pod) (PreviewRoute, error) {
	if pod == nil || !pod.IsActive() || pod.RunnerID == 0 {
		return PreviewRoute{}, ErrPodNotActive
	}
	if pod.PreviewPort <= 0 {
		return PreviewRoute{}, ErrPreviewDisabled
	}
	return PreviewRoute{
		RunnerID: pod.RunnerID,
		Target:   fmt.Sprintf("127.0.0.1:%d", pod.PreviewPort),
		Path:     pod.PreviewPath,
	}, nil
}
