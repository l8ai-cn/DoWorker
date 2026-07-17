package infra

import (
	"context"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

func (repo *orchestrationResourceRepo) RunPromptApplyTransaction(
	ctx context.Context,
	scope control.Scope,
	planID string,
	build controlservice.ApplyBuilder,
) (control.ResourceHead, error) {
	return repo.runIdempotentApplyTransaction(
		ctx,
		scope,
		planID,
		resource.KindPrompt,
		"PromptSpec",
		build,
	)
}
