package workercreation

import (
	"context"
	"errors"
	"fmt"

	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
)

func (service *Service) ValidateWorkerTypeSnapshot(
	ctx context.Context,
	scope specservice.Scope,
	expected specdomain.WorkerType,
) error {
	if service == nil || service.workerTypes == nil {
		return specservice.ErrResolverUnavailable
	}
	current, err := service.workerTypes.ResolveWorkerType(
		ctx,
		scope,
		expected.Slug,
	)
	if err != nil {
		if errors.Is(err, specservice.ErrInvalidDraft) {
			return fmt.Errorf("%w: %q: %v", ErrWorkerTypeDefinitionChanged, expected.Slug, err)
		}
		return err
	}
	if current.WorkerType != expected {
		return fmt.Errorf(
			"%w: %q",
			ErrWorkerTypeDefinitionChanged,
			expected.Slug,
		)
	}
	return nil
}
