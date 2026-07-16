package orchestrationresource

import (
	"encoding/json"
	"testing"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

type rootValueUnmarshaler struct {
	Value string `json:"value"`
}

func (rootValueUnmarshaler) UnmarshalJSON([]byte) error {
	return nil
}

type rootPointerUnmarshaler struct {
	Value string `json:"value"`
}

func (*rootPointerUnmarshaler) UnmarshalJSON([]byte) error {
	return nil
}

type namedRegistrySpecPointer *registrySpec

func TestRegisterRejectsInvalidRootSchemaAtRegistration(t *testing.T) {
	tests := []struct {
		name    string
		factory func() any
		message string
	}{
		{
			name:    "nil",
			factory: func() any { return nil },
			message: "non-nil pointer to struct",
		},
		{
			name:    "typed nil",
			factory: func() any { return (*registrySpec)(nil) },
			message: "non-nil pointer to struct",
		},
		{
			name:    "non pointer",
			factory: func() any { return registrySpec{} },
			message: "non-nil pointer to struct",
		},
		{
			name: "pointer to map",
			factory: func() any {
				value := map[string]registryNestedSpec{}
				return &value
			},
			message: "non-nil pointer to struct",
		},
		{
			name: "pointer to interface",
			factory: func() any {
				var value any
				return &value
			},
			message: "non-nil pointer to struct",
		},
		{
			name:    "pointer to raw message",
			factory: func() any { return new(json.RawMessage) },
			message: "non-nil pointer to struct",
		},
		{
			name: "named pointer type",
			factory: func() any {
				return namedRegistrySpecPointer(&registrySpec{})
			},
			message: "ordinary pointer to struct",
		},
		{
			name:    "root value unmarshaler",
			factory: func() any { return &rootValueUnmarshaler{} },
			message: "must not implement json.Unmarshaler",
		},
		{
			name:    "root pointer unmarshaler",
			factory: func() any { return &rootPointerUnmarshaler{} },
			message: "must not implement json.Unmarshaler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRegistry().Register(registryMeta, Schema{
				NewSpec:  tt.factory,
				Validate: func(Metadata, any) error { return nil },
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.message)
		})
	}
}

func TestRegisterFreezesRootTypeAndDecodeCreatesFreshInstances(t *testing.T) {
	calls := 0
	shared := &registrySpec{}
	registry := NewRegistry()
	require.NoError(t, registry.Register(registryMeta, Schema{
		NewSpec: func() any {
			calls++
			return shared
		},
		Validate: validatingRegistrySchema().Validate,
	}))
	require.Equal(t, 1, calls)

	firstValue, err := registry.DecodeAndValidate(validRegistryManifest())
	require.NoError(t, err)
	secondValue, err := registry.DecodeAndValidate(validRegistryManifest())
	require.NoError(t, err)
	require.Equal(t, 1, calls)

	first := firstValue.(*registrySpec)
	second := secondValue.(*registrySpec)
	require.NotSame(t, shared, first)
	require.NotSame(t, first, second)
	first.ModelRef.Name = slugkit.MustNewForTest("changed-model")
	require.Equal(t, slugkit.Slug("model-one"), second.ModelRef.Name)
}
