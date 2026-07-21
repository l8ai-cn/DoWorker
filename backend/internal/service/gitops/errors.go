package gitops

import (
	"errors"

	"github.com/l8ai-cn/agentcloud/backend/internal/infra/gitea"
)

var (
	// ErrNotConfigured mirrors the gitea sentinel so consumers of gitops
	// never need to import internal/infra/gitea directly.
	ErrNotConfigured = gitea.ErrNotConfigured

	// ErrNotFound is returned by content methods (ReadFile/ListDir/ListTree)
	// when the underlying Gitea call responds with a 404, so callers can
	// branch without string-matching transport errors.
	ErrNotFound = errors.New("gitops: not found")
)
