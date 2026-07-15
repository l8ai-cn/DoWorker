package agent

import (
	"context"

	"github.com/anthropics/agentsmesh/agentfile/schema"
)

// ResolveConfigSchema returns the config + credential schema for an agent.
// Both field sets are extracted from AgentFile declarations (CONFIG and
// ENV SECRET/TEXT). Credential UX grouping (oneof auth methods, labels) stays
// in the frontend override registry.
func ResolveConfigSchema(ctx context.Context, provider AgentConfigProvider, agentSlug string) (*ConfigSchemaResponse, error) {
	agentDef, err := provider.GetAgent(ctx, agentSlug)
	if err != nil {
		return nil, err
	}
	if agentDef.AgentfileSource == nil || *agentDef.AgentfileSource == "" {
		return &ConfigSchemaResponse{Fields: []ConfigFieldResponse{}}, nil
	}
	return ConfigSchemaFromAgentfile(*agentDef.AgentfileSource)
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
