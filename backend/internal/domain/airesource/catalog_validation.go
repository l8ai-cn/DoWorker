package airesource

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func ValidateProviderDefinition(definition ProviderDefinition) error {
	if err := slugkit.Validate(definition.Key.String()); err != nil {
		return fmt.Errorf("provider key: %w", err)
	}
	if strings.TrimSpace(definition.DisplayName) == "" {
		return fmt.Errorf("provider %q has no display name", definition.Key)
	}
	if len(definition.Modalities) == 0 {
		return fmt.Errorf("provider %q has no modalities", definition.Key)
	}
	for _, modality := range definition.Modalities {
		if !modality.Valid() {
			return fmt.Errorf("provider %q has invalid modality %q", definition.Key, modality)
		}
	}
	if len(definition.CredentialFields) == 0 {
		return fmt.Errorf("provider %q has no credential fields", definition.Key)
	}
	credentialKeys := make(map[string]struct{}, len(definition.CredentialFields))
	for _, field := range definition.CredentialFields {
		if strings.TrimSpace(field.Key) == "" || strings.TrimSpace(field.Label) == "" {
			return fmt.Errorf("provider %q has an incomplete credential field", definition.Key)
		}
		credentialKeys[field.Key] = struct{}{}
	}
	if err := slugkit.Validate(definition.ProtocolAdapter); err != nil {
		return fmt.Errorf("provider %q protocol adapter: %w", definition.Key, err)
	}
	if err := validateConnectionCheck(definition, credentialKeys); err != nil {
		return err
	}
	return nil
}

func validateConnectionCheck(definition ProviderDefinition, credentialKeys map[string]struct{}) error {
	check := definition.ConnectionCheck
	if check.AuthStrategy == ConnectionAuthUnsupported {
		if check.Method != "" || check.Path != "" || check.CredentialKey != "" || check.AuthName != "" || len(check.StaticHeaders) > 0 {
			return fmt.Errorf("provider %q unsupported connection check has request fields", definition.Key)
		}
		return nil
	}
	if check.Method != http.MethodGet || !strings.HasPrefix(check.Path, "/") {
		return fmt.Errorf("provider %q has invalid connection check request", definition.Key)
	}
	if check.AuthStrategy != ConnectionAuthBearer && check.AuthStrategy != ConnectionAuthHeader && check.AuthStrategy != ConnectionAuthQuery {
		return fmt.Errorf("provider %q has invalid connection check auth strategy", definition.Key)
	}
	if strings.TrimSpace(check.CredentialKey) == "" {
		return fmt.Errorf("provider %q connection check has no credential key", definition.Key)
	}
	if _, declared := credentialKeys[check.CredentialKey]; !declared {
		return fmt.Errorf("provider %q connection check uses undeclared credential", definition.Key)
	}
	if (check.AuthStrategy == ConnectionAuthHeader || check.AuthStrategy == ConnectionAuthQuery) && strings.TrimSpace(check.AuthName) == "" {
		return fmt.Errorf("provider %q connection check has no auth name", definition.Key)
	}
	for _, header := range check.StaticHeaders {
		if strings.TrimSpace(header.Name) == "" || strings.TrimSpace(header.Value) == "" {
			return fmt.Errorf("provider %q connection check has invalid static header", definition.Key)
		}
	}
	return nil
}

func validateProviderRegistry(definitions []ProviderDefinition) error {
	providerKeys := make(map[slugkit.Slug]struct{}, len(definitions))
	for _, definition := range definitions {
		if err := ValidateProviderDefinition(definition); err != nil {
			return err
		}
		if _, exists := providerKeys[definition.Key]; exists {
			return fmt.Errorf("duplicate provider key %q", definition.Key)
		}
		providerKeys[definition.Key] = struct{}{}

		credentialKeys := make(map[string]struct{}, len(definition.CredentialFields))
		for _, field := range definition.CredentialFields {
			if _, exists := credentialKeys[field.Key]; exists {
				return fmt.Errorf("provider %q has duplicate credential field key %q", definition.Key, field.Key)
			}
			credentialKeys[field.Key] = struct{}{}
		}
	}
	return nil
}

func mustProviderRegistry(definitions []ProviderDefinition) []ProviderDefinition {
	if err := validateProviderRegistry(definitions); err != nil {
		panic(err)
	}
	return definitions
}
