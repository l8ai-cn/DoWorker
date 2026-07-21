package workercreation

import (
	"context"
	"errors"
	"fmt"
	"os"

	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func (service *Service) listWorkerTypeOptions(
	ctx context.Context,
	scope specservice.Scope,
	filter string,
) ([]WorkerTypeOption, error) {
	agents, err := service.agents.ListBuiltinAgents(ctx)
	if err != nil {
		return nil, err
	}
	options := make([]WorkerTypeOption, 0, len(agents))
	includeInternal := os.Getenv("AGENTCLOUD_INCLUDE_INTERNAL_AGENTS") == "true"
	for _, agent := range agents {
		if agent == nil || !agent.IsActive ||
			(agent.IsInternal && !includeInternal) ||
			(filter != "" && agent.Slug != filter) {
			continue
		}
		slug, err := slugkit.NewFromTrusted(agent.Slug)
		if err != nil {
			return nil, fmt.Errorf("invalid worker type slug %q: %w", agent.Slug, err)
		}
		option := WorkerTypeOption{Slug: agent.Slug, Name: agent.Name}
		if agent.Description != nil {
			option.Description = *agent.Description
		}
		resolved, err := service.workerTypes.ResolveWorkerType(ctx, scope, slug)
		if err != nil {
			if !errors.Is(err, specservice.ErrInvalidDraft) {
				return nil, err
			}
			option.BlockingReason = err.Error()
			options = append(options, option)
			continue
		}
		definition, exists := service.workspaceDeps.Definitions.Get(agent.Slug)
		if !exists {
			return nil, fmt.Errorf("canonical definition for %q does not exist", agent.Slug)
		}
		option.Schema = resolved.TypeSchema
		option.SupportedInteractionModes = append(
			[]specdomain.InteractionMode{},
			resolved.SupportedInteractionModes...,
		)
		option.RequiresModelResource = resolved.ModelRequirement.Required
		option.ModelProtocolAdapters = modelProtocolAdapters(
			resolved.ModelRequirement.ProtocolAdapters,
		)
		option.ToolModelRequirements = append(
			[]specdomain.ToolModelRequirement{},
			resolved.ToolModelRequirements...,
		)
		option.CredentialRequirements, option.ConfigDocumentRequirements =
			workerDefinitionRequirements(definition)
		if !hasEnabledRuntimeImage(service.catalog, agent.Slug) {
			option.BlockingReason = "No runtime image is available for this worker type"
			options = append(options, option)
			continue
		}
		available, err := service.runners.HasAvailableRunnerForAgent(
			ctx, scope.OrgID, scope.UserID, agent.Slug,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"check Runner availability for worker type %q: %w",
				agent.Slug, err,
			)
		}
		if !available {
			option.BlockingReason = "No online Runner currently supports this worker type"
			options = append(options, option)
			continue
		}
		option.Selectable = true
		options = append(options, option)
	}
	return options, nil
}
