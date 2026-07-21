package workerspec

import (
	"context"
	"fmt"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

type Service struct {
	resolver   *Resolver
	repository SnapshotRepository
}

func NewService(resolver *Resolver, repository SnapshotRepository) *Service {
	return &Service{resolver: resolver, repository: repository}
}

func (service *Service) CreateSnapshot(
	ctx context.Context,
	scope Scope,
	draft Draft,
) (domain.Snapshot, error) {
	if service == nil || service.resolver == nil {
		return domain.Snapshot{}, ErrResolverUnavailable
	}
	if service.repository == nil {
		return domain.Snapshot{}, ErrSnapshotRepositoryUnavailable
	}
	resolved, err := service.resolver.Resolve(ctx, scope, draft)
	if err != nil {
		return domain.Snapshot{}, err
	}
	snapshot, err := service.repository.Create(ctx, resolved)
	if err != nil {
		return domain.Snapshot{}, fmt.Errorf("create workerspec snapshot: %w", err)
	}
	return snapshot, nil
}
