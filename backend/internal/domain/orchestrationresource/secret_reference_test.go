package orchestrationresource

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func validSecretReferenceForTest() SecretReference {
	return SecretReference{
		Name: slugkit.MustNewForTest("registry-credentials"),
		Key:  slugkit.MustNewForTest("access-token"),
	}
}

func TestSecretReferenceShapeDoesNotContainSecretValue(t *testing.T) {
	secretReferenceType := reflect.TypeOf(SecretReference{})
	expectedFields := map[string]struct {
		goType  reflect.Type
		jsonTag string
		yamlTag string
	}{
		"Name": {
			goType:  reflect.TypeOf(slugkit.Slug("")),
			jsonTag: "name",
			yamlTag: "name",
		},
		"Key": {
			goType:  reflect.TypeOf(slugkit.Slug("")),
			jsonTag: "key",
			yamlTag: "key",
		},
		"Revision": {
			goType:  reflect.TypeOf(int64(0)),
			jsonTag: "revision,omitempty",
			yamlTag: "revision,omitempty",
		},
	}

	require.Equal(t, len(expectedFields), secretReferenceType.NumField())
	for name, expected := range expectedFields {
		field, found := secretReferenceType.FieldByName(name)
		require.True(t, found, "missing field %s", name)
		require.Equal(t, expected.goType, field.Type)
		require.Equal(t, expected.jsonTag, field.Tag.Get("json"))
		require.Equal(t, expected.yamlTag, field.Tag.Get("yaml"))
	}
}

func TestSecretReferenceSerializationContainsOnlyReferenceFields(t *testing.T) {
	ref := validSecretReferenceForTest()
	ref.Revision = 7

	jsonData, err := json.Marshal(ref)
	require.NoError(t, err)
	var jsonObject map[string]any
	require.NoError(t, json.Unmarshal(jsonData, &jsonObject))
	require.Equal(t, map[string]any{
		"name":     "registry-credentials",
		"key":      "access-token",
		"revision": float64(7),
	}, jsonObject)

	yamlData, err := yaml.Marshal(ref)
	require.NoError(t, err)
	var yamlObject map[string]any
	require.NoError(t, yaml.Unmarshal(yamlData, &yamlObject))
	require.Equal(t, map[string]any{
		"name":     "registry-credentials",
		"key":      "access-token",
		"revision": 7,
	}, yamlObject)
}

func TestSecretReferenceValidateAcceptsUnpinnedAndPositiveRevision(t *testing.T) {
	tests := []struct {
		name     string
		revision int64
	}{
		{name: "unpinned", revision: 0},
		{name: "pinned", revision: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := validSecretReferenceForTest()
			ref.Revision = tt.revision

			require.NoError(t, ref.Validate())
		})
	}
}

func TestSecretReferenceValidateRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
		mutate  func(*SecretReference)
	}{
		{
			name:    "invalid name",
			path:    "secretReference.name",
			wantErr: slugkit.ErrInvalidFormat,
			mutate:  func(ref *SecretReference) { ref.Name = "Registry_Credentials" },
		},
		{
			name:    "invalid key",
			path:    "secretReference.key",
			wantErr: slugkit.ErrInvalidFormat,
			mutate:  func(ref *SecretReference) { ref.Key = "access.token" },
		},
		{
			name:   "negative revision",
			path:   "secretReference.revision",
			mutate: func(ref *SecretReference) { ref.Revision = -1 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := validSecretReferenceForTest()
			tt.mutate(&ref)

			err := ref.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.path)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
