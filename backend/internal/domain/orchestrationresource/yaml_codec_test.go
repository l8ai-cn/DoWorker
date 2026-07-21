package orchestrationresource

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const validSubmissionYAML = `apiVersion: agentcloud.io/v1alpha1
kind: WorkerTemplate
metadata:
  name: worker-one
  namespace: team-one
  displayName: Worker One
  labels:
    role: build-agent
spec:
  runtime: codex
`

const submissionYAMLNodeOverhead = 16

func TestDecodeYAMLSubmissionAcceptsValidManifestWithJSONParity(t *testing.T) {
	fromYAML, err := DecodeYAMLSubmission([]byte(validSubmissionYAML))
	require.NoError(t, err)

	fromJSON, err := DecodeJSONSubmission([]byte(validSubmissionJSON))
	require.NoError(t, err)

	yamlJSON, err := EncodeJSON(fromYAML)
	require.NoError(t, err)
	jsonJSON, err := EncodeJSON(fromJSON)
	require.NoError(t, err)
	require.Equal(t, jsonJSON, yamlJSON)
}

func TestDecodeYAMLSubmissionRejectsEmptyAndOversizedSources(t *testing.T) {
	_, err := DecodeYAMLSubmission(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exactly one document")

	_, err = DecodeYAMLSubmission([]byte(" \n\t"))
	require.Error(t, err)

	_, err = DecodeYAMLSubmission(make([]byte, maxYAMLManifestBytes+1))
	require.Error(t, err)
	require.Contains(t, err.Error(), "262144")
}

func TestDecodeYAMLSubmissionRejectsDuplicateKeysAtAnyDepth(t *testing.T) {
	tests := []string{
		strings.Replace(validSubmissionYAML, "kind: WorkerTemplate", "kind: WorkerTemplate\nkind: RunnerTemplate", 1),
		strings.Replace(validSubmissionYAML, "  runtime: codex", "  runtime: codex\n  nested: {enabled: true, enabled: false}", 1),
	}
	for _, source := range tests {
		_, err := DecodeYAMLSubmission([]byte(source))
		require.ErrorIs(t, err, ErrDuplicateYAMLKey)
	}

	const secret = "sk-live-duplicate-secret"
	key := secret + strings.Repeat("untrusted", 30)
	source := strings.Replace(validSubmissionYAML, "  runtime: codex",
		"  runtime: codex\n  "+key+": true\n  "+key+": false", 1)
	_, err := DecodeYAMLSubmission([]byte(source))
	require.ErrorIs(t, err, ErrDuplicateYAMLKey)
	require.NotContains(t, err.Error(), secret)
	require.NotContains(t, err.Error(), strings.Repeat("untrusted", 11))
	require.LessOrEqual(t, len(err.Error()), 512)
}

func TestDecodeYAMLSubmissionRejectsYAMLOnlyFeatures(t *testing.T) {
	tests := map[string]string{
		"anchor":     "  runtime: &runtime codex",
		"merge key":  "  runtime: codex\n  <<: {enabled: true}",
		"custom tag": "  runtime: !runtime codex",
		"timestamp":  "  runtime: codex\n  created: 2026-07-14",
		"binary":     "  runtime: codex\n  payload: !!binary SGVsbG8=",
	}
	for name, replacement := range tests {
		t.Run(name, func(t *testing.T) {
			source := strings.Replace(validSubmissionYAML, "  runtime: codex", replacement, 1)
			_, err := DecodeYAMLSubmission([]byte(source))
			require.Error(t, err)
		})
	}

	quoted := strings.Replace(validSubmissionYAML, "  runtime: codex",
		"  runtime: codex\n  created: \"2026-07-14\"", 1)
	_, err := DecodeYAMLSubmission([]byte(quoted))
	require.NoError(t, err)
}

func TestDecodeYAMLSubmissionAllowsStringMergeKeyWithJSONParity(t *testing.T) {
	jsonSource := strings.Replace(validSubmissionJSON, `"runtime": "codex"`,
		`"runtime": "codex", "<<": "literal"`, 1)
	fromJSON, err := DecodeJSONSubmission([]byte(jsonSource))
	require.NoError(t, err)
	expected, err := EncodeJSON(fromJSON)
	require.NoError(t, err)

	for name, key := range map[string]string{
		"quoted":    `"<<"`,
		"explicit":  `!!str <<`,
		"long form": `!<tag:yaml.org,2002:str> <<`,
	} {
		t.Run(name, func(t *testing.T) {
			source := strings.Replace(validSubmissionYAML, "  runtime: codex",
				"  runtime: codex\n  "+key+": literal", 1)
			fromYAML, err := DecodeYAMLSubmission([]byte(source))
			require.NoError(t, err)
			actual, err := EncodeJSON(fromYAML)
			require.NoError(t, err)
			require.JSONEq(t, string(expected), string(actual))
		})
	}
}

func TestDecodeYAMLSubmissionValidatesExplicitNullLexemes(t *testing.T) {
	for _, value := range []string{"", "~", "null", "Null", "NULL", "nUlL"} {
		source := strings.Replace(validSubmissionYAML, "  runtime: codex",
			"  runtime: codex\n  value: !!null "+value, 1)
		manifest, err := DecodeYAMLSubmission([]byte(source))
		require.NoError(t, err)
		require.JSONEq(t, `{"runtime":"codex","value":null}`, string(manifest.Spec))
	}

	source := strings.Replace(validSubmissionYAML, "  runtime: codex",
		"  runtime: codex\n  value: !!null not-null", 1)
	_, err := DecodeYAMLSubmission([]byte(source))
	require.Error(t, err)
	require.Contains(t, err.Error(), "null")
}

func TestValidateYAMLTreeRejectsAliasAndMalformedMapping(t *testing.T) {
	alias := &yaml.Node{
		Kind: yaml.DocumentNode,
		Content: []*yaml.Node{{
			Kind: yaml.AliasNode,
		}},
	}
	require.Error(t, validateYAMLTree(alias))

	oddMapping := &yaml.Node{
		Kind: yaml.DocumentNode,
		Content: []*yaml.Node{{
			Kind:    yaml.MappingNode,
			Tag:     "!!map",
			Content: []*yaml.Node{{Kind: yaml.ScalarNode, Tag: "!!str", Value: "key"}},
		}},
	}
	err := validateYAMLTree(oddMapping)
	require.Error(t, err)
	require.Contains(t, err.Error(), "even")
}

func TestValidateYAMLTreeStopsBeforeSchedulingTooManyNodes(t *testing.T) {
	mapping := &yaml.Node{
		Kind:    yaml.MappingNode,
		Tag:     "!!map",
		Content: make([]*yaml.Node, maxYAMLNodes),
	}
	document := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{mapping},
	}

	err := validateYAMLTree(document)
	require.Error(t, err)
	require.Contains(t, err.Error(), "10000 nodes")
}

func TestDecodeYAMLSubmissionRejectsAdditionalDocuments(t *testing.T) {
	for _, suffix := range []string{"---\n", "---\n{}\n"} {
		_, err := DecodeYAMLSubmission([]byte(validSubmissionYAML + suffix))
		require.Error(t, err)
		require.Contains(t, err.Error(), "exactly one document")
	}
}

func TestDecodeYAMLSubmissionReusesStrictJSONValidation(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectedErr error
		field       string
	}{
		{
			name:        "unknown top level",
			source:      validSubmissionYAML + "unexpected: true\n",
			expectedErr: ErrUnknownJSONField,
			field:       "unexpected",
		},
		{
			name: "unknown metadata",
			source: strings.Replace(validSubmissionYAML, "  name: worker-one",
				"  name: worker-one\n  unexpected: true", 1),
			expectedErr: ErrUnknownJSONField,
			field:       "unexpected",
		},
		{
			name: "server field",
			source: strings.Replace(validSubmissionYAML, "  name: worker-one",
				"  name: worker-one\n  uid: resource-123", 1),
			field: "metadata.uid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeYAMLSubmission([]byte(tt.source))
			require.Error(t, err)
			if tt.expectedErr != nil {
				require.True(t, errors.Is(err, tt.expectedErr))
				require.NotContains(t, err.Error(), tt.field)
				require.LessOrEqual(t, len(err.Error()), 512)
				return
			}
			require.Contains(t, err.Error(), tt.field)
		})
	}
}

func TestDecodeYAMLSubmissionRejectsNonStringMappingKeys(t *testing.T) {
	for _, key := range []string{"1", "true", "null", "1e9999", "18446744073709551616"} {
		source := validSubmissionYAML + key + ": value\n"
		_, err := DecodeYAMLSubmission([]byte(source))
		require.Error(t, err)
		require.Contains(t, err.Error(), "mapping keys must be strings")
	}
}

func TestDecodeYAMLSubmissionEnforcesMaximumContainerDepth(t *testing.T) {
	manifest, err := DecodeYAMLSubmission(nestedSubmissionYAML(64))
	require.NoError(t, err)
	require.NotEmpty(t, manifest.Spec)

	_, err = DecodeYAMLSubmission(nestedSubmissionYAML(65))
	require.Error(t, err)
	require.Contains(t, err.Error(), "depth 64")
}

func TestDecodeYAMLSubmissionEnforcesMaximumNodeCount(t *testing.T) {
	boundaryItems := maxYAMLNodes - submissionYAMLNodeOverhead
	manifest, err := DecodeYAMLSubmission(submissionYAMLWithSequenceItems(boundaryItems))
	require.NoError(t, err)
	require.NotEmpty(t, manifest.Spec)

	_, err = DecodeYAMLSubmission(submissionYAMLWithSequenceItems(boundaryItems + 1))
	require.Error(t, err)
	require.Contains(t, err.Error(), "10000 nodes")
}

func TestDecodeYAMLSubmissionPreservesJSONNumberLexemes(t *testing.T) {
	valid := []string{"0", "-0", "1", "-12", "1.25", "1e3", "1E+3", "-1.2e-3"}
	for _, number := range valid {
		t.Run(number, func(t *testing.T) {
			source := strings.Replace(validSubmissionYAML, "  runtime: codex",
				"  runtime: codex\n  value: "+number, 1)
			manifest, err := DecodeYAMLSubmission([]byte(source))
			require.NoError(t, err)
			require.Contains(t, string(manifest.Spec), `"value":`+number)
		})
	}

	invalid := []string{"0x10", "01", "1_000", ".inf", ".nan", ".NaN", "+1", "1.", ".5"}
	for _, number := range invalid {
		t.Run(number, func(t *testing.T) {
			source := strings.Replace(validSubmissionYAML, "  runtime: codex",
				"  runtime: codex\n  value: "+number, 1)
			_, err := DecodeYAMLSubmission([]byte(source))
			require.Error(t, err)
			require.Contains(t, err.Error(), "JSON")
		})
	}
}

func TestEncodeYAMLProducesStableStoredManifestWithoutMutation(t *testing.T) {
	manifest := validManifestForTest()
	manifest.Spec = json.RawMessage(`{
		"timestampString":"2026-07-14",
		"booleanString":"true",
		"numberString":"01",
		"integer":9007199254740993,
		"decimal":1.2300,
		"exponent":1e+30,
		"runtime":"codex"
	}`)
	manifest.Status = json.RawMessage(`{"ready":true}`)
	before := cloneManifestForJSONCodecTest(manifest)

	encoded, err := EncodeYAML(manifest)
	require.NoError(t, err)
	require.Equal(t, before, manifest)
	require.NotEmpty(t, encoded)
	require.True(t, strings.HasSuffix(string(encoded), "\n"))

	encodedAgain, err := EncodeYAML(manifest)
	require.NoError(t, err)
	require.Equal(t, encoded, encodedAgain)

	var document yaml.Node
	require.NoError(t, yaml.Unmarshal(encoded, &document))
	requireYAMLMappingKeysSorted(t, document.Content[0])
	metadata := yamlMappingValue(t, document.Content[0], "metadata")
	require.Equal(t, "res_123", yamlMappingValue(t, metadata, "uid").Value)
	require.Equal(t, "42+abc", yamlMappingValue(t, metadata, "resourceVersion").Value)
	require.Equal(t, "2", yamlMappingValue(t, metadata, "generation").Value)
	require.Equal(t, "true", yamlMappingValue(t,
		yamlMappingValue(t, document.Content[0], "status"), "ready").Value)
	spec := yamlMappingValue(t, document.Content[0], "spec")
	for _, key := range []string{"timestampString", "booleanString", "numberString"} {
		require.Equal(t, "!!str", yamlMappingValue(t, spec, key).ShortTag())
	}
	require.Equal(t, "!!int", yamlMappingValue(t, spec, "integer").ShortTag())
	for _, key := range []string{"decimal", "exponent"} {
		require.Equal(t, "!!float", yamlMappingValue(t, spec, key).ShortTag())
	}
	require.Contains(t, string(encoded), `booleanString: "true"`)
	require.Contains(t, string(encoded), `numberString: "01"`)
	require.Contains(t, string(encoded), `timestampString: "2026-07-14"`)
	require.Contains(t, string(encoded), "integer: 9007199254740993")
	require.Contains(t, string(encoded), "decimal: 1.2300")
	require.Contains(t, string(encoded), "exponent: 1e+30")
	require.NotContains(t, string(encoded), "!!str")
}

func TestEncodeYAMLRejectsInvalidStoredManifest(t *testing.T) {
	manifest := validManifestForTest()
	manifest.Spec = json.RawMessage(`{}`)

	encoded, err := EncodeYAML(manifest)
	require.Error(t, err)
	require.Nil(t, encoded)
	require.Contains(t, err.Error(), "spec")
}

func TestEncodeYAMLPreservesMultilineStringTrailingNewlines(t *testing.T) {
	manifest := validManifestForTest()
	manifest.Spec = json.RawMessage(`{"runtime":"codex","text":"line\n\n"}`)

	encoded, err := EncodeYAML(manifest)
	require.NoError(t, err)

	var document yaml.Node
	require.NoError(t, yaml.Unmarshal(encoded, &document))
	spec := yamlMappingValue(t, document.Content[0], "spec")
	require.Equal(t, "line\n\n", yamlMappingValue(t, spec, "text").Value)
}

func nestedSubmissionYAML(depth int) []byte {
	value := "true"
	for currentDepth := 3; currentDepth <= depth; currentDepth++ {
		if currentDepth%2 == 0 {
			value = `{"value":` + value + `}`
		} else {
			value = `[` + value + `]`
		}
	}
	return []byte(strings.Replace(validSubmissionYAML, "  runtime: codex", "  payload: "+value, 1))
}

func submissionYAMLWithSequenceItems(items int) []byte {
	var source strings.Builder
	source.WriteString("apiVersion: agentcloud.io/v1alpha1\nkind: WorkerTemplate\n")
	source.WriteString("metadata: {name: worker-one, namespace: team-one}\nspec:\n  payload:\n")
	source.WriteString(strings.Repeat("    - null\n", items))
	return []byte(source.String())
}

func yamlMappingValue(t *testing.T, mapping *yaml.Node, key string) *yaml.Node {
	t.Helper()
	for index := 0; index < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1]
		}
	}
	t.Fatalf("missing YAML key %q", key)
	return nil
}

func requireYAMLMappingKeysSorted(t *testing.T, node *yaml.Node) {
	t.Helper()
	if node.Kind == yaml.MappingNode {
		keys := make([]string, 0, len(node.Content)/2)
		for index := 0; index < len(node.Content); index += 2 {
			keys = append(keys, node.Content[index].Value)
			requireYAMLMappingKeysSorted(t, node.Content[index+1])
		}
		require.True(t, sort.StringsAreSorted(keys), "mapping keys are not sorted: %v", keys)
		return
	}
	for _, child := range node.Content {
		requireYAMLMappingKeysSorted(t, child)
	}
}
