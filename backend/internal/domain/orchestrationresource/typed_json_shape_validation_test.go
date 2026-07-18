package orchestrationresource

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type registryNestedSpec struct {
	Name string `json:"name"`
}

type registryEmbeddedSpec struct {
	Embedded string `json:"embedded"`
	Ignored  string `json:"-"`
}

type registryShapeSpec struct {
	registryEmbeddedSpec
	Items   []registryNestedSpec          `json:"items"`
	Entries map[string]registryNestedSpec `json:"entries"`
	Opaque  json.RawMessage               `json:"opaque"`
	Number  any                           `json:"number"`
}

func registryForShape(t *testing.T, kind string, factory func() any) (*Registry, Manifest) {
	t.Helper()
	meta := TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: kind}
	registry := NewRegistry()
	require.NoError(t, registry.Register(meta, Schema{
		NewSpec:  factory,
		Validate: func(Metadata, any) error { return nil },
	}))
	manifest := validRegistryManifest()
	manifest.TypeMeta = meta
	return registry, manifest
}

func TestRegistryShapeValidationSupportsEmbeddingCollectionsAndUnmarshalers(t *testing.T) {
	registry, manifest := registryForShape(t, "RegistryShape", func() any {
		return &registryShapeSpec{}
	})
	manifest.Spec = json.RawMessage(`{
		"embedded":"value",
		"items":[{"name":"one"}],
		"entries":{"first":{"name":"two"}},
		"opaque":{"NonCanonicalField":true},
		"number":9007199254740993
	}`)

	decoded, err := registry.DecodeAndValidate(manifest)
	require.NoError(t, err)
	spec := decoded.(*registryShapeSpec)
	require.Equal(t, "value", spec.Embedded)
	require.Equal(t, "one", spec.Items[0].Name)
	require.Equal(t, "two", spec.Entries["first"].Name)
	require.JSONEq(t, `{"NonCanonicalField":true}`, string(spec.Opaque))
	require.Equal(t, json.Number("9007199254740993"), spec.Number)

	tests := []struct {
		name string
		spec string
		key  string
		path string
	}{
		{
			name: "embedded field case",
			spec: `{"Embedded":"value","items":[],"entries":{},"opaque":{},"number":1}`,
			key:  "Embedded",
			path: "spec",
		},
		{
			name: "slice element field case",
			spec: `{"embedded":"value","items":[{"Name":"one"}],"entries":{},` +
				`"opaque":{},"number":1}`,
			key:  "Name",
			path: "spec.items[0]",
		},
		{
			name: "map value field case",
			spec: `{"embedded":"value","items":[],"entries":{"first":{"Name":"two"}},` +
				`"opaque":{},"number":1}`,
			key:  "Name",
			path: "spec.entries[map value]",
		},
		{
			name: "ignored field",
			spec: `{"embedded":"value","Ignored":"secret","items":[],"entries":{},` +
				`"opaque":{},"number":1}`,
			key:  "Ignored",
			path: "spec",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest.Spec = json.RawMessage(tt.spec)
			_, err := registry.DecodeAndValidate(manifest)
			require.Error(t, err)
			require.ErrorIs(t, err, ErrTypedJSONUnknownField)
			require.Contains(t, err.Error(), "at path "+tt.path)
			require.NotContains(t, err.Error(), tt.key)
		})
	}
}

func TestRegistryShapeValidationRejectsUnsupportedMapKeyType(t *testing.T) {
	type unsupportedMapSpec struct {
		Entries map[int]registryNestedSpec `json:"entries"`
	}
	registry, manifest := registryForShape(t, "UnsupportedMap", func() any {
		return &unsupportedMapSpec{}
	})

	for _, spec := range []string{
		`{"entries":{"1":{"name":"valid"}}}`,
		`{"entries":{"1":{"Name":"case-variant"}}}`,
	} {
		manifest.Spec = json.RawMessage(spec)
		_, err := registry.DecodeAndValidate(manifest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported schema map key type int")
	}
}

func TestRegistryShapeValidationRejectsNullForTypedValues(t *testing.T) {
	registry, manifest := registryForShape(t, "NullShape", func() any {
		return &registryShapeSpec{}
	})
	tests := []struct {
		name string
		spec string
		path string
	}{
		{
			name: "primitive",
			spec: `{"embedded":null,"items":[],"entries":{},"opaque":{},"number":1}`,
			path: "spec.embedded",
		},
		{
			name: "array",
			spec: `{"embedded":"value","items":null,"entries":{},"opaque":{},"number":1}`,
			path: "spec.items",
		},
		{
			name: "map",
			spec: `{"embedded":"value","items":[],"entries":null,"opaque":{},"number":1}`,
			path: "spec.entries",
		},
		{
			name: "map value",
			spec: `{"embedded":"value","items":[],"entries":{"first":null},` +
				`"opaque":{},"number":1}`,
			path: "spec.entries[map value]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest.Spec = json.RawMessage(tt.spec)
			_, err := registry.DecodeAndValidate(manifest)
			require.ErrorIs(t, err, ErrTypedJSONType)
			require.Contains(t, err.Error(), "at path "+tt.path)
			require.NotContains(t, err.Error(), tt.spec)
		})
	}

	manifest.Spec = json.RawMessage(
		`{"embedded":"value","items":[],"entries":{},"opaque":null,"number":null}`,
	)
	_, err := registry.DecodeAndValidate(manifest)
	require.NoError(t, err)
}

func TestRegistryShapeValidationMatchesDashCommaJSONTag(t *testing.T) {
	type dashTagSpec struct {
		Dash    string `json:"-,"`
		Ignored string `json:"-"`
	}
	registry, manifest := registryForShape(t, "DashTag", func() any {
		return &dashTagSpec{}
	})
	manifest.Spec = json.RawMessage(`{"-":"accepted"}`)

	decoded, err := registry.DecodeAndValidate(manifest)
	require.NoError(t, err)
	require.Equal(t, "accepted", decoded.(*dashTagSpec).Dash)

	for _, key := range []string{"Dash", "-,", "Ignored"} {
		manifest.Spec = json.RawMessage(`{"` + key + `":"rejected"}`)
		_, err := registry.DecodeAndValidate(manifest)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrTypedJSONUnknownField)
		require.Contains(t, err.Error(), "at path spec")
		require.NotContains(t, err.Error(), key)
	}
}

func TestRegistryShapeValidationDoesNotExposeUnknownInputKey(t *testing.T) {
	registry, manifest := registryForShape(t, "UnknownInputKey", func() any {
		return &registryShapeSpec{}
	})

	for name, key := range map[string]string{
		"short": "sk_live_short_unknown",
		"long":  "sk_live_long_unknown_" + strings.Repeat("secret-", 200),
	} {
		t.Run(name, func(t *testing.T) {
			manifest.Spec = json.RawMessage(`{"` + key + `":1}`)

			_, err := registry.DecodeAndValidate(manifest)
			require.ErrorIs(t, err, ErrTypedJSONUnknownField)
			require.Contains(t, err.Error(), "at path spec")
			require.LessOrEqual(t, len(err.Error()), 512)
			requireInputKeyNotExposed(t, err.Error(), key)
		})
	}
}

func TestRegistryShapeValidationDoesNotExposeDynamicMapKey(t *testing.T) {
	registry, manifest := registryForShape(t, "DynamicMapKey", func() any {
		return &registryShapeSpec{}
	})
	const nestedUnknownKey = "sk_live_nested_unknown"

	for name, key := range map[string]string{
		"short": "sk_live_short_map_key",
		"long":  "sk_live_long_map_key_" + strings.Repeat("secret-", 200),
	} {
		t.Run(name, func(t *testing.T) {
			manifest.Spec = json.RawMessage(`{
				"embedded":"value",
				"items":[],
				"entries":{"` + key + `":{"` + nestedUnknownKey + `":"case-variant"}},
				"opaque":{},
				"number":1
			}`)

			_, err := registry.DecodeAndValidate(manifest)
			require.ErrorIs(t, err, ErrTypedJSONUnknownField)
			require.Contains(t, err.Error(), "at path spec.entries[map value]")
			require.LessOrEqual(t, len(err.Error()), 512)
			requireInputKeyNotExposed(t, err.Error(), key)
			requireInputKeyNotExposed(t, err.Error(), nestedUnknownKey)
		})
	}
}

func requireInputKeyNotExposed(t *testing.T, message, key string) {
	t.Helper()
	require.NotContains(t, message, key)
	prefix := []rune(key)
	if len(prefix) > 16 {
		prefix = prefix[:16]
	}
	require.NotContains(t, message, string(prefix))
}
