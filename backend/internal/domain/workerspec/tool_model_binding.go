package workerspec

import (
	"fmt"
	"regexp"

	resourcedomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

var toolModelEnvironmentPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

type ToolModelBinding struct {
	Role         slugkit.Slug              `json:"role"`
	ModelBinding ModelBinding              `json:"model_binding"`
	Modality     resourcedomain.Modality   `json:"modality"`
	Capability   resourcedomain.Capability `json:"capability"`
	Environment  ToolModelEnvironment      `json:"environment"`
}

type ToolModelEnvironment struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	ModelID string `json:"model_id"`
}

func validateToolModelBindings(bindings []ToolModelBinding) error {
	roles := make(map[slugkit.Slug]struct{}, len(bindings))
	targets := map[string]struct{}{}
	for _, binding := range bindings {
		if err := slugkit.Validate(binding.Role.String()); err != nil {
			return fmt.Errorf("runtime tool model role: %w", err)
		}
		if _, exists := roles[binding.Role]; exists {
			return fmt.Errorf("duplicate tool model role %q", binding.Role)
		}
		roles[binding.Role] = struct{}{}
		if err := validateModelBinding(binding.ModelBinding); err != nil {
			return fmt.Errorf("runtime tool model %q: %w", binding.Role, err)
		}
		if !binding.Modality.Valid() || !binding.Capability.Valid() {
			return fmt.Errorf("runtime tool model %q has invalid capability contract", binding.Role)
		}
		for _, target := range []string{
			binding.Environment.APIKey,
			binding.Environment.BaseURL,
			binding.Environment.ModelID,
		} {
			if !toolModelEnvironmentPattern.MatchString(target) {
				return fmt.Errorf("runtime tool model %q has invalid environment target", binding.Role)
			}
			if _, exists := targets[target]; exists {
				return fmt.Errorf("duplicate tool model environment target %q", target)
			}
			targets[target] = struct{}{}
		}
	}
	return nil
}

func cloneToolModelBindings(bindings []ToolModelBinding) []ToolModelBinding {
	if bindings == nil {
		return nil
	}
	return append([]ToolModelBinding{}, bindings...)
}
