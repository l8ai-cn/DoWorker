package workercreation

import (
	"fmt"
	"strings"

	agentdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	agentservice "github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func validateWorkerTypeProjection(
	agent *agentdomain.Agent,
	slug slugkit.Slug,
	definition workerdefinition.Definition,
) error {
	if agent == nil {
		return invalidWorkerType("worker type does not exist")
	}
	if agent.Slug != slug.String() || definition.Slug != slug.String() {
		return invalidWorkerType(fmt.Sprintf("resolved slug does not match %q", slug))
	}
	if !agent.IsActive {
		return invalidWorkerType("worker type is disabled")
	}
	if agent.IsInternal {
		return invalidWorkerType("internal worker type is not selectable")
	}
	if agent.AgentfileSource == nil || *agent.AgentfileSource == "" {
		return invalidWorkerType("worker type has no AgentFile projection")
	}
	if agent.Executable != definition.Executable ||
		agent.AdapterID != definition.AdapterID ||
		*agent.AgentfileSource != definition.AgentFile {
		return invalidWorkerType("database projection does not match canonical definition")
	}
	agentModes, err := parseSupportedInteractionModes(strings.Split(agent.SupportedModes, ","))
	if err != nil {
		return invalidWorkerType("database projection has invalid interaction modes")
	}
	definitionModes, err := parseSupportedInteractionModes(definition.Modes)
	if err != nil {
		return err
	}
	if !sameInteractionModes(agentModes, definitionModes) {
		return invalidWorkerType("database projection does not match canonical interaction modes")
	}
	return nil
}

func typeSchemaFromDefinition(
	definition workerdefinition.Definition,
) (specdomain.TypeSchema, error) {
	schema, err := agentservice.ConfigSchemaFromAgentfile(definition.AgentFile)
	if err != nil {
		return specdomain.TypeSchema{}, invalidWorkerType(
			fmt.Sprintf("invalid canonical AgentFile schema: %v", err),
		)
	}
	return convertTypeSchema(definition, schema)
}

func convertTypeSchema(
	definition workerdefinition.Definition,
	schema *agentservice.ConfigSchemaResponse,
) (specdomain.TypeSchema, error) {
	if schema == nil {
		return specdomain.TypeSchema{}, invalidWorkerType("worker type schema is missing")
	}
	managedFields := modelResourceManagedFields(definition)
	fields := make(map[string]specdomain.TypeFieldSchema, len(schema.Fields)+len(schema.CredentialFields))
	for _, field := range schema.Fields {
		if _, exists := managedFields[field.Name]; exists {
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
		fields[field.Name] = specdomain.TypeFieldSchema{
			Kind:    kind,
			Options: options,
		}
	}
	for _, field := range schema.CredentialFields {
		if _, exists := managedFields[field.Name]; exists {
			continue
		}
		if _, exists := fields[field.Name]; exists {
			return specdomain.TypeSchema{}, invalidWorkerType(
				fmt.Sprintf("duplicate config field %q", field.Name),
			)
		}
		fields[field.Name] = specdomain.TypeFieldSchema{
			Kind: specdomain.TypeFieldSecret,
		}
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

func sameInteractionModes(
	left, right []specdomain.InteractionMode,
) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
