package orchestrationresource

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

type registrySpec struct {
	ModelRef      Reference       `json:"modelRef"`
	CredentialRef SecretReference `json:"credentialRef"`
}

var registryMeta = TypeMeta{
	APIVersion: APIVersionV1Alpha1,
	Kind:       "RegistryResource",
}

func validRegistryManifest() Manifest {
	return Manifest{
		TypeMeta: registryMeta,
		Metadata: Metadata{
			Name:      slugkit.MustNewForTest("registry-resource"),
			Namespace: slugkit.MustNewForTest("team-alpha"),
		},
		Spec: json.RawMessage(`{
			"modelRef":{"kind":"Model","name":"model-one"},
			"credentialRef":{"name":"model-secret","key":"api-key"}
		}`),
	}
}

func validatingRegistrySchema() Schema {
	return Schema{
		NewSpec: func() any { return &registrySpec{} },
		Validate: func(metadata Metadata, value any) error {
			spec := value.(*registrySpec)
			if err := spec.ModelRef.ValidateDraft(metadata.Namespace.String()); err != nil {
				return fmt.Errorf("modelRef: %w", err)
			}
			if err := spec.CredentialRef.Validate(); err != nil {
				return fmt.Errorf("credentialRef: %w", err)
			}
			return nil
		},
	}
}

func TestRegistryRegistersDecodesAndPassesMetadataToValidator(t *testing.T) {
	registry := NewRegistry()
	var received Metadata
	schema := validatingRegistrySchema()
	originalValidate := schema.Validate
	schema.Validate = func(metadata Metadata, value any) error {
		received = metadata
		return originalValidate(metadata, value)
	}

	require.NoError(t, registry.Register(registryMeta, schema))
	decoded, err := registry.DecodeAndValidate(validRegistryManifest())
	require.NoError(t, err)
	require.Equal(t, validRegistryManifest().Metadata, received)

	spec, ok := decoded.(*registrySpec)
	require.True(t, ok)
	require.Equal(t, slugkit.Slug("model-one"), spec.ModelRef.Name)
	require.Equal(t, slugkit.Slug("model-secret"), spec.CredentialRef.Name)
}

func TestRegistryRegisterRejectsDuplicateInvalidMetaAndNilCallbacks(t *testing.T) {
	t.Run("duplicate", func(t *testing.T) {
		registry := NewRegistry()
		require.NoError(t, registry.Register(registryMeta, validatingRegistrySchema()))
		err := registry.Register(registryMeta, validatingRegistrySchema())
		require.ErrorIs(t, err, ErrDuplicateSchema)
	})

	t.Run("invalid meta", func(t *testing.T) {
		err := NewRegistry().Register(
			TypeMeta{APIVersion: "v1", Kind: "RegistryResource"},
			validatingRegistrySchema(),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "typeMeta.APIVersion")
	})

	tests := []struct {
		name   string
		schema Schema
	}{
		{name: "nil factory", schema: Schema{Validate: func(Metadata, any) error { return nil }}},
		{name: "nil validator", schema: Schema{NewSpec: func() any { return &registrySpec{} }}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, NewRegistry().Register(registryMeta, tt.schema))
		})
	}
}

func TestRegistryDecodeRejectsUnknownTypeAfterStoredValidation(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, registry.Register(registryMeta, validatingRegistrySchema()))

	t.Run("unknown kind", func(t *testing.T) {
		manifest := validRegistryManifest()
		manifest.Kind = "UnknownResource"
		_, err := registry.DecodeAndValidate(manifest)
		require.ErrorIs(t, err, ErrUnknownSchema)
	})

	t.Run("invalid version is rejected first", func(t *testing.T) {
		manifest := validRegistryManifest()
		manifest.APIVersion = "agentcloud.io/v2"
		_, err := registry.DecodeAndValidate(manifest)
		require.Error(t, err)
		require.NotErrorIs(t, err, ErrUnknownSchema)
		require.Contains(t, err.Error(), "typeMeta.APIVersion")
	})

	t.Run("invalid stored metadata is rejected first", func(t *testing.T) {
		manifest := validRegistryManifest()
		manifest.Kind = "UnknownResource"
		manifest.Metadata.Namespace = ""
		_, err := registry.DecodeAndValidate(manifest)
		require.Error(t, err)
		require.NotErrorIs(t, err, ErrUnknownSchema)
		require.Contains(t, err.Error(), "metadata.namespace")
	})
}

func TestRegistryDecodeRejectsUnknownAndNonCanonicalFields(t *testing.T) {
	tests := []struct {
		name string
		spec string
		key  string
		path string
	}{
		{
			name: "unknown top-level field is deterministic",
			spec: `{"zzz":1,"aaa":2,"modelRef":{"kind":"Model","name":"model-one"},` +
				`"credentialRef":{"name":"model-secret","key":"api-key"}}`,
			key:  "aaa",
			path: "spec",
		},
		{
			name: "top-level case variant",
			spec: `{"ModelRef":{"kind":"Model","name":"model-one"},` +
				`"credentialRef":{"name":"model-secret","key":"api-key"}}`,
			key:  "ModelRef",
			path: "spec",
		},
		{
			name: "nested reference case variant",
			spec: `{"modelRef":{"Kind":"Model","name":"model-one"},` +
				`"credentialRef":{"name":"model-secret","key":"api-key"}}`,
			key:  "Kind",
			path: "spec.modelRef",
		},
		{
			name: "secret value",
			spec: `{"modelRef":{"kind":"Model","name":"model-one"},` +
				`"credentialRef":{"name":"model-secret","key":"api-key","value":"secret"}}`,
			key:  "value",
			path: "spec.credentialRef",
		},
		{
			name: "secret Value",
			spec: `{"modelRef":{"kind":"Model","name":"model-one"},` +
				`"credentialRef":{"name":"model-secret","key":"api-key","Value":"secret"}}`,
			key:  "Value",
			path: "spec.credentialRef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()
			require.NoError(t, registry.Register(registryMeta, validatingRegistrySchema()))
			manifest := validRegistryManifest()
			manifest.Spec = json.RawMessage(tt.spec)

			_, err := registry.DecodeAndValidate(manifest)
			require.Error(t, err)
			require.ErrorIs(t, err, ErrTypedJSONUnknownField)
			require.Contains(t, err.Error(), "unknown field")
			require.Contains(t, err.Error(), "at path "+tt.path)
			require.NotContains(t, err.Error(), tt.key)
		})
	}
}

func TestRegistryDecodeRejectsMalformedSpecStructure(t *testing.T) {
	deep := strings.Repeat(`{"items":`, 65) + `null` + strings.Repeat(`}`, 65)
	tests := []struct {
		name string
		spec []byte
		msg  string
	}{
		{
			name: "duplicate nested key",
			spec: []byte(`{"modelRef":{"kind":"Model","kind":"Other","name":"model-one"},` +
				`"credentialRef":{"name":"model-secret","key":"api-key"}}`),
			msg: "duplicate JSON key",
		},
		{name: "invalid UTF-8", spec: []byte("{\"modelRef\":\"\xff\"}"), msg: "valid UTF-8"},
		{name: "excessive depth", spec: []byte(deep), msg: "maximum depth"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()
			require.NoError(t, registry.Register(registryMeta, validatingRegistrySchema()))
			manifest := validRegistryManifest()
			manifest.Spec = json.RawMessage(tt.spec)
			_, err := registry.DecodeAndValidate(manifest)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.msg)
		})
	}
}

func TestRegistryDecodePreservesValidationErrors(t *testing.T) {
	t.Run("cross namespace reference", func(t *testing.T) {
		registry := NewRegistry()
		require.NoError(t, registry.Register(registryMeta, validatingRegistrySchema()))
		manifest := validRegistryManifest()
		manifest.Spec = json.RawMessage(`{
			"modelRef":{"kind":"Model","namespace":"team-beta","name":"model-one"},
			"credentialRef":{"name":"model-secret","key":"api-key"}
		}`)
		_, err := registry.DecodeAndValidate(manifest)
		require.ErrorIs(t, err, ErrCrossNamespaceReference)
	})

	t.Run("schema validator sentinel", func(t *testing.T) {
		sentinel := errors.New("validator failed")
		registry := NewRegistry()
		require.NoError(t, registry.Register(registryMeta, Schema{
			NewSpec: func() any { return &registrySpec{} },
			Validate: func(Metadata, any) error {
				return fmt.Errorf("validation detail: %w", sentinel)
			},
		}))
		_, err := registry.DecodeAndValidate(validRegistryManifest())
		require.ErrorIs(t, err, sentinel)
	})
}

func TestRegistryDecodeDoesNotMutateManifest(t *testing.T) {
	registry := NewRegistry()
	schema := validatingRegistrySchema()
	originalValidate := schema.Validate
	schema.Validate = func(metadata Metadata, value any) error {
		metadata.Labels["mutated"] = "yes"
		return originalValidate(metadata, value)
	}
	require.NoError(t, registry.Register(registryMeta, schema))
	manifest := validRegistryManifest()
	manifest.Metadata.Labels = map[string]string{"role": "builder"}
	beforeMetadata := manifest.Metadata
	beforeMetadata.Labels = maps.Clone(manifest.Metadata.Labels)
	beforeSpec := append([]byte(nil), manifest.Spec...)

	_, err := registry.DecodeAndValidate(manifest)
	require.NoError(t, err)
	require.Equal(t, beforeMetadata, manifest.Metadata)
	require.Equal(t, beforeSpec, []byte(manifest.Spec))
}
