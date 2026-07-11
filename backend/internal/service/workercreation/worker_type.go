package workercreation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	agentdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	agentservice "github.com/anthropics/agentsmesh/backend/internal/service/agent"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

const workerTypeSchemaVersion uint32 = 1

type AgentProvider interface {
	GetAgent(context.Context, string) (*agentdomain.Agent, error)
	ListBuiltinAgents(context.Context) ([]*agentdomain.Agent, error)
}

type workerTypeResolver struct {
	agents AgentProvider
}

func newWorkerTypeResolver(agents AgentProvider) *workerTypeResolver {
	return &workerTypeResolver{agents: agents}
}

func (resolver *workerTypeResolver) ResolveWorkerType(
	ctx context.Context,
	_ specservice.Scope,
	slug slugkit.Slug,
) (specservice.WorkerTypeResolution, error) {
	if resolver == nil || resolver.agents == nil {
		return specservice.WorkerTypeResolution{}, specservice.ErrResolverUnavailable
	}
	agent, err := resolver.agents.GetAgent(ctx, slug.String())
	if err != nil {
		if errors.Is(err, agentservice.ErrAgentNotFound) {
			return specservice.WorkerTypeResolution{}, invalidWorkerType("worker type does not exist")
		}
		return specservice.WorkerTypeResolution{}, err
	}
	if agent == nil {
		return specservice.WorkerTypeResolution{}, invalidWorkerType("worker type does not exist")
	}
	if agent.Slug != slug.String() {
		return specservice.WorkerTypeResolution{}, invalidWorkerType(
			fmt.Sprintf("resolved slug %q does not match %q", agent.Slug, slug),
		)
	}
	if !agent.IsActive {
		return specservice.WorkerTypeResolution{}, invalidWorkerType("worker type is disabled")
	}
	if agent.IsInternal {
		return specservice.WorkerTypeResolution{}, invalidWorkerType("internal worker type is not selectable")
	}
	if agent.AgentfileSource == nil || *agent.AgentfileSource == "" {
		return specservice.WorkerTypeResolution{}, invalidWorkerType("worker type has no AgentFile definition")
	}
	schema, err := agentservice.ResolveConfigSchema(ctx, resolver.agents, slug.String())
	if err != nil {
		return specservice.WorkerTypeResolution{}, invalidWorkerType(
			fmt.Sprintf("invalid AgentFile schema: %v", err),
		)
	}
	typeSchema, err := convertTypeSchema(agent.Slug, schema)
	if err != nil {
		return specservice.WorkerTypeResolution{}, err
	}
	supportedModes, err := parseSupportedInteractionModes(agent.SupportedModes)
	if err != nil {
		return specservice.WorkerTypeResolution{}, err
	}
	hash, err := definitionHash(agent, typeSchema)
	if err != nil {
		return specservice.WorkerTypeResolution{}, err
	}
	return specservice.WorkerTypeResolution{
		WorkerType: specdomain.WorkerType{
			Slug:           slug,
			DefinitionHash: hash,
		},
		TypeSchema:                typeSchema,
		SupportedInteractionModes: supportedModes,
	}, nil
}

func parseSupportedInteractionModes(value string) ([]specdomain.InteractionMode, error) {
	modes := make([]specdomain.InteractionMode, 0, 2)
	seen := map[specdomain.InteractionMode]struct{}{}
	for _, rawMode := range strings.Split(value, ",") {
		mode := specdomain.InteractionMode(strings.TrimSpace(rawMode))
		switch mode {
		case specdomain.InteractionModePTY, specdomain.InteractionModeACP:
		default:
			return nil, invalidWorkerType(
				fmt.Sprintf("unsupported interaction mode %q in definition", mode),
			)
		}
		if _, exists := seen[mode]; exists {
			continue
		}
		seen[mode] = struct{}{}
		modes = append(modes, mode)
	}
	if len(modes) == 0 {
		return nil, invalidWorkerType("definition has no supported interaction modes")
	}
	return modes, nil
}

func convertTypeSchema(
	workerType string,
	schema *agentservice.ConfigSchemaResponse,
) (specdomain.TypeSchema, error) {
	if schema == nil {
		return specdomain.TypeSchema{}, invalidWorkerType("worker type schema is missing")
	}
	fields := make(map[string]specdomain.TypeFieldSchema, len(schema.Fields)+len(schema.CredentialFields))
	for _, field := range schema.Fields {
		if isModelResourceManagedTypeField(workerType, field.Name) {
			continue
		}
		kind, err := typeFieldKind(field.Type)
		if err != nil {
			return specdomain.TypeSchema{}, invalidWorkerType(err.Error())
		}
		options := make([]string, len(field.Options))
		for index, option := range field.Options {
			options[index] = option.Value
		}
		fields[field.Name] = specdomain.TypeFieldSchema{Kind: kind, Options: options}
	}
	for _, field := range schema.CredentialFields {
		if isModelResourceManagedTypeField(workerType, field.Name) {
			continue
		}
		if _, exists := fields[field.Name]; exists {
			return specdomain.TypeSchema{}, invalidWorkerType(
				fmt.Sprintf("duplicate config field %q", field.Name),
			)
		}
		fields[field.Name] = specdomain.TypeFieldSchema{Kind: specdomain.TypeFieldSecret}
	}
	return specdomain.TypeSchema{Version: workerTypeSchemaVersion, Fields: fields}, nil
}

func typeFieldKind(value string) (specdomain.TypeFieldKind, error) {
	switch value {
	case "boolean":
		return specdomain.TypeFieldBoolean, nil
	case "string":
		return specdomain.TypeFieldString, nil
	case "number":
		return specdomain.TypeFieldNumber, nil
	case "select":
		return specdomain.TypeFieldSelect, nil
	default:
		return "", fmt.Errorf("unsupported config field type %q", value)
	}
}

func definitionHash(agent *agentdomain.Agent, schema specdomain.TypeSchema) (string, error) {
	document := struct {
		Slug           string                `json:"slug"`
		Executable     string                `json:"executable"`
		SupportedModes string                `json:"supported_modes"`
		Agentfile      string                `json:"agentfile"`
		Schema         specdomain.TypeSchema `json:"schema"`
	}{
		Slug:           agent.Slug,
		Executable:     agent.Executable,
		SupportedModes: agent.SupportedModes,
		Agentfile:      *agent.AgentfileSource,
		Schema:         schema,
	}
	data, err := json.Marshal(document)
	if err != nil {
		return "", fmt.Errorf("encode worker type definition: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func invalidWorkerType(reason string) error {
	return fmt.Errorf("%w: worker type: %s", specservice.ErrInvalidDraft, reason)
}
