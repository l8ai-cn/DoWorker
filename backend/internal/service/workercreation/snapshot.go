package workercreation

import (
	"context"

	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
)

func (service *Service) PrepareSnapshot(
	ctx context.Context,
	scope specservice.Scope,
	snapshot specdomain.Snapshot,
) (PreparedSnapshot, error) {
	if service == nil || service.workerTypes == nil || scope.OrgID <= 0 ||
		scope.UserID <= 0 || snapshot.ID <= 0 ||
		snapshot.OrganizationID != scope.OrgID {
		return PreparedSnapshot{}, specservice.ErrResolverUnavailable
	}
	spec, err := specdomain.NormalizeAndValidate(snapshot.Spec)
	if err != nil {
		return PreparedSnapshot{}, err
	}
	workspace := newWorkspaceResolver(service.workspaceDeps)
	layer, err := newCompiler(workspace).Compile(ctx, scope, spec)
	if err != nil {
		return PreparedSnapshot{}, err
	}
	return PreparedSnapshot{
		Spec:           spec,
		AgentfileLayer: layer,
		Repository:     workspace.resolvedRepository(spec.Workspace.RepositoryID),
	}, nil
}
