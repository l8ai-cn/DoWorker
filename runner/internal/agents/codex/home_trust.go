package codex

import (
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

func AppendCodexProjectTrust(configPath, workDir string) error {
	if workDir == "" {
		return nil
	}
	if err := ensureCodexHeadlessDefaults(configPath); err != nil {
		return err
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var cfg map[string]interface{}
	if len(data) > 0 {
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse codex config: %w", err)
		}
	}
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	projects, _ := cfg["projects"].(map[string]interface{})
	if projects == nil {
		projects = map[string]interface{}{}
	}
	for _, path := range []string{workDir, filepath.Join(workDir, ".codex")} {
		entry, _ := projects[path].(map[string]interface{})
		if entry == nil {
			entry = map[string]interface{}{}
		}
		entry["trust_level"] = "trusted"
		projects[path] = entry
	}
	cfg["projects"] = projects
	out, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal codex config: %w", err)
	}
	return os.WriteFile(configPath, out, 0644)
}
