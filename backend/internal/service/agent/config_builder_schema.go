package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/agentsmesh/agentfile/schema"
)

// ResolveConfigSchema returns the config + credential schema for an agent.
// Both field sets are extracted from AgentFile declarations (CONFIG and
// ENV SECRET/TEXT). Credential UX grouping (oneof auth methods, labels) stays
// in the frontend override registry.
func ResolveConfigSchema(
	ctx context.Context,
	provider AgentConfigProvider,
	credentialSources CredentialFieldSourceProvider,
	agentSlug string,
) (*ConfigSchemaResponse, error) {
	agentDef, err := provider.GetAgent(ctx, agentSlug)
	if err != nil {
		return nil, err
	}
	if agentDef.AgentfileSource == nil || *agentDef.AgentfileSource == "" {
		return &ConfigSchemaResponse{Fields: []ConfigFieldResponse{}}, nil
	}
	schema, err := ConfigSchemaFromAgentfile(*agentDef.AgentfileSource)
	if err != nil {
		return nil, err
	}
	if credentialSources == nil {
		return nil, fmt.Errorf("credential field source is required")
	}
	fields, isFormalWorker := credentialSources.CredentialBundleFields(agentSlug)
	if !isFormalWorker {
		return schema, nil
	}
	schema.CredentialFields = filterCredentialFields(schema.CredentialFields, fields)
	return schema, nil
}

func ConfigSchemaFromAgentfile(source string) (*ConfigSchemaResponse, error) {
	agentSchema, err := schema.FromSource(source)
	if err != nil {
		return nil, err
	}

	result := &ConfigSchemaResponse{
		Fields:           make([]ConfigFieldResponse, 0, len(agentSchema.ConfigFields)),
		CredentialFields: make([]CredentialFieldResponse, 0, len(agentSchema.CredentialFields)),
		ConfigFiles:      make([]ConfigFileResponse, 0, len(agentSchema.ConfigFiles)),
	}
	for _, cfg := range agentSchema.ConfigFields {
		field := ConfigFieldResponse{
			Name:    cfg.Name,
			Type:    cfg.Type,
			Default: cfg.Default,
		}
		if len(cfg.Options) > 0 {
			field.Options = make([]FieldOptionResponse, 0, len(cfg.Options))
			for _, opt := range cfg.Options {
				field.Options = append(field.Options, FieldOptionResponse{Value: opt})
			}
		}
		result.Fields = append(result.Fields, field)
	}
	for _, cred := range agentSchema.CredentialFields {
		result.CredentialFields = append(result.CredentialFields, CredentialFieldResponse{
			Name:     cred.Name,
			Type:     cred.Type,
			Optional: cred.Optional,
		})
	}
	for _, cf := range agentSchema.ConfigFiles {
		result.ConfigFiles = append(result.ConfigFiles, ConfigFileResponse{
			ID:       cf.ID,
			PathEnv:  cf.PathEnv,
			Format:   cf.Format,
			PathHint: cf.PathHint,
		})
	}
	return result, nil
}

func filterCredentialFields(
	fields []CredentialFieldResponse,
	allowed []string,
) []CredentialFieldResponse {
	allowedFields := make(map[string]struct{}, len(allowed))
	for _, field := range allowed {
		allowedFields[field] = struct{}{}
	}
	filtered := make([]CredentialFieldResponse, 0, len(allowed))
	for _, field := range fields {
		if _, ok := allowedFields[field.Name]; ok {
			filtered = append(filtered, field)
		}
	}
	return filtered
}
