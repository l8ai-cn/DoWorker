package codex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// WriteAuthJSONFromEnv materialises CODEX_HOME/auth.json from an API key the
// backend injects at session create (sourced from the ai_models pool). Codex
// with wire_api=responses reads the key from this file, so the platform must
// generate it per-pod instead of depending on a pre-baked container copy.
// An empty key is a no-op so a copied container auth.json still applies.
func WriteAuthJSONFromEnv(codexHome, apiKey string) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil
	}
	if err := os.MkdirAll(codexHome, 0o755); err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]string{
		"OPENAI_API_KEY": apiKey,
		"auth_mode":      "apikey",
	})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(codexHome, "auth.json"), payload, 0o600)
}
