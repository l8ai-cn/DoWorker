package codex

import (
	"fmt"
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

func mergeTomlMcpServers(configPath, platformContent string) error {
	var platformConfig map[string]interface{}
	if err := toml.Unmarshal([]byte(platformContent), &platformConfig); err != nil {
		return fmt.Errorf("failed to parse platform TOML: %w", err)
	}

	platformServers, _ := platformConfig["mcp_servers"].(map[string]interface{})
	if len(platformServers) == 0 {
		if _, err := os.Stat(configPath); err == nil {
			return ensureCodexHeadlessDefaults(configPath)
		}
		return nil
	}

	var existingConfig map[string]interface{}
	existingData, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(configPath, []byte(platformContent), 0644); err != nil {
				return err
			}
			return ensureCodexHeadlessDefaults(configPath)
		}
		return fmt.Errorf("failed to read existing config: %w", err)
	}

	if err := toml.Unmarshal(existingData, &existingConfig); err != nil {
		return fmt.Errorf("failed to parse existing config: %w", err)
	}
	if existingConfig == nil {
		existingConfig = make(map[string]interface{})
	}

	existingServers, _ := existingConfig["mcp_servers"].(map[string]interface{})
	if existingServers == nil {
		existingServers = make(map[string]interface{})
	}
	for k, v := range platformServers {
		existingServers[k] = v
	}
	existingConfig["mcp_servers"] = existingServers

	merged, err := toml.Marshal(existingConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal merged config: %w", err)
	}

	if err := os.WriteFile(configPath, merged, 0644); err != nil {
		return err
	}
	return ensureCodexHeadlessDefaults(configPath)
}

func ensureCodexHeadlessDefaults(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var cfg map[string]interface{}
	if len(data) == 0 {
		cfg = map[string]interface{}{}
	} else if err := toml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse codex config: %w", err)
	}
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	cfg["approval_policy"] = "never"
	cfg["sandbox_mode"] = "danger-full-access"
	out, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal codex config: %w", err)
	}
	return os.WriteFile(configPath, out, 0644)
}
