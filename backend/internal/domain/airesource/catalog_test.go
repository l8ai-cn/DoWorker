package airesource

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func TestCatalogContainsRequiredProviders(t *testing.T) {
	required := []string{
		"openai", "anthropic", "gemini", "azure-openai", "openrouter",
		"deepseek", "dashscope", "doubao", "minimax", "zhipu", "moonshot",
		"xai", "mistral", "stability-ai", "black-forest-labs", "replicate",
		"fal", "ideogram", "elevenlabs", "azure-speech", "runway", "kling",
		"hailuo", "luma", "custom-openai-compatible", "custom-anthropic-compatible",
		"sub2api-seedance",
	}

	for _, key := range required {
		definition, ok := Provider(key)
		if !ok {
			t.Errorf("Provider(%q) was not registered", key)
			continue
		}
		if len(definition.Modalities) == 0 {
			t.Errorf("Provider(%q) has no modalities", key)
		}
		if len(definition.CredentialFields) == 0 {
			t.Errorf("Provider(%q) has no credential fields", key)
		}
	}
}

func TestProviderCapabilityMetadata(t *testing.T) {
	openAI, ok := Provider("openai")
	if !ok {
		t.Fatal("openai provider was not registered")
	}
	if openAI.SupportsCustomEndpoint || !openAI.SupportsModelDiscovery {
		t.Fatal("openai must use its fixed endpoint and expose model discovery")
	}

	customOpenAI, ok := Provider("custom-openai-compatible")
	if !ok {
		t.Fatal("custom-openai-compatible provider was not registered")
	}
	if !customOpenAI.SupportsCustomEndpoint || !customOpenAI.SupportsModelDiscovery {
		t.Fatal("custom-openai-compatible must expose endpoint and discovery support")
	}

	runway, ok := Provider("runway")
	if !ok {
		t.Fatal("runway provider was not registered")
	}
	if runway.SupportsCustomEndpoint || runway.SupportsModelDiscovery {
		t.Fatal("runway must use its fixed endpoint and declared model catalog")
	}
}

func TestProviderConnectionChecksAreCodeOwned(t *testing.T) {
	tests := []struct {
		key      string
		strategy ConnectionAuthStrategy
		path     string
	}{
		{key: "openai", strategy: ConnectionAuthBearer, path: "/models"},
		{key: "openrouter", strategy: ConnectionAuthBearer, path: "/key"},
		{key: "deepseek", strategy: ConnectionAuthBearer, path: "/models"},
		{key: "xai", strategy: ConnectionAuthBearer, path: "/models"},
		{key: "mistral", strategy: ConnectionAuthBearer, path: "/models"},
		{key: "anthropic", strategy: ConnectionAuthHeader, path: "/v1/models"},
		{key: "gemini", strategy: ConnectionAuthQuery, path: "/v1beta/models"},
		{key: "azure-openai", strategy: ConnectionAuthHeader, path: "/openai/v1/models"},
		{key: "stability-ai", strategy: ConnectionAuthBearer, path: "/v1/user/account"},
		{key: "black-forest-labs", strategy: ConnectionAuthHeader, path: "/credits"},
		{key: "elevenlabs", strategy: ConnectionAuthHeader, path: "/models"},
		{key: "runway", strategy: ConnectionAuthBearer, path: "/organization"},
		{key: "luma", strategy: ConnectionAuthBearer, path: "/generations"},
		{key: "ideogram", strategy: ConnectionAuthHeader, path: "/models"},
		{key: "kling", strategy: ConnectionAuthUnsupported, path: ""},
		{key: "sub2api-seedance", strategy: ConnectionAuthBearer, path: "/contents/generations/tasks"},
	}
	for _, test := range tests {
		definition, ok := Provider(test.key)
		if !ok {
			t.Fatalf("Provider(%q) missing", test.key)
		}
		if definition.ConnectionCheck.AuthStrategy != test.strategy {
			t.Errorf("Provider(%q) auth strategy = %q", test.key, definition.ConnectionCheck.AuthStrategy)
		}
		if definition.ConnectionCheck.Path != test.path {
			t.Errorf("Provider(%q) check path = %q", test.key, definition.ConnectionCheck.Path)
		}
	}
}

func TestEveryProviderHasAnExplicitConnectionCheckPolicy(t *testing.T) {
	for _, definition := range Providers() {
		if definition.ConnectionCheck.AuthStrategy == "" {
			t.Errorf("provider %q has no explicit connection check", definition.Key)
		}
	}
}

func TestProvidersReturnsDeepCopies(t *testing.T) {
	original, ok := Provider("anthropic")
	if !ok {
		t.Fatal("anthropic provider was not registered")
	}
	providers := Providers()
	for i := range providers {
		if providers[i].Key != original.Key {
			continue
		}
		providers[i].DisplayName = "changed"
		providers[i].Modalities[0] = Modality("changed")
		providers[i].CredentialFields[0].Key = "changed"
		providers[i].ConnectionCheck.StaticHeaders[0].Value = "changed"
		break
	}

	after, _ := Provider("anthropic")
	if after.DisplayName != original.DisplayName {
		t.Fatal("Providers allowed mutation of the registered display name")
	}
	if after.Modalities[0] != original.Modalities[0] {
		t.Fatal("Providers allowed mutation of registered modalities")
	}
	if after.CredentialFields[0].Key != original.CredentialFields[0].Key {
		t.Fatal("Providers allowed mutation of registered credential fields")
	}
	if after.ConnectionCheck.StaticHeaders[0].Value != original.ConnectionCheck.StaticHeaders[0].Value {
		t.Fatal("Providers allowed mutation of registered connection check headers")
	}
}

func TestProviderReturnsDeepCopies(t *testing.T) {
	returned, ok := Provider("anthropic")
	if !ok {
		t.Fatal("anthropic provider was not registered")
	}
	returned.DisplayName = "changed"
	returned.Modalities[0] = Modality("changed")
	returned.CredentialFields[0].Key = "changed"
	returned.ConnectionCheck.StaticHeaders[0].Value = "changed"

	after, _ := Provider("anthropic")
	if after.DisplayName == "changed" {
		t.Fatal("Provider allowed mutation of the registered display name")
	}
	if after.Modalities[0] == Modality("changed") {
		t.Fatal("Provider allowed mutation of registered modalities")
	}
	if after.CredentialFields[0].Key == "changed" {
		t.Fatal("Provider allowed mutation of registered credential fields")
	}
	if after.ConnectionCheck.StaticHeaders[0].Value == "changed" {
		t.Fatal("Provider allowed mutation of registered connection check headers")
	}
}

func TestValidateProviderDefinitionRejectsInvalidDefinitions(t *testing.T) {
	valid := ProviderDefinition{
		Key:              slugkit.Slug("provider"),
		DisplayName:      "Provider",
		Modalities:       []Modality{ModalityChat},
		CredentialFields: []CredentialField{{Key: "api-key", Label: "API key", Secret: true, Required: true}},
		ProtocolAdapter:  "openai-compatible",
		ConnectionCheck:  bearerCheck("/models", "api-key"),
	}
	if err := ValidateProviderDefinition(valid); err != nil {
		t.Fatalf("test baseline must be valid: %v", err)
	}
	tests := []struct {
		name       string
		definition ProviderDefinition
	}{
		{name: "uppercase key", definition: replaceProviderKey(valid, "OpenAI")},
		{name: "underscore key", definition: replaceProviderKey(valid, "open_ai")},
		{name: "empty modalities", definition: replaceModalities(valid, nil)},
		{name: "empty credential fields", definition: replaceCredentialFields(valid, nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateProviderDefinition(test.definition); err == nil {
				t.Fatalf("ValidateProviderDefinition accepted %s", test.name)
			}
		})
	}
}

func TestValidateProviderDefinitionRejectsUndeclaredProbeCredentials(t *testing.T) {
	definition, ok := Provider("openai")
	if !ok {
		t.Fatal("openai provider was not registered")
	}
	definition.ConnectionCheck.CredentialKey = "missing_key"
	if err := ValidateProviderDefinition(definition); err == nil {
		t.Fatal("ValidateProviderDefinition accepted an undeclared probe credential")
	}

}

func TestValidateModelResourceEnforcesIdentifierContract(t *testing.T) {
	valid := ModelResource{
		Identifier: slugkit.Slug("gpt-4-1"),
		ModelID:    "gpt-4.1",
		Modalities: []Modality{ModalityChat},
	}
	invalid := []slugkit.Slug{"GPT-4", "gpt_4"}
	for _, identifier := range invalid {
		resource := valid
		resource.Identifier = identifier
		if err := ValidateModelResource(resource); err == nil {
			t.Errorf("ValidateModelResource accepted identifier %q", identifier)
		}
	}
	if err := ValidateModelResource(valid); err != nil {
		t.Fatalf("ValidateModelResource rejected a valid resource: %v", err)
	}
}

func TestConnectionValidateIdentifiers(t *testing.T) {
	valid := Connection{
		Identifier:  slugkit.Slug("work-openai"),
		ProviderKey: slugkit.Slug("openai"),
	}
	if err := valid.ValidateIdentifiers(); err != nil {
		t.Fatalf("ValidateIdentifiers rejected valid identifiers: %v", err)
	}

	tests := []struct {
		name       string
		connection Connection
	}{
		{name: "uppercase identifier", connection: replaceConnectionIdentifier(valid, "Work-openai")},
		{name: "underscore identifier", connection: replaceConnectionIdentifier(valid, "work_openai")},
		{name: "uppercase provider key", connection: replaceConnectionProviderKey(valid, "OpenAI")},
		{name: "underscore provider key", connection: replaceConnectionProviderKey(valid, "open_ai")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.connection.ValidateIdentifiers(); err == nil {
				t.Fatalf("ValidateIdentifiers accepted %s", test.name)
			}
		})
	}
}

func TestModelResourceJSONCarriesValidationAndOptionalUsage(t *testing.T) {
	validatedAt := time.Date(2026, time.July, 10, 6, 0, 0, 0, time.UTC)
	usageTotal := 12.5
	resource := ModelResource{
		Status:          ConnectionStatusInvalid,
		LastValidatedAt: &validatedAt,
		ValidationError: "model unavailable",
		UsageSummary:    &UsageSummary{UsageTotal: &usageTotal},
	}
	encoded, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	for _, field := range []string{"status", "last_validated_at", "validation_error", "usage_summary"} {
		if !strings.Contains(string(encoded), `"`+field+`"`) {
			t.Errorf("model resource JSON omitted %q", field)
		}
	}

	resource.UsageSummary = nil
	encoded, err = json.Marshal(resource)
	if err != nil {
		t.Fatalf("json.Marshal without usage failed: %v", err)
	}
	if strings.Contains(string(encoded), `"usage_summary"`) {
		t.Fatal("model resource JSON must omit usage_summary when accounting is unavailable")
	}
}

func TestConnectionJSONOmitsEncryptedCredentials(t *testing.T) {
	connection := Connection{
		CredentialsEncrypted: "plaintext-must-not-leak",
		ConfiguredFields:     []string{"api_key"},
	}
	encoded, err := json.Marshal(connection)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	if strings.Contains(string(encoded), "plaintext-must-not-leak") {
		t.Fatal("connection JSON exposed credentials")
	}
	if !strings.Contains(string(encoded), `"configured_fields"`) {
		t.Fatal("connection JSON omitted safe configured field names")
	}
}

func replaceProviderKey(definition ProviderDefinition, key string) ProviderDefinition {
	definition.Key = slugkit.Slug(key)
	return definition
}

func replaceModalities(definition ProviderDefinition, modalities []Modality) ProviderDefinition {
	definition.Modalities = modalities
	return definition
}

func replaceCredentialFields(definition ProviderDefinition, fields []CredentialField) ProviderDefinition {
	definition.CredentialFields = fields
	return definition
}

func replaceConnectionIdentifier(connection Connection, identifier string) Connection {
	connection.Identifier = slugkit.Slug(identifier)
	return connection
}

func replaceConnectionProviderKey(connection Connection, key string) Connection {
	connection.ProviderKey = slugkit.Slug(key)
	return connection
}
