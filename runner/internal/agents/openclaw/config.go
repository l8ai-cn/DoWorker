package openclaw

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ApplyOpenAIProviderFromEnv(configPath, workspace, baseURL, model string) error {
	workspace = strings.TrimSpace(workspace)
	baseURL = strings.TrimSpace(baseURL)
	model = strings.TrimSpace(model)
	if workspace == "" {
		return fmt.Errorf("openclaw workspace is required")
	}
	if baseURL == "" {
		return fmt.Errorf("openclaw OpenAI base URL is required")
	}
	if model == "" {
		return fmt.Errorf("openclaw model is required")
	}

	config, err := readConfig(configPath)
	if err != nil {
		return err
	}
	agents := objectField(config, "agents")
	defaults := objectField(agents, "defaults")
	defaults["workspace"] = workspace
	defaults["model"] = qualifyModel(model)

	models := objectField(config, "models")
	models["mode"] = "merge"
	providers := objectField(models, "providers")
	provider := objectField(providers, "openai")
	provider["baseUrl"] = baseURL
	provider["apiKey"] = map[string]interface{}{
		"source":   "env",
		"provider": "default",
		"id":       "OPENAI_API_KEY",
	}
	provider["auth"] = "api-key"
	provider["api"] = "openai-completions"
	provider["models"] = []interface{}{
		map[string]interface{}{"id": modelID(model), "name": modelID(model)},
	}

	return writeConfig(configPath, config)
}

func MergeConfig(configPath, content string) error {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(content), &config); err != nil {
		return fmt.Errorf("parse openclaw config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("openclaw config must be a JSON object")
	}
	return writeConfig(configPath, config)
}

func readConfig(configPath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("read openclaw config: %w", err)
	}
	config := map[string]interface{}{}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("parse openclaw config: %w", err)
		}
	}
	if config == nil {
		config = map[string]interface{}{}
	}
	return config, nil
}

func writeConfig(configPath string, config map[string]interface{}) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal openclaw config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("create openclaw config directory: %w", err)
	}
	if err := os.WriteFile(configPath, append(data, '\n'), 0600); err != nil {
		return fmt.Errorf("write openclaw config: %w", err)
	}
	return nil
}

func objectField(parent map[string]interface{}, key string) map[string]interface{} {
	if value, ok := parent[key].(map[string]interface{}); ok {
		return value
	}
	value := map[string]interface{}{}
	parent[key] = value
	return value
}

func qualifyModel(model string) string {
	if strings.Contains(model, "/") {
		return model
	}
	return "openai/" + model
}

func modelID(model string) string {
	if index := strings.LastIndex(model, "/"); index >= 0 {
		return model[index+1:]
	}
	return model
}
