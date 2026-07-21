package orchestrationworker

import (
	"context"
	"fmt"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
)

type WorkerTemplateApplyService struct {
	registry   *resource.Registry
	repository WorkerTemplateApplyRepository
}

func NewWorkerTemplateApplyService(
	registry *resource.Registry,
	repository WorkerTemplateApplyRepository,
) (*WorkerTemplateApplyService, error) {
	if registry == nil || repository == nil ||
		!registry.Has(resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindWorkerTemplate,
		}) {
		return nil, fmt.Errorf(
			"%w: worker template apply dependencies are incomplete",
			controlservice.ErrUnavailable,
		)
	}
	return &WorkerTemplateApplyService{
		registry: registry, repository: repository,
	}, nil
}

func (service *WorkerTemplateApplyService) Apply(
	ctx context.Context,
	scope control.Scope,
	planID string,
) (AppliedWorkerTemplate, error) {
	if service == nil || service.registry == nil || service.repository == nil {
		return AppliedWorkerTemplate{}, controlservice.ErrUnavailable
	}
	return service.repository.RunWorkerTemplateApplyTransaction(
		ctx,
		scope,
		planID,
		func(
			state controlservice.LockedApplyState,
			snapshotID int64,
		) (controlservice.ApplyMutation, error) {
			return buildWorkerTemplateApplyMutation(
				service.registry,
				state,
				snapshotID,
			)
		},
	)
}
