package orchestrationresource

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeYAMLSubmissionValidatesExplicitNumericTags(t *testing.T) {
	valid := map[string]string{
		"integer":           "!!int 1",
		"decimal float":     "!!float 1.0",
		"exponential float": "!!float 1e3",
	}
	for name, scalar := range valid {
		t.Run(name, func(t *testing.T) {
			source := strings.Replace(validSubmissionYAML, "  runtime: codex",
				"  runtime: codex\n  value: "+scalar, 1)
			_, err := DecodeYAMLSubmission([]byte(source))
			require.NoError(t, err)
		})
	}

	invalid := map[string]string{
		"decimal int":     "!!int 1.0",
		"exponential int": "!!int 1e3",
		"integer float":   "!!float 1",
	}
	for name, scalar := range invalid {
		t.Run(name, func(t *testing.T) {
			source := strings.Replace(validSubmissionYAML, "  runtime: codex",
				"  runtime: codex\n  value: "+scalar, 1)
			_, err := DecodeYAMLSubmission([]byte(source))
			require.Error(t, err)
			require.Contains(t, err.Error(), "JSON")
		})
	}
}

func TestDecodeYAMLSubmissionPreservesExtremeJSONNumbers(t *testing.T) {
	for _, number := range []string{"1e9999", "18446744073709551616"} {
		t.Run(number, func(t *testing.T) {
			source := strings.Replace(validSubmissionYAML, "  runtime: codex",
				"  runtime: codex\n  value: "+number, 1)

			manifest, err := DecodeYAMLSubmission([]byte(source))

			require.NoError(t, err)
			require.Contains(t, string(manifest.Spec), `"value":`+number)
		})
	}
}

func TestDecodeYAMLSubmissionPreservesExplicitStrings(t *testing.T) {
	tests := map[string]struct {
		scalar string
		value  string
	}{
		"quoted":    {scalar: `"1e9999"`, value: "1e9999"},
		"core tag":  {scalar: `!!str 18446744073709551616`, value: "18446744073709551616"},
		"long form": {scalar: `!<tag:yaml.org,2002:str> 0x10`, value: "0x10"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			source := strings.Replace(validSubmissionYAML, "  runtime: codex",
				"  runtime: codex\n  value: "+tt.scalar, 1)

			manifest, err := DecodeYAMLSubmission([]byte(source))

			require.NoError(t, err)
			require.JSONEq(t,
				`{"runtime":"codex","value":`+strconv.Quote(tt.value)+`}`,
				string(manifest.Spec),
			)
		})
	}
}

func TestDecodeYAMLSubmissionPreservesExplicitExtremeNumbers(t *testing.T) {
	tests := map[string]string{
		"integer": `!!int 18446744073709551616`,
		"float":   `!!float 1e9999`,
	}
	for name, scalar := range tests {
		t.Run(name, func(t *testing.T) {
			source := strings.Replace(validSubmissionYAML, "  runtime: codex",
				"  runtime: codex\n  value: "+scalar, 1)

			manifest, err := DecodeYAMLSubmission([]byte(source))

			require.NoError(t, err)
			require.NotContains(t, string(manifest.Spec), `"value":"`)
		})
	}
}

func TestDecodeYAMLSubmissionBoundsGenerationOverflowError(t *testing.T) {
	const tailMarker = "314159265358979323846"
	number := strings.Repeat("9", 50_000) + tailMarker
	source := strings.Replace(validSubmissionYAML, "  namespace: team-one",
		"  namespace: team-one\n  generation: "+number, 1)
	require.Less(t, len("  generation: "+number), maxYAMLLineBytes)

	_, err := DecodeYAMLSubmission([]byte(source))

	require.Error(t, err)
	require.ErrorIs(t, err, ErrTypedJSONType)
	require.LessOrEqual(t, len(err.Error()), 512)
	require.Contains(t, err.Error(), "decode JSON manifest")
	require.Contains(t, err.Error(), "generation")
	require.Contains(t, err.Error(), "int64")
	require.Contains(t, err.Error(), "offset")
	require.NotContains(t, err.Error(), tailMarker)
}
