package agentpod

import (
	"fmt"
	"strings"

	resourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
)

func modelResourceEnvironment(agentSlug string, resource *resourcesvc.ResolvedResource) (map[string]string, error) {
	apiKey := modelResourceAPIKey(resource)
	if apiKey == "" {
		return nil, ErrMissingModelResource
	}
	modelID := strings.TrimSpace(resource.Resource.ModelID)
	baseURL := strings.TrimSpace(resource.Connection.BaseURL)
	switch agentSlug {
	case "codex-cli":
		return compactEnv(map[string]string{
			"OPENAI_API_KEY":  apiKey,
			"OPENAI_BASE_URL": baseURL,
			"OPENAI_MODEL":    modelID,
		}), nil
	case "claude-code":
		return compactEnv(map[string]string{
			"ANTHROPIC_API_KEY":  apiKey,
			"ANTHROPIC_BASE_URL": baseURL,
		}), nil
	case "gemini-cli":
		if modelID == "" {
			return nil, ErrMissingModelResource
		}
		return map[string]string{"GOOGLE_API_KEY": apiKey}, nil
	case "grok-build":
		if resource.Connection.ProviderKey.String() != "xai" {
			return nil, fmt.Errorf("%w: grok-build requires xai", ErrModelResourceProviderUnsupported)
		}
		return map[string]string{"XAI_API_KEY": apiKey}, nil
	case "minimax-cli":
		if modelID == "" {
			return nil, ErrMissingModelResource
		}
		return map[string]string{"MINIMAX_API_KEY": apiKey}, nil
	case "openclaw", "hermes":
		env := multiProviderModelResourceEnvironment(resource, apiKey, baseURL, modelID)
		if len(env) == 0 {
			return nil, ErrMissingModelResource
		}
		return env, nil
	default:
		return nil, ErrMissingModelResource
	}
}

func modelResourceAPIKey(resource *resourcesvc.ResolvedResource) string {
	if resource == nil || resource.Credentials == nil {
		return ""
	}
	if value := strings.TrimSpace(resource.Credentials["api_key"]); value != "" {
		return value
	}
	if value := strings.TrimSpace(resource.Credentials["auth_token"]); value != "" {
		return value
	}
	return strings.TrimSpace(resource.Credentials["api_token"])
}

func doAgentModelSettings(resource *resourcesvc.ResolvedResource) (map[string]interface{}, error) {
	if resource == nil {
		return nil, ErrMissingModelResource
	}
	apiKey := modelResourceAPIKey(resource)
	model := strings.TrimSpace(resource.Resource.ModelID)
	if apiKey == "" || model == "" {
		return nil, ErrMissingModelResource
	}
	options := map[string]interface{}{"apiKey": apiKey}
	if baseURL := strings.TrimSpace(resource.Connection.BaseURL); baseURL != "" {
		options["baseURL"] = baseURL
	}
	if resource.Provider.ProtocolAdapter == "openai-compatible" {
		options["kind"] = "openai"
	} else {
		options["kind"] = "anthropic"
	}
	provider := resource.Provider.Key.String()
	if !strings.Contains(model, "/") {
		model = provider + "/" + model
	}
	return map[string]interface{}{
		"provider": map[string]interface{}{provider: map[string]interface{}{"options": options}},
		"model":    model,
	}, nil
}

func multiProviderModelResourceEnvironment(
	resource *resourcesvc.ResolvedResource,
	apiKey, baseURL, modelID string,
) map[string]string {
	switch resource.Provider.ProtocolAdapter {
	case "openai-compatible":
		env := map[string]string{
			"OPENAI_API_KEY":  apiKey,
			"OPENAI_BASE_URL": baseURL,
			"OPENAI_MODEL":    modelID,
		}
		if resource.Provider.Key.String() == "xai" {
			env["XAI_API_KEY"] = apiKey
		}
		return compactEnv(env)
	case "anthropic":
		return compactEnv(map[string]string{
			"ANTHROPIC_API_KEY":  apiKey,
			"ANTHROPIC_BASE_URL": baseURL,
			"ANTHROPIC_MODEL":    modelID,
		})
	case "gemini":
		return compactEnv(map[string]string{
			"GOOGLE_API_KEY": apiKey,
			"GEMINI_API_KEY": apiKey,
			"GEMINI_MODEL":   modelID,
		})
	default:
		return nil
	}
}

func compactEnv(values map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out[key] = trimmed
		}
	}
	return out
}

func applyModelResourceEnv(existing, modelEnv map[string]string) error {
	for key, value := range modelEnv {
		if current, exists := existing[key]; exists && current != value {
			return ErrModelResourceEnvConflict
		}
	}
	for key, value := range modelEnv {
		existing[key] = value
	}
	return nil
}

func applyModelResourceArgs(existing, modelArgs []string) ([]string, error) {
	if len(modelArgs) == 0 {
		return existing, nil
	}
	if len(modelArgs) != 2 || modelArgs[0] != "--model" || strings.TrimSpace(modelArgs[1]) == "" {
		return nil, ErrModelResourceCommandConflict
	}
	for i, arg := range existing {
		if arg != "--model" && arg != "-m" {
			continue
		}
		if i+1 >= len(existing) || existing[i+1] != modelArgs[1] {
			return nil, ErrModelResourceCommandConflict
		}
		return existing, nil
	}
	out := append([]string(nil), existing...)
	return append(out, modelArgs...), nil
}
