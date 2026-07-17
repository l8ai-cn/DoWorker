package orchestrationresource

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const maxExpectedPublicTypedJSONErrorBytes = 512

type boundedDecodeSpec struct {
	Count   int64 `json:"count"`
	Encoded int64 `json:"encoded,string"`
	Plain   int64 `json:"plain"`
}

type leakingNestedValue struct{}

func (*leakingNestedValue) UnmarshalJSON([]byte) error {
	return errors.New(
		"sk_live_generic_secret_" +
			strings.Repeat("generic-sensitive-", 30_000) +
			"GENERIC_SECRET_TAIL",
	)
}

type genericDecodeErrorSpec struct {
	Value leakingNestedValue `json:"value"`
}

func boundedDecodeRegistry(t *testing.T) (*Registry, Manifest) {
	t.Helper()
	meta := TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: "BoundedDecode"}
	registry := NewRegistry()
	require.NoError(t, registry.Register(meta, Schema{
		NewSpec:  func() any { return &boundedDecodeSpec{} },
		Validate: func(Metadata, any) error { return nil },
	}))
	manifest := validRegistryManifest()
	manifest.TypeMeta = meta
	return registry, manifest
}

func requireBoundedTypedJSONError(t *testing.T, err error) string {
	t.Helper()
	require.Error(t, err)
	message := err.Error()
	require.LessOrEqual(t, len(message), maxExpectedPublicTypedJSONErrorBytes)
	return message
}

func TestTypedJSONDecodeBoundsLargeIntegerOverflowError(t *testing.T) {
	registry, manifest := boundedDecodeRegistry(t)
	const secretDigits = "314159265358979323846"
	largeInteger := secretDigits + strings.Repeat("9", 899_000)
	manifest.Spec = json.RawMessage(`{"count":` + largeInteger + `,"encoded":"1","plain":1}`)

	_, err := registry.DecodeAndValidate(manifest)
	message := requireBoundedTypedJSONError(t, err)
	require.ErrorIs(t, err, ErrTypedJSONType)
	require.Contains(t, message, "typed JSON type error")
	require.Contains(t, message, summarizeValue("count"))
	require.Contains(t, message, summarizeValue("int64"))
	require.Contains(t, message, "offset")
	require.NotContains(t, message, secretDigits)
	require.NotContains(t, message, strings.Repeat("9", 200))
	var originalTypeError *json.UnmarshalTypeError
	require.False(t, errors.As(err, &originalTypeError))
}

func TestTypedJSONDecodeBoundsLargeStringTagPayload(t *testing.T) {
	tests := []struct {
		name       string
		payload    string
		secret     string
		tailMarker string
	}{
		{
			name:    "short secret",
			payload: "sk_live_short_secret",
			secret:  "sk_live_short_secret",
		},
		{
			name:       "long secret and escaped payload",
			payload:    "sk_live_long_secret_" + strings.Repeat(`\"`, 120_000) + "STRING_TAG_SECRET_TAIL",
			secret:     "sk_live_long_secret_",
			tailMarker: "STRING_TAG_SECRET_TAIL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, manifest := boundedDecodeRegistry(t)
			encodedPayload, err := json.Marshal(tt.payload)
			require.NoError(t, err)
			manifest.Spec = json.RawMessage(
				`{"count":1,"encoded":` + string(encodedPayload) + `,"plain":1}`,
			)

			_, err = registry.DecodeAndValidate(manifest)
			message := requireBoundedTypedJSONError(t, err)
			require.ErrorIs(t, err, ErrTypedJSONStringTag)
			require.Contains(t, message, "typed JSON string-tag error")
			require.NotContains(t, message, "offset")
			require.NotContains(t, message, tt.secret)
			if tt.tailMarker != "" {
				require.NotContains(t, message, tt.tailMarker)
			}
		})
	}
}

func TestTypedJSONDecodeBoundsUnknownFieldAndGeneralTypeErrors(t *testing.T) {
	t.Run("unknown field", func(t *testing.T) {
		const tailMarker = "UNKNOWN_FIELD_SECRET_TAIL"
		largeKey := strings.Repeat("unknown-", 40_000) + tailMarker
		registry, manifest := boundedDecodeRegistry(t)
		manifest.Spec = json.RawMessage(`{"` + largeKey + `":1}`)

		_, err := registry.DecodeAndValidate(manifest)
		message := requireBoundedTypedJSONError(t, err)
		require.ErrorIs(t, err, ErrTypedJSONUnknownField)
		require.Contains(t, message, "typed JSON unknown field")
		require.NotContains(t, message, "offset")
		require.NotContains(t, message, tailMarker)
		require.NotContains(t, message, strings.Repeat("unknown-", 100))
	})

	t.Run("ordinary type error", func(t *testing.T) {
		registry, manifest := boundedDecodeRegistry(t)
		const sensitiveValue = "ordinary-type-sensitive-value"
		manifest.Spec = json.RawMessage(
			`{"count":1,"encoded":"1","plain":"` + sensitiveValue + `"}`,
		)

		_, err := registry.DecodeAndValidate(manifest)
		message := requireBoundedTypedJSONError(t, err)
		require.ErrorIs(t, err, ErrTypedJSONType)
		require.Contains(t, message, "typed JSON type error")
		require.Contains(t, message, summarizeValue("plain"))
		require.Contains(t, message, summarizeValue("int64"))
		require.Contains(t, message, "offset")
		require.NotContains(t, message, sensitiveValue)
	})

	t.Run("generic nested unmarshaler error", func(t *testing.T) {
		meta := TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: "GenericDecodeError"}
		registry := NewRegistry()
		require.NoError(t, registry.Register(meta, Schema{
			NewSpec:  func() any { return &genericDecodeErrorSpec{} },
			Validate: func(Metadata, any) error { return nil },
		}))
		manifest := validRegistryManifest()
		manifest.TypeMeta = meta
		manifest.Spec = json.RawMessage(`{"value":{}}`)

		_, err := registry.DecodeAndValidate(manifest)
		message := requireBoundedTypedJSONError(t, err)
		require.ErrorIs(t, err, ErrTypedJSONDecode)
		require.Contains(t, message, "typed JSON decode error")
		require.NotContains(t, message, "offset")
		require.NotContains(t, message, "sk_live_generic_secret_")
		require.NotContains(t, message, "GENERIC_SECRET_TAIL")
		require.NotContains(t, message, strings.Repeat("generic-sensitive-", 100))
	})
}

func TestTypedJSONDecodeClassifiesSyntaxAndInvalidTargetErrors(t *testing.T) {
	t.Run("syntax error has reliable offset", func(t *testing.T) {
		target := boundedDecodeSpec{}
		err := decodeTypedJSON([]byte(`{"count":!}`), &target)
		message := requireBoundedTypedJSONError(t, err)
		require.ErrorIs(t, err, ErrTypedJSONSyntax)
		require.Contains(t, message, "typed JSON syntax error")
		require.Contains(t, message, "offset")
	})

	t.Run("unexpected EOF has no fabricated offset", func(t *testing.T) {
		target := boundedDecodeSpec{}
		err := decodeTypedJSON([]byte(`{"count":`), &target)
		message := requireBoundedTypedJSONError(t, err)
		require.ErrorIs(t, err, ErrTypedJSONSyntax)
		require.Contains(t, message, "typed JSON syntax error")
		require.NotContains(t, message, "offset")
	})

	t.Run("invalid target", func(t *testing.T) {
		err := decodeTypedJSON([]byte(`{}`), boundedDecodeSpec{})
		message := requireBoundedTypedJSONError(t, err)
		require.ErrorIs(t, err, ErrTypedJSONInvalidTarget)
		require.Contains(t, message, "typed JSON internal target error")
		require.Contains(t, message, summarizeValue("orchestrationresource.boundedDecodeSpec"))
		require.NotContains(t, message, "offset")
	})
}
