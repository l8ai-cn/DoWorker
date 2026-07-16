package orchestrationresource

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type cacheCycleSpec struct {
	Value string `json:"value"`
	*cacheCycleSpec
}

type cacheDiamondBase struct {
	Value string `json:"value"`
}

type cacheDiamondLeft struct {
	cacheDiamondBase
}

type cacheDiamondRight struct {
	cacheDiamondBase
}

type cacheDiamondDominantSpec struct {
	cacheDiamondLeft
	cacheDiamondRight
	Value string `json:"value"`
}

type cacheDiamondAmbiguousSpec struct {
	cacheDiamondLeft
	cacheDiamondRight
}

type cacheArrayItem struct {
	Name string `json:"name"`
}

type cacheArraySpec struct {
	Items []cacheArrayItem `json:"items"`
}

func TestJSONFieldCacheMatchesEncodingJSONCycleAndDiamondRules(t *testing.T) {
	t.Run("cycle", func(t *testing.T) {
		source := []byte(`{"value":"accepted"}`)
		var standard cacheCycleSpec
		requireEncodingJSONDecode(t, source, &standard)

		registry, manifest := registryForShape(t, "CacheCycle", func() any {
			return &cacheCycleSpec{}
		})
		manifest.Spec = source
		decoded, err := registry.DecodeAndValidate(manifest)
		require.NoError(t, err)
		require.Equal(t, standard.Value, decoded.(*cacheCycleSpec).Value)
	})

	t.Run("diamond direct field dominates", func(t *testing.T) {
		source := []byte(`{"value":"accepted"}`)
		var standard cacheDiamondDominantSpec
		requireEncodingJSONDecode(t, source, &standard)

		registry, manifest := registryForShape(t, "CacheDiamondDominant", func() any {
			return &cacheDiamondDominantSpec{}
		})
		manifest.Spec = source
		decoded, err := registry.DecodeAndValidate(manifest)
		require.NoError(t, err)
		require.Equal(t, standard, *decoded.(*cacheDiamondDominantSpec))
	})

	t.Run("diamond remains ambiguous", func(t *testing.T) {
		source := []byte(`{"value":"rejected"}`)
		var standard cacheDiamondAmbiguousSpec
		decoder := json.NewDecoder(bytes.NewReader(source))
		decoder.DisallowUnknownFields()
		require.Error(t, decoder.Decode(&standard))

		registry, manifest := registryForShape(t, "CacheDiamondAmbiguous", func() any {
			return &cacheDiamondAmbiguousSpec{}
		})
		manifest.Spec = source
		_, err := registry.DecodeAndValidate(manifest)
		require.ErrorIs(t, err, ErrTypedJSONUnknownField)
		require.Contains(t, err.Error(), "at path spec")
		require.NotContains(t, err.Error(), "value")
	})
}

func TestJSONFieldCacheHandlesHighCardinalityArray(t *testing.T) {
	const itemCount = 2_000
	input := cacheArraySpec{Items: make([]cacheArrayItem, itemCount)}
	for index := range input.Items {
		input.Items[index].Name = "accepted"
	}
	source, err := json.Marshal(input)
	require.NoError(t, err)

	var standard cacheArraySpec
	requireEncodingJSONDecode(t, source, &standard)
	registry, manifest := registryForShape(t, "CacheArray", func() any {
		return &cacheArraySpec{}
	})
	manifest.Spec = source
	decoded, err := registry.DecodeAndValidate(manifest)
	require.NoError(t, err)
	require.Equal(t, standard, *decoded.(*cacheArraySpec))

	cached, exists := jsonFieldCache.Load(reflect.TypeOf(cacheArrayItem{}))
	require.True(t, exists)
	fields := cached.(map[string]reflect.Type)
	require.Equal(t, reflect.TypeOf(""), fields["name"])
}
