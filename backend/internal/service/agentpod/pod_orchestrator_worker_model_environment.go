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
	case "codex-cli", "openclaw", "hermes", "opencode":
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
		return map[string]string{"GEMINI_API_KEY": apiKey}, nil
	case "minimax-cli":
		if modelID == "" {
			return nil, ErrMissingModelResource
		}
		return map[string]string{"MINIMAX_API_KEY": apiKey}, nil
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
	if len(modelArgs) != 2 || !supportedModelResourceArg(modelArgs[0]) || strings.TrimSpace(modelArgs[1]) == "" {
		return nil, ErrModelResourceCommandConflict
	}
	for i, arg := range existing {
		if !sameModelResourceArg(arg, modelArgs[0]) {
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

func supportedModelResourceArg(arg string) bool {
	return arg == "--model" || arg == "--base-url"
}

func sameModelResourceArg(existing, requested string) bool {
	return existing == requested || (requested == "--model" && existing == "-m")
}
