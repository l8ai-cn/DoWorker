package workerdefinition

import (
	"encoding/json"
	"fmt"
	"regexp"

	resourcedomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

var environmentTargetPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

type toolModelRequirementDocument struct {
	ID               string                       `json:"id"`
	ProviderKeys     []string                     `json:"provider_keys"`
	ProtocolAdapters []string                     `json:"protocol_adapters"`
	Modality         string                       `json:"modality"`
	Capability       string                       `json:"capability"`
	Environment      toolModelEnvironmentDocument `json:"environment"`
}

type toolModelEnvironmentDocument struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	ModelID string `json:"model_id"`
}

func decodeToolModelRequirements(rawItems []json.RawMessage) ([]ToolModelRequirement, error) {
	requirements := make([]ToolModelRequirement, 0, len(rawItems))
	roles := make(map[string]struct{}, len(rawItems))
	targets := map[string]struct{}{}
	for _, raw := range rawItems {
		var document toolModelRequirementDocument
		if err := decodeStrict(raw, &document); err != nil {
			return nil, err
		}
		if err := validateToolModelRequirementDocument(document, roles, targets); err != nil {
			return nil, err
		}
		roles[document.ID] = struct{}{}
		requirements = append(requirements, ToolModelRequirement{
			ID:               document.ID,
			ProviderKeys:     append([]string{}, document.ProviderKeys...),
			ProtocolAdapters: append([]string{}, document.ProtocolAdapters...),
			Modality:         document.Modality,
			Capability:       document.Capability,
			Environment: ToolModelEnvironment{
				APIKey: document.Environment.APIKey, BaseURL: document.Environment.BaseURL,
				ModelID: document.Environment.ModelID,
			},
		})
	}
	return requirements, nil
}

func validateToolModelRequirementDocument(
	document toolModelRequirementDocument,
	roles, targets map[string]struct{},
) error {
	if err := slugkit.Validate(document.ID); err != nil {
		return fmt.Errorf("invalid tool model requirement id %q: %w", document.ID, err)
	}
	if _, exists := roles[document.ID]; exists {
		return fmt.Errorf("duplicate tool model requirement %q", document.ID)
	}
	if len(document.ProviderKeys) == 0 || len(document.ProtocolAdapters) == 0 {
		return fmt.Errorf("tool model requirement %q must declare providers and protocols", document.ID)
	}
	if err := validateUniqueSlugs("provider key", document.ProviderKeys); err != nil {
		return err
	}
	if err := validateUniqueSlugs("protocol adapter", document.ProtocolAdapters); err != nil {
		return err
	}
	if !resourcedomain.Modality(document.Modality).Valid() {
		return fmt.Errorf("tool model requirement %q has invalid modality", document.ID)
	}
	if !resourcedomain.Capability(document.Capability).Valid() {
		return fmt.Errorf("tool model requirement %q has invalid capability", document.ID)
	}
	for _, target := range []string{
		document.Environment.APIKey,
		document.Environment.BaseURL,
		document.Environment.ModelID,
	} {
		if !environmentTargetPattern.MatchString(target) {
			return fmt.Errorf("tool model requirement %q has invalid environment target", document.ID)
		}
		if _, exists := targets[target]; exists {
			return fmt.Errorf("duplicate environment target %q", target)
		}
		targets[target] = struct{}{}
	}
	return nil
}

func validateUniqueSlugs(label string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if err := slugkit.Validate(value); err != nil {
			return fmt.Errorf("invalid %s %q: %w", label, value, err)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("duplicate %s %q", label, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}
