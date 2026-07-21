package airesource

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type modalityDefaultSetter interface {
	SetDefault(ctx context.Context, resourceID int64, modality Modality) error
}

var _ modalityDefaultSetter = (Repository)(nil)

func TestValidateModelResourceAcceptsPerModalityDefault(t *testing.T) {
	resource := validMultimodalResource()
	resource.DefaultModalities = []Modality{ModalityChat}

	if err := ValidateModelResource(resource); err != nil {
		t.Fatalf("ValidateModelResource rejected a chat-only default: %v", err)
	}
	if len(resource.DefaultModalities) != 1 || resource.DefaultModalities[0] != ModalityChat {
		t.Fatal("chat should be default while image remains non-default")
	}
}

func TestValidateModelResourceRejectsInvalidDefaultModalities(t *testing.T) {
	tests := []struct {
		name     string
		defaults []Modality
	}{
		{name: "not supported by resource", defaults: []Modality{ModalityAudio}},
		{name: "duplicate", defaults: []Modality{ModalityChat, ModalityChat}},
		{name: "unknown", defaults: []Modality{Modality("CHAT")}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resource := validMultimodalResource()
			resource.DefaultModalities = test.defaults
			if err := ValidateModelResource(resource); err == nil {
				t.Fatalf("ValidateModelResource accepted %s default modalities", test.name)
			}
		})
	}
}

func TestModelResourceJSONUsesDefaultModalities(t *testing.T) {
	resource := validMultimodalResource()
	resource.DefaultModalities = []Modality{ModalityChat}
	encoded, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	if !strings.Contains(string(encoded), `"default_modalities":["chat"]`) {
		t.Fatalf("model resource JSON omitted per-modality defaults: %s", encoded)
	}
	if strings.Contains(string(encoded), `"is_default"`) {
		t.Fatalf("model resource JSON retained the scalar default contract: %s", encoded)
	}
}

func TestValidateProviderRegistryRejectsDuplicateProviderKeys(t *testing.T) {
	definition := providerDefinitionForInvariantTest("provider", "api_key")
	if err := validateProviderRegistry([]ProviderDefinition{definition, definition}); err == nil {
		t.Fatal("validateProviderRegistry accepted a duplicate provider key")
	}
}

func TestValidateProviderRegistryRejectsDuplicateCredentialKeys(t *testing.T) {
	definition := providerDefinitionForInvariantTest("provider", "api_key", "api_key")
	if err := validateProviderRegistry([]ProviderDefinition{definition}); err == nil {
		t.Fatal("validateProviderRegistry accepted duplicate credential field keys")
	}
}

func TestCodeOwnedProviderRegistrySatisfiesInvariants(t *testing.T) {
	if err := validateProviderRegistry(Providers()); err != nil {
		t.Fatalf("code-owned provider registry is invalid: %v", err)
	}
}

func TestValidateModelResourceRejectsUnknownCapability(t *testing.T) {
	resource := validMultimodalResource()
	resource.Capabilities = []Capability{CapabilityTextGeneration}
	if err := ValidateModelResource(resource); err != nil {
		t.Fatalf("ValidateModelResource rejected a known capability: %v", err)
	}

	resource.Capabilities = []Capability{Capability("text-generation-v2")}
	if err := ValidateModelResource(resource); err == nil {
		t.Fatal("ValidateModelResource accepted an unknown capability")
	}
	if Capability("text-generation-v2").Valid() {
		t.Fatal("Capability.Valid accepted an unknown capability")
	}
}

func validMultimodalResource() ModelResource {
	return ModelResource{
		Identifier: slugkit.Slug("multi-model"),
		ModelID:    "provider/multi-model",
		Modalities: []Modality{ModalityChat, ModalityImage},
	}
}

func providerDefinitionForInvariantTest(key string, credentialKeys ...string) ProviderDefinition {
	fields := make([]CredentialField, len(credentialKeys))
	for i, credentialKey := range credentialKeys {
		fields[i] = CredentialField{Key: credentialKey, Label: credentialKey, Secret: true, Required: true}
	}
	return ProviderDefinition{
		Key:              slugkit.Slug(key),
		DisplayName:      key,
		Modalities:       []Modality{ModalityChat},
		CredentialFields: fields,
		ProtocolAdapter:  "test-adapter",
		ConnectionCheck:  bearerCheck("/models", credentialKeys[0]),
	}
}
