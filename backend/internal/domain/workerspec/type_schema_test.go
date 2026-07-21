package workerspec

import (
	"encoding/json"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTypeConfigAgainstSchemaAcceptsTypedValuesAndSecretRefs(t *testing.T) {
	config := TypeConfig{
		SchemaVersion: 1,
		Values: map[string]any{
			"enabled":      true,
			"model":        "gpt-5",
			"temperature":  json.Number("0.2"),
			"mode":         "careful",
			"token_budget": "20000",
		},
		SecretRefs: map[string]SecretReference{
			"api_token": {
				Kind: slugkit.MustNewForTest("vault-secret"),
				ID:   91,
			},
		},
		InteractionMode: InteractionModeACP,
		AutomationLevel: AutomationLevelAutonomous,
	}

	require.NoError(t, ValidateTypeConfigAgainstSchema(config, workerTypeSchemaForTest()))
}

func TestValidateTypeConfigAgainstSchemaRejectsInvalidAssignments(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*TypeConfig)
		match  string
	}{
		{"schema version", func(config *TypeConfig) {
			config.SchemaVersion = 2
		}, "schema version"},
		{"unknown value", func(config *TypeConfig) {
			config.Values["other"] = "value"
		}, "not declared"},
		{"secret in values", func(config *TypeConfig) {
			config.Values["api_token"] = "plaintext"
		}, "must use secret_refs"},
		{"ordinary secret ref", func(config *TypeConfig) {
			config.SecretRefs["model"] = SecretReference{
				Kind: slugkit.MustNewForTest("vault-secret"),
				ID:   92,
			}
		}, "does not accept secret_refs"},
		{"wrong boolean type", func(config *TypeConfig) {
			config.Values["enabled"] = "true"
		}, "must be boolean"},
		{"wrong number type", func(config *TypeConfig) {
			config.Values["temperature"] = "0.2"
		}, "must be number"},
		{"invalid select", func(config *TypeConfig) {
			config.Values["mode"] = "fallback"
		}, "invalid option"},
		{"nested object", func(config *TypeConfig) {
			config.Values["model"] = map[string]any{"id": "gpt-5"}
		}, "must be string"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := typeConfigForSchemaTest()
			test.mutate(&config)

			err := ValidateTypeConfigAgainstSchema(config, workerTypeSchemaForTest())
			require.Error(t, err)
			assert.ErrorContains(t, err, test.match)
		})
	}
}

func TestValidateTypeConfigAgainstSchemaRejectsMissingRequiredSecretRef(t *testing.T) {
	schema := workerTypeSchemaForTest()
	schema.Fields["api_token"] = TypeFieldSchema{
		Kind:     TypeFieldSecret,
		Required: true,
	}

	err := ValidateTypeConfigAgainstSchema(typeConfigForSchemaTest(), schema)

	require.Error(t, err)
	assert.ErrorContains(t, err, `secret ref "api_token" is required`)
}

func TestValidateTypeConfigAgainstSchemaRejectsMissingCredentialGroup(t *testing.T) {
	schema := workerTypeSchemaForTest()
	schema.Fields["anthropic_api_key"] = TypeFieldSchema{
		Kind: TypeFieldSecret,
	}
	schema.SecretRequirementGroups = []SecretRequirementGroup{{
		ID: "provider-api-key", AnyOf: []string{"api_token", "anthropic_api_key"},
	}}

	err := ValidateTypeConfigAgainstSchema(typeConfigForSchemaTest(), schema)

	require.Error(t, err)
	assert.ErrorContains(t, err, `credential group "provider-api-key"`)

	config := typeConfigForSchemaTest()
	config.SecretRefs["anthropic_api_key"] = SecretReference{
		Kind: slugkit.MustNewForTest("vault-secret"), ID: 92,
	}
	require.NoError(t, ValidateTypeConfigAgainstSchema(config, schema))
}

func typeConfigForSchemaTest() TypeConfig {
	return TypeConfig{
		SchemaVersion: 1,
		Values: map[string]any{
			"enabled":     true,
			"model":       "gpt-5",
			"temperature": json.Number("0.2"),
			"mode":        "careful",
		},
		SecretRefs: map[string]SecretReference{},
	}
}

func workerTypeSchemaForTest() TypeSchema {
	return TypeSchema{
		Version: 1,
		Fields: map[string]TypeFieldSchema{
			"enabled":      {Kind: TypeFieldBoolean},
			"model":        {Kind: TypeFieldString},
			"temperature":  {Kind: TypeFieldNumber},
			"mode":         {Kind: TypeFieldSelect, Options: []string{"fast", "careful"}},
			"token_budget": {Kind: TypeFieldString},
			"api_token":    {Kind: TypeFieldSecret},
		},
	}
}
