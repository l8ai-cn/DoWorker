package relay

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

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
	// ErrInvalidPreviewPort indicates preview is neither disabled nor bound to
	// a non-privileged TCP port.
	ErrInvalidPreviewPort = errors.New("invalid_preview_port")
	// ErrInvalidPreviewPath indicates preview metadata cannot be routed safely.
	ErrInvalidPreviewPath = errors.New("invalid_preview_path")
)

func NormalizePreviewConfig(port int, rawPath string) (string, error) {
	if port != 0 && (port < 1024 || port > 65535) {
		return "", ErrInvalidPreviewPort
	}
	return NormalizePreviewPath(rawPath)
}

func NormalizePreviewPath(raw string) (string, error) {
	if raw == "" {
		return "/", nil
	}
	if strings.ContainsAny(raw, "?#") {
		return "", ErrInvalidPreviewPath
	}
	decoded, err := url.PathUnescape(raw)
	if err != nil || !strings.HasPrefix(decoded, "/") {
		return "", ErrInvalidPreviewPath
	}
	for _, segment := range strings.Split(decoded, "/") {
		if segment == ".." {
			return "", ErrInvalidPreviewPath
		}
	}
	cleaned := path.Clean(decoded)
	canonical := (&url.URL{Path: cleaned}).EscapedPath()
	if len(canonical) > 255 {
		return "", ErrInvalidPreviewPath
	}
	return canonical, nil
}

// ResolvePreviewRoute derives the preview routing target from pod metadata
// without any table lookup at request time (routing later relies on the JWT
// claim). It enforces that the pod is active, has a runner, and has preview on.
func ResolvePreviewRoute(pod *agentpod.Pod) (PreviewRoute, error) {
	if pod == nil || !pod.IsActive() || pod.RunnerID == 0 {
		return PreviewRoute{}, ErrPodNotActive
	}
	if pod.PreviewPort == 0 {
		return PreviewRoute{}, ErrPreviewDisabled
	}
	previewPath, err := NormalizePreviewConfig(pod.PreviewPort, pod.PreviewPath)
	if err != nil {
		return PreviewRoute{}, err
	}
	return PreviewRoute{
		RunnerID: pod.RunnerID,
		Target:   fmt.Sprintf("127.0.0.1:%d", pod.PreviewPort),
		Path:     previewPath,
	}, nil
}
