package workerdefinition

import (
	"encoding/json"
	"fmt"

	agentfileschema "github.com/anthropics/agentsmesh/agentfile/schema"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type credentialBindingDocument struct {
	ID     string                   `json:"id"`
	Source credentialSourceDocument `json:"source"`
	Target credentialTargetDocument `json:"target"`
}

type credentialSourceDocument struct {
	Kind string `json:"kind"`
	Ref  string `json:"ref"`
}

type credentialTargetDocument struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

func validateCredentialBindings(bindings []json.RawMessage) error {
	_, err := decodeCredentialBindings(bindings)
	return err
}

func decodeCredentialBindings(rawBindings []json.RawMessage) ([]CredentialBinding, error) {
	bindings := make([]CredentialBinding, 0, len(rawBindings))
	ids := make(map[string]struct{}, len(rawBindings))
	targets := make(map[string]struct{}, len(rawBindings))
	for _, raw := range rawBindings {
		var document credentialBindingDocument
		if err := decodeStrict(raw, &document); err != nil {
			return nil, err
		}
		if err := slugkit.Validate(document.ID); err != nil {
			return nil, fmt.Errorf("invalid binding id %q: %w", document.ID, err)
		}
		if err := slugkit.Validate(document.Source.Ref); err != nil {
			return nil, fmt.Errorf("invalid binding source ref %q: %w", document.Source.Ref, err)
		}
		if document.Target.Name == "" ||
			(document.Source.Kind != "model_resource" && document.Source.Kind != "credential_bundle") ||
			document.Target.Kind != "env" {
			return nil, fmt.Errorf("credential binding must reference a supported source and env target")
		}
		if _, exists := ids[document.ID]; exists {
			return nil, fmt.Errorf("duplicate credential binding id %q", document.ID)
		}
		if _, exists := targets[document.Target.Name]; exists {
			return nil, fmt.Errorf("duplicate credential binding target %q", document.Target.Name)
		}
		ids[document.ID] = struct{}{}
		targets[document.Target.Name] = struct{}{}
		bindings = append(bindings, CredentialBinding{
			ID: document.ID,
			Source: CredentialSource{
				Kind: document.Source.Kind,
				Ref:  document.Source.Ref,
			},
			Target: CredentialTarget{
				Kind: document.Target.Kind,
				Name: document.Target.Name,
			},
		})
	}
	return bindings, nil
}

func validateCredentialBindingSchema(
	requirement ModelRequirement,
	schema *agentfileschema.AgentSchema,
	bindings []CredentialBinding,
	toolRequirements []ToolModelRequirement,
) error {
	if schema == nil {
		return fmt.Errorf("AgentFile schema is missing")
	}
	allowedAdapters := make(map[string]struct{}, len(requirement.ProtocolAdapters))
	for _, adapter := range requirement.ProtocolAdapters {
		allowedAdapters[adapter] = struct{}{}
	}
	byTarget := make(map[string]CredentialBinding, len(bindings))
	for _, binding := range bindings {
		if binding.Source.Kind == "model_resource" {
			if !requirement.Required {
				return fmt.Errorf("model resource binding %q is declared for a non-model worker", binding.ID)
			}
			if _, exists := allowedAdapters[binding.Source.Ref]; !exists {
				return fmt.Errorf(
					"model resource binding %q references unsupported protocol %q",
					binding.ID,
					binding.Source.Ref,
				)
			}
		}
		byTarget[binding.Target.Name] = binding
	}
	toolTargets := make(map[string]struct{}, len(toolRequirements)*3)
	for _, toolRequirement := range toolRequirements {
		for _, target := range []string{
			toolRequirement.Environment.APIKey,
			toolRequirement.Environment.BaseURL,
			toolRequirement.Environment.ModelID,
		} {
			if _, exists := byTarget[target]; exists {
				return fmt.Errorf("tool model environment target %q conflicts with credential binding", target)
			}
			toolTargets[target] = struct{}{}
		}
	}
	for _, field := range schema.CredentialFields {
		if _, exists := toolTargets[field.Name]; exists {
			delete(toolTargets, field.Name)
			continue
		}
		if _, exists := byTarget[field.Name]; !exists {
			return fmt.Errorf("AgentFile credential field %q has no binding", field.Name)
		}
		delete(byTarget, field.Name)
	}
	for target := range byTarget {
		return fmt.Errorf("credential binding target %q is not declared by AgentFile", target)
	}
	for target := range toolTargets {
		return fmt.Errorf("tool model environment target %q is not declared by AgentFile", target)
	}
	return nil
}
