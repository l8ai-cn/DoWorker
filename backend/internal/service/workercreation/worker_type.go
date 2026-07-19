package workercreation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	agentdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	agentservice "github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

const workerTypeSchemaVersion uint32 = 1

type AgentProvider interface {
	GetAgent(context.Context, string) (*agentdomain.Agent, error)
	ListBuiltinAgents(context.Context) ([]*agentdomain.Agent, error)
}

type workerTypeResolver struct {
	agents      AgentProvider
	definitions WorkerDefinitionProvider
}

func newWorkerTypeResolver(
	agents AgentProvider,
	definitions WorkerDefinitionProvider,
) *workerTypeResolver {
	return &workerTypeResolver{agents: agents, definitions: definitions}
}

func (resolver *workerTypeResolver) ResolveWorkerType(
	ctx context.Context,
	_ specservice.Scope,
	slug slugkit.Slug,
) (specservice.WorkerTypeResolution, error) {
	if resolver == nil || resolver.agents == nil || resolver.definitions == nil {
		return specservice.WorkerTypeResolution{}, specservice.ErrResolverUnavailable
	}
	definition, ok := resolver.definitions.Get(slug.String())
	if !ok {
		return specservice.WorkerTypeResolution{}, invalidWorkerType("missing canonical definition")
	}
	agent, err := resolver.agents.GetAgent(ctx, slug.String())
	if err != nil {
		if errors.Is(err, agentservice.ErrAgentNotFound) {
			return specservice.WorkerTypeResolution{}, invalidWorkerType("worker type does not exist")
		}
		return specservice.WorkerTypeResolution{}, err
	}
	if err := validateWorkerTypeProjection(agent, slug, definition); err != nil {
		return specservice.WorkerTypeResolution{}, err
	}
	typeSchema, err := typeSchemaFromDefinition(definition)
	if err != nil {
		return specservice.WorkerTypeResolution{}, err
	}
	modes, err := parseSupportedInteractionModes(definition.Modes)
	if err != nil {
		return specservice.WorkerTypeResolution{}, err
	}
	modelRequirement, err := modelRequirementFromDefinition(slug, definition)
	if err != nil {
		return specservice.WorkerTypeResolution{}, err
	}
	toolModelRequirements, err := toolModelRequirementsFromDefinition(definition)
	if err != nil {
		return specservice.WorkerTypeResolution{}, err
	}
	return specservice.WorkerTypeResolution{
		WorkerType: specdomain.WorkerType{
			Slug:           slug,
			DefinitionHash: definition.DefinitionHash,
		},
		TypeSchema:                typeSchema,
		SupportedInteractionModes: modes,
		ModelRequirement:          modelRequirement,
		ToolModelRequirements:     toolModelRequirements,
	}, nil
}

func modelRequirementFromDefinition(
	workerType slugkit.Slug,
	definition workerdefinition.Definition,
) (specdomain.ModelRequirement, error) {
	adapters := make(
		[]slugkit.Slug,
		len(definition.ModelRequirement.ProtocolAdapters),
	)
	for index, rawAdapter := range definition.ModelRequirement.ProtocolAdapters {
		adapter, err := slugkit.NewFromTrusted(rawAdapter)
		if err != nil {
			return specdomain.ModelRequirement{}, invalidWorkerType(
				fmt.Sprintf("invalid model protocol adapter %q", rawAdapter),
			)
		}
		adapters[index] = adapter
	}
	requirement := specdomain.ModelRequirement{
		Required:         definition.ModelRequirement.Required,
		ProtocolAdapters: adapters,
	}
	if err := specservice.ValidateModelRequirement(requirement); err != nil {
		return specdomain.ModelRequirement{}, invalidWorkerType(err.Error())
	}
	if err := validateWorkerModelRequirement(workerType, requirement); err != nil {
		return specdomain.ModelRequirement{}, err
	}
	return requirement, nil
}

func toolModelRequirementsFromDefinition(
	definition workerdefinition.Definition,
) ([]specdomain.ToolModelRequirement, error) {
	requirements := make(
		[]specdomain.ToolModelRequirement,
		0,
		len(definition.ToolModelRequirements),
	)
	for _, item := range definition.ToolModelRequirements {
		role, err := slugkit.NewFromTrusted(item.ID)
		if err != nil {
			return nil, invalidWorkerType("invalid tool model role")
		}
		providers, err := trustedSlugs(item.ProviderKeys)
		if err != nil {
			return nil, invalidWorkerType("invalid tool model provider")
		}
		adapters, err := trustedSlugs(item.ProtocolAdapters)
		if err != nil {
			return nil, invalidWorkerType("invalid tool model protocol adapter")
		}
		requirements = append(requirements, specdomain.ToolModelRequirement{
			Role:             role,
			ProviderKeys:     providers,
			ProtocolAdapters: adapters,
			Modality:         resourcedomain.Modality(item.Modality),
			Capability:       resourcedomain.Capability(item.Capability),
			Environment: specdomain.ToolModelEnvironment{
				APIKey: item.Environment.APIKey, BaseURL: item.Environment.BaseURL,
				ModelID: item.Environment.ModelID,
			},
		})
	}
	if err := specservice.ValidateToolModelRequirements(requirements); err != nil {
		return nil, invalidWorkerType(err.Error())
	}
	return requirements, nil
}

func trustedSlugs(values []string) ([]slugkit.Slug, error) {
	slugs := make([]slugkit.Slug, len(values))
	for index, value := range values {
		slug, err := slugkit.NewFromTrusted(value)
		if err != nil {
			return nil, err
		}
		slugs[index] = slug
	}
	return slugs, nil
}

func parseSupportedInteractionModes(value []string) ([]specdomain.InteractionMode, error) {
	modes := make([]specdomain.InteractionMode, 0, len(value))
	seen := map[specdomain.InteractionMode]struct{}{}
	for _, rawMode := range value {
		mode := specdomain.InteractionMode(strings.TrimSpace(rawMode))
		switch mode {
		case specdomain.InteractionModePTY, specdomain.InteractionModeACP:
		default:
			return nil, invalidWorkerType(
				fmt.Sprintf("unsupported interaction mode %q in definition", mode),
			)
		}
		if _, exists := seen[mode]; exists {
			return nil, invalidWorkerType(
				fmt.Sprintf("duplicate interaction mode %q in definition", mode),
			)
		}
		seen[mode] = struct{}{}
		modes = append(modes, mode)
	}
	if len(modes) == 0 {
		return nil, invalidWorkerType("definition has no supported interaction modes")
	}
	return modes, nil
}

func invalidWorkerType(reason string) error {
	return fmt.Errorf("%w: worker type: %s", specservice.ErrInvalidDraft, reason)
}
