package agentpod

import (
	"context"
	"errors"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	relayservice "github.com/l8ai-cn/agentcloud/backend/internal/service/relay"
	"gorm.io/gorm"
)

func (s *PodService) UpdatePreviewConfig(
	ctx context.Context,
	podKey string,
	createdByID int64,
	previewPort int,
	previewPath string,
) (*agentpod.Pod, error) {
	normalizedPath, err := relayservice.NormalizePreviewConfig(previewPort, previewPath)
	if err != nil {
		return nil, err
	}
	pod, err := s.repo.UpdatePreviewConfig(ctx, podKey, createdByID, previewPort, normalizedPath)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPodNotFound
	}
	return pod, err
}
