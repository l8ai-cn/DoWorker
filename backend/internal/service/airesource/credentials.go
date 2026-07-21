package airesource

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
)

func validateCredentials(provider domain.ProviderDefinition, credentials map[string]string) ([]string, error) {
	allowed := make(map[string]domain.CredentialField, len(provider.CredentialFields))
	for _, field := range provider.CredentialFields {
		allowed[field.Key] = field
	}
	for key := range credentials {
		if _, exists := allowed[key]; !exists {
			return nil, fmt.Errorf("%w: unknown credential field", ErrInvalidCredentials)
		}
	}
	configured := make([]string, 0, len(credentials))
	for _, field := range provider.CredentialFields {
		value := strings.TrimSpace(credentials[field.Key])
		if field.Required && value == "" {
			return nil, fmt.Errorf("%w: required credential is empty", ErrInvalidCredentials)
		}
		if value != "" {
			configured = append(configured, field.Key)
		}
	}
	sort.Strings(configured)
	return configured, nil
}

func (s *Service) encryptCredentials(provider domain.ProviderDefinition, credentials map[string]string) (string, []string, error) {
	configured, err := validateCredentials(provider, credentials)
	if err != nil {
		return "", nil, err
	}
	encoded, err := json.Marshal(credentials)
	if err != nil {
		return "", nil, fmt.Errorf("encode AI resource credentials: %w", err)
	}
	encrypted, err := s.cipher.Encrypt(string(encoded))
	if err != nil {
		return "", nil, ErrEncrypt
	}
	return encrypted, configured, nil
}

func (s *Service) decryptCredentials(connection *domain.Connection) (map[string]string, error) {
	plaintext, err := s.cipher.Decrypt(connection.CredentialsEncrypted)
	if err != nil {
		return nil, ErrDecrypt
	}
	credentials := make(map[string]string)
	if err := json.Unmarshal([]byte(plaintext), &credentials); err != nil {
		return nil, fmt.Errorf("%w: malformed credential envelope", ErrDecrypt)
	}
	provider, exists := domain.Provider(connection.ProviderKey.String())
	if !exists {
		return nil, ErrInvalidProvider
	}
	if _, err := validateCredentials(provider, credentials); err != nil {
		return nil, fmt.Errorf("%w: stored credential envelope invalid", ErrDecrypt)
	}
	return credentials, nil
}
