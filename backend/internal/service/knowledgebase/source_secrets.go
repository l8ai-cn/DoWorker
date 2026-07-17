package knowledgebase

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
)

// Connector credentials live inside source_config JSONB. Values of these
// keys are encrypted at rest (same platform convention as
// user_git_credentials' encrypted fields) and only decrypted in-process
// right before a connector call.
var secretConfigKeys = map[string]bool{
	"app_secret":             true,
	"access_token":           true,
	readOnlyDeployKeyConfig:  true,
	readWriteDeployKeyConfig: true,
}

var internalSecretConfigKeys = map[string]bool{
	readOnlyDeployKeyConfig:  true,
	readWriteDeployKeyConfig: true,
}

const encPrefix = "enc:v1:"
const secretPlaceholder = "***"

// RedactedSourceConfigJSON returns source_config safe for API responses.
func RedactedSourceConfigJSON(raw json.RawMessage) string {
	redacted, err := redactSourceSecrets(raw)
	if err != nil || len(redacted) == 0 {
		return string(raw)
	}
	return string(redacted)
}

func redactSourceSecrets(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return raw, nil
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	changed := false
	for key, val := range cfg {
		if internalSecretConfigKeys[key] {
			delete(cfg, key)
			changed = true
			continue
		}
		str, ok := val.(string)
		if !ok || !secretConfigKeys[key] || str == "" {
			continue
		}
		cfg[key] = secretPlaceholder
		changed = true
	}
	if !changed {
		return raw, nil
	}
	return json.Marshal(cfg)
}

// mergeSourceConfigUpdate overlays incoming onto existing. Blank or "***"
// secret values preserve the stored credential.
func (s *Service) mergeSourceConfigUpdate(existing, incoming json.RawMessage) (json.RawMessage, error) {
	var base map[string]any
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &base); err != nil {
			return nil, fmt.Errorf("%w: existing source_config invalid: %v", ErrInvalidInput, err)
		}
	}
	if base == nil {
		base = map[string]any{}
	}
	var patch map[string]any
	if err := json.Unmarshal(incoming, &patch); err != nil {
		return nil, fmt.Errorf("%w: source_config must be a JSON object: %v", ErrInvalidInput, err)
	}
	for key, val := range patch {
		str, isStr := val.(string)
		if isStr && secretConfigKeys[key] && (str == "" || str == secretPlaceholder) {
			continue
		}
		base[key] = val
	}
	return json.Marshal(base)
}

// SetSecretsEncryptor enables at-rest encryption of connector credentials.
// A nil encryptor (tests, misconfigured deployments) stores configs verbatim.
func (s *Service) SetSecretsEncryptor(enc *crypto.Encryptor) { s.secrets = enc }

func (s *Service) encryptSourceSecrets(raw json.RawMessage) (json.RawMessage, error) {
	return s.mapSourceSecrets(raw, func(v string) (string, error) {
		if strings.HasPrefix(v, encPrefix) {
			return v, nil
		}
		ct, err := s.secrets.Encrypt(v)
		if err != nil {
			return "", err
		}
		return encPrefix + ct, nil
	})
}

func (s *Service) decryptSourceSecrets(raw json.RawMessage) (json.RawMessage, error) {
	return s.mapSourceSecrets(raw, func(v string) (string, error) {
		if !strings.HasPrefix(v, encPrefix) {
			return v, nil
		}
		return s.secrets.Decrypt(strings.TrimPrefix(v, encPrefix))
	})
}

func (s *Service) mapSourceSecrets(raw json.RawMessage, transform func(string) (string, error)) (json.RawMessage, error) {
	if s.secrets == nil || len(raw) == 0 {
		return raw, nil
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("%w: source_config must be a JSON object: %v", ErrInvalidInput, err)
	}
	changed := false
	for key, val := range cfg {
		str, ok := val.(string)
		if !ok || !secretConfigKeys[key] || str == "" {
			continue
		}
		next, err := transform(str)
		if err != nil {
			return nil, fmt.Errorf("source_config secret %q: %w", key, err)
		}
		if next != str {
			cfg[key] = next
			changed = true
		}
	}
	if !changed {
		return raw, nil
	}
	return json.Marshal(cfg)
}
