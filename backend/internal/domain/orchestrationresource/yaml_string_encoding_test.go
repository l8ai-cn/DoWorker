package orchestrationresource

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestEncodeYAMLPreservesRiskyStringIdentityForKeysAndValues(t *testing.T) {
	tests := map[string]string{
		"extreme exponent": "1e9999",
		"large integer":    "18446744073709551616",
		"boolean":          "true",
		"null":             "null",
		"timestamp":        "2026-07-14",
		"merge":            "<<",
		"hexadecimal":      "0x10",
		"infinity":         ".inf",
		"nan":              ".NaN",
		"custom tag":       "!custom",
		"alias":            "*alias",
		"leading dash":     "- item",
		"leading hash":     "#comment",
		"multiline":        "line\n\n",
	}

	for name, risky := range tests {
		t.Run(name, func(t *testing.T) {
			manifest := submissionManifestForYAMLTest()
			spec, err := json.Marshal(map[string]any{
				"runtime": "codex",
				"probe":   risky,
				risky:     "literal",
			})
			require.NoError(t, err)
			manifest.Spec = spec

			canonical, err := EncodeJSON(manifest)
			require.NoError(t, err)
			encoded, err := EncodeYAML(manifest)
			require.NoError(t, err)
			require.NotContains(t, string(encoded), "!!str")

			decoded, err := DecodeYAMLSubmission(encoded)
			require.NoError(t, err)
			actual, err := EncodeJSON(decoded)
			require.NoError(t, err)
			require.Equal(t, canonical, actual)

			var document yaml.Node
			require.NoError(t, yaml.Unmarshal(encoded, &document))
			specNode := yamlMappingValue(t, document.Content[0], "spec")
			requireQuotedYAMLString(t, yamlMappingKey(t, specNode, risky))
			requireQuotedYAMLString(t, yamlMappingValue(t, specNode, "probe"))
		})
	}
}

func TestEncodeYAMLKeepsOrdinaryStringNatural(t *testing.T) {
	manifest := submissionManifestForYAMLTest()
	manifest.Spec = json.RawMessage(`{
		"runtime":"codex",
		"probe":"worker-one",
		"worker-one":"literal"
	}`)

	encoded, err := EncodeYAML(manifest)

	require.NoError(t, err)
	require.Contains(t, string(encoded), "probe: worker-one")
	require.Contains(t, string(encoded), "worker-one: literal")
}

func yamlMappingKey(t *testing.T, mapping *yaml.Node, key string) *yaml.Node {
	t.Helper()
	for index := 0; index < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index]
		}
	}
	t.Fatalf("missing YAML key %q", key)
	return nil
}

func requireQuotedYAMLString(t *testing.T, node *yaml.Node) {
	t.Helper()
	require.Equal(t, "!!str", node.ShortTag())
	require.NotZero(t, node.Style&(yaml.DoubleQuotedStyle|yaml.SingleQuotedStyle))
}
