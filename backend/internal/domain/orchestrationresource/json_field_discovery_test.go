package orchestrationresource

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	promotionLevel00 struct {
		Promoted string `json:"promoted"`
	}
	promotionLevel01 struct{ promotionLevel00 }
	promotionLevel02 struct{ promotionLevel01 }
	promotionLevel03 struct{ promotionLevel02 }
	promotionLevel04 struct{ promotionLevel03 }
	promotionLevel05 struct{ promotionLevel04 }
	promotionLevel06 struct{ promotionLevel05 }
	promotionLevel07 struct{ promotionLevel06 }
	promotionLevel08 struct{ promotionLevel07 }
	promotionLevel09 struct{ promotionLevel08 }
	promotionLevel10 struct{ promotionLevel09 }
	promotionLevel11 struct{ promotionLevel10 }
	promotionLevel12 struct{ promotionLevel11 }
	promotionLevel13 struct{ promotionLevel12 }
	promotionLevel14 struct{ promotionLevel13 }
	promotionLevel15 struct{ promotionLevel14 }
	promotionLevel16 struct{ promotionLevel15 }
	promotionLevel17 struct{ promotionLevel16 }
	promotionLevel18 struct{ promotionLevel17 }
	promotionLevel19 struct{ promotionLevel18 }
	promotionLevel20 struct{ promotionLevel19 }
	promotionLevel21 struct{ promotionLevel20 }
	promotionLevel22 struct{ promotionLevel21 }
	promotionLevel23 struct{ promotionLevel22 }
	promotionLevel24 struct{ promotionLevel23 }
	promotionLevel25 struct{ promotionLevel24 }
	promotionLevel26 struct{ promotionLevel25 }
	promotionLevel27 struct{ promotionLevel26 }
	promotionLevel28 struct{ promotionLevel27 }
	promotionLevel29 struct{ promotionLevel28 }
	promotionLevel30 struct{ promotionLevel29 }
	promotionLevel31 struct{ promotionLevel30 }
	promotionLevel32 struct{ promotionLevel31 }
	promotionLevel33 struct{ promotionLevel32 }
	promotionLevel34 struct{ promotionLevel33 }
	promotionLevel35 struct{ promotionLevel34 }
	promotionLevel36 struct{ promotionLevel35 }
	promotionLevel37 struct{ promotionLevel36 }
	promotionLevel38 struct{ promotionLevel37 }
	promotionLevel39 struct{ promotionLevel38 }
	promotionLevel40 struct{ promotionLevel39 }
	promotionLevel41 struct{ promotionLevel40 }
	promotionLevel42 struct{ promotionLevel41 }
	promotionLevel43 struct{ promotionLevel42 }
	promotionLevel44 struct{ promotionLevel43 }
	promotionLevel45 struct{ promotionLevel44 }
	promotionLevel46 struct{ promotionLevel45 }
	promotionLevel47 struct{ promotionLevel46 }
	promotionLevel48 struct{ promotionLevel47 }
	promotionLevel49 struct{ promotionLevel48 }
	promotionLevel50 struct{ promotionLevel49 }
	promotionLevel51 struct{ promotionLevel50 }
	promotionLevel52 struct{ promotionLevel51 }
	promotionLevel53 struct{ promotionLevel52 }
	promotionLevel54 struct{ promotionLevel53 }
	promotionLevel55 struct{ promotionLevel54 }
	promotionLevel56 struct{ promotionLevel55 }
	promotionLevel57 struct{ promotionLevel56 }
	promotionLevel58 struct{ promotionLevel57 }
	promotionLevel59 struct{ promotionLevel58 }
	promotionLevel60 struct{ promotionLevel59 }
	promotionLevel61 struct{ promotionLevel60 }
	promotionLevel62 struct{ promotionLevel61 }
	promotionLevel63 struct{ promotionLevel62 }
	promotionLevel64 struct{ promotionLevel63 }
	promotionLevel65 struct{ promotionLevel64 }
	promotionLevel66 struct{ promotionLevel65 }
)

func requireEncodingJSONDecode(t *testing.T, source []byte, target any) {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(target))
	require.NoError(t, requireJSONEOF(decoder))
}

func TestJSONFieldDiscoveryMatchesEncodingJSONInvalidTagFallback(t *testing.T) {
	t.Run("ordinary field falls back to Go name", func(t *testing.T) {
		type invalidTagSpec struct {
			FieldName string `json:"invalid\\name"`
		}
		source := []byte(`{"FieldName":"accepted"}`)
		var standard invalidTagSpec
		requireEncodingJSONDecode(t, source, &standard)
		require.Equal(t, "accepted", standard.FieldName)

		registry, manifest := registryForShape(t, "InvalidTag", func() any {
			return &invalidTagSpec{}
		})
		manifest.Spec = source
		decoded, err := registry.DecodeAndValidate(manifest)
		require.NoError(t, err)
		require.Equal(t, standard, *decoded.(*invalidTagSpec))
	})

	t.Run("anonymous struct falls back to promotion", func(t *testing.T) {
		type invalidTagFields struct {
			Promoted string `json:"promoted"`
		}
		type invalidTagSpec struct {
			invalidTagFields `json:"invalid\\name"`
		}
		source := []byte(`{"promoted":"accepted"}`)
		var standard invalidTagSpec
		requireEncodingJSONDecode(t, source, &standard)

		registry, manifest := registryForShape(t, "InvalidAnonymousTag", func() any {
			return &invalidTagSpec{}
		})
		manifest.Spec = source
		decoded, err := registry.DecodeAndValidate(manifest)
		require.NoError(t, err)
		require.Equal(t, standard, *decoded.(*invalidTagSpec))
	})
}

func TestJSONFieldDiscoveryMatchesTaggedUnexportedAnonymousStruct(t *testing.T) {
	type privateFields struct {
		Value string `json:"value"`
	}
	type taggedPrivateSpec struct {
		privateFields `json:"private"`
	}
	source := []byte(`{"private":{"value":"accepted"}}`)
	var standard taggedPrivateSpec
	requireEncodingJSONDecode(t, source, &standard)
	require.Equal(t, "accepted", standard.Value)

	registry, manifest := registryForShape(t, "TaggedPrivate", func() any {
		return &taggedPrivateSpec{}
	})
	manifest.Spec = source
	decoded, err := registry.DecodeAndValidate(manifest)
	require.NoError(t, err)
	require.Equal(t, standard, *decoded.(*taggedPrivateSpec))
}

func TestJSONFieldDiscoveryAllowsMoreThanSixtyFivePromotionLevels(t *testing.T) {
	source := []byte(`{"promoted":"accepted"}`)
	var standard promotionLevel66
	requireEncodingJSONDecode(t, source, &standard)
	require.Equal(t, "accepted", standard.Promoted)

	registry, manifest := registryForShape(t, "DeepPromotion", func() any {
		return &promotionLevel66{}
	})
	manifest.Spec = source
	decoded, err := registry.DecodeAndValidate(manifest)
	require.NoError(t, err)
	require.Equal(t, standard, *decoded.(*promotionLevel66))
}
