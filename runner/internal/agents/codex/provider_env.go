package codex

import (
	"fmt"
	"os"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

// ApplyOpenAIProviderFromEnv patches codex config.toml when OPENAI_BASE_URL
// arrives from an env_bundles credential (proxy setups). Container ~/.codex
// copies often lack model_providers.OpenAI, so Codex would hit api.openai.com.
func ApplyOpenAIProviderFromEnv(configPath, baseURL, model string) error {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil
	}

	cfg, err := readCodexConfigMap(configPath)
	if err != nil {
		return err
	}

	if m := strings.TrimSpace(model); m != "" {
		cfg["model"] = m
	}
	cfg["model_provider"] = "OpenAI"

	providers, _ := cfg["model_providers"].(map[string]interface{})
	if providers == nil {
		providers = map[string]interface{}{}
	}
	openAI, _ := providers["OpenAI"].(map[string]interface{})
	if openAI == nil {
		openAI = map[string]interface{}{}
	}
	openAI["name"] = "OpenAI"
	openAI["base_url"] = baseURL
	openAI["wire_api"] = "responses"
	// env_key makes Codex attach `Authorization: Bearer $OPENAI_API_KEY`; the
	// proxy gateways we target reject unauthenticated /responses calls (401).
	openAI["env_key"] = "OPENAI_API_KEY"
	openAI["requires_openai_auth"] = false
	providers["OpenAI"] = openAI
	cfg["model_providers"] = providers

	out, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal codex provider config: %w", err)
	}
	return os.WriteFile(configPath, out, 0o644)
}

func readCodexConfigMap(configPath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, err
	}
	cfg := map[string]interface{}{}
	if len(data) > 0 {
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse codex config: %w", err)
		}
	}
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	return cfg, nil
}
