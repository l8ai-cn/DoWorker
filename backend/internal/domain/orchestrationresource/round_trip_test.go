package orchestrationresource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const roundTripWorkerJSON = `{
  "spec": {
    "credentialRef": {"revision": 3, "key": "access-token", "name": "registry-credentials"},
    "configuration": {
      "environment": [
        {"value": "-count=1", "name": "GOFLAGS"},
        {"name": "CI", "value": "true"}
      ],
      "features": ["review", "test"],
      "modelParameters": {
        "temperature": 0.2,
        "maxOutputTokens": 4096,
        "stopSequences": ["DONE", "HALT"]
      },
      "runtime": {
        "limits": {"timeoutSeconds": 900, "maxParallel": 3},
        "mode": "autonomous"
      }
    },
    "toolBindingRef": {"name": "repository-tools", "kind": "ToolBinding"},
    "promptRef": {
      "revision": 4,
      "name": "code-review-system",
      "namespace": "team-alpha",
      "kind": "Prompt",
      "apiVersion": "agentsmesh.io/v1alpha1"
    },
    "modelBindingRef": {"revision": 7, "name": "coding-primary", "kind": "ModelBinding"}
  },
  "metadata": {
    "labels": {"track": "phase-one", "role": "reviewer"},
    "displayName": "Canonical Worker",
    "namespace": "team-alpha",
    "name": "canonical-worker"
  },
  "kind": "WorkerTemplate",
  "apiVersion": "agentsmesh.io/v1alpha1"
}`

const roundTripWorkerJSONVariant = `{"kind":"WorkerTemplate",
"metadata":{"name":"canonical-worker","namespace":"team-alpha","displayName":"Canonical Worker",
"labels":{"role":"reviewer","track":"phase-one"}},"apiVersion":"agentsmesh.io/v1alpha1",
"spec":{"modelBindingRef":{"kind":"ModelBinding","name":"coding-primary","revision":7},
"promptRef":{"apiVersion":"agentsmesh.io/v1alpha1","kind":"Prompt","namespace":"team-alpha",
"name":"code-review-system","revision":4},"toolBindingRef":{"kind":"ToolBinding","name":"repository-tools"},
"credentialRef":{"name":"registry-credentials","key":"access-token","revision":3},
"configuration":{"runtime":{"mode":"autonomous","limits":{"maxParallel":3,"timeoutSeconds":900}},
"modelParameters":{"maxOutputTokens":4096,"stopSequences":["DONE","HALT"],"temperature":0.2},
"features":["review","test"],"environment":[{"name":"GOFLAGS","value":"-count=1"},{"name":"CI","value":"true"}]}}}`

const roundTripWorkerYAML = `kind: WorkerTemplate
apiVersion: "agentsmesh.io/v1alpha1"
metadata:
  namespace: team-alpha
  labels:
    role: reviewer
    track: "phase-one"
  name: 'canonical-worker'
  displayName: Canonical Worker
spec:
  promptRef:
    name: code-review-system
    revision: 4
    apiVersion: agentsmesh.io/v1alpha1
    kind: Prompt
    namespace: team-alpha
  modelBindingRef:
    name: coding-primary
    kind: ModelBinding
    revision: 7
  configuration:
    features:
      - review
      - "test"
    modelParameters:
      stopSequences:
        - DONE
        - HALT
      temperature: 0.2
      maxOutputTokens: 4096
    runtime:
      limits:
        timeoutSeconds: 900
        maxParallel: 3
      mode: autonomous
    environment:
      - value: "-count=1"
        name: GOFLAGS
      - name: CI
        value: "true"
  credentialRef:
    key: access-token
    revision: 3
    name: registry-credentials
  toolBindingRef:
    name: repository-tools
    kind: ToolBinding
`

const roundTripWorkerYAMLVariant = `# Equivalent draft with flow collections, blank lines, and different quoting.
metadata: {labels: {track: 'phase-one', role: reviewer}, namespace: "team-alpha",
  displayName: 'Canonical Worker', name: canonical-worker}
apiVersion: agentsmesh.io/v1alpha1

spec:
  toolBindingRef: {kind: ToolBinding, name: repository-tools}
  modelBindingRef: {name: coding-primary, revision: 7, kind: ModelBinding}
  credentialRef: {revision: 3, name: registry-credentials, key: access-token}
  configuration:
    environment:
      - {value: '-count=1', name: GOFLAGS}
      - {value: 'true', name: CI}
    runtime: {limits: {timeoutSeconds: 900, maxParallel: 3}, mode: autonomous}
    modelParameters: {stopSequences: [DONE, 'HALT'], temperature: 0.2, maxOutputTokens: 4096}
    features: [review, "test"]
  promptRef: {revision: 4, kind: Prompt, name: code-review-system,
    namespace: team-alpha, apiVersion: "agentsmesh.io/v1alpha1"}
kind: 'WorkerTemplate'
`

type roundTripWorkerTemplateSpec struct {
	ModelBindingRef Reference                    `json:"modelBindingRef"`
	PromptRef       Reference                    `json:"promptRef"`
	ToolBindingRef  Reference                    `json:"toolBindingRef"`
	CredentialRef   SecretReference              `json:"credentialRef"`
	Configuration   roundTripWorkerConfiguration `json:"configuration"`
}

type roundTripWorkerConfiguration struct {
	Runtime         roundTripRuntimeConfig         `json:"runtime"`
	ModelParameters roundTripModelParameters       `json:"modelParameters"`
	Features        []string                       `json:"features"`
	Environment     []roundTripEnvironmentVariable `json:"environment"`
}

type roundTripRuntimeConfig struct {
	Mode   string                 `json:"mode"`
	Limits roundTripRuntimeLimits `json:"limits"`
}

type roundTripRuntimeLimits struct {
	MaxParallel    int `json:"maxParallel"`
	TimeoutSeconds int `json:"timeoutSeconds"`
}

type roundTripModelParameters struct {
	Temperature     float64  `json:"temperature"`
	MaxOutputTokens int      `json:"maxOutputTokens"`
	StopSequences   []string `json:"stopSequences"`
}

type roundTripEnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type roundTripResult struct {
	Manifest      Manifest
	Spec          roundTripWorkerTemplateSpec
	CanonicalSpec []byte
}

func TestWorkerTemplateCanonicalRoundTripParity(t *testing.T) {
	registry := NewRegistry()
	meta := TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: "WorkerTemplate"}
	require.NoError(t, registry.Register(meta, roundTripWorkerSchema()))
	expected := expectedRoundTripResult(t)

	fromJSON := decodeRoundTripSubmission(t, registry, roundTripWorkerJSON, DecodeJSONSubmission)
	fromYAML := decodeRoundTripSubmission(t, registry, roundTripWorkerYAML, DecodeYAMLSubmission)
	requireRoundTripGroundTruth(t, expected, fromJSON)
	requireRoundTripGroundTruth(t, expected, fromYAML)
	requireValidRoundTripReferences(t, fromJSON.Manifest.Metadata, fromJSON.Spec)
	requireValidRoundTripReferences(t, fromYAML.Manifest.Metadata, fromYAML.Spec)

	variants := []struct {
		name   string
		source string
		decode func([]byte) (Manifest, error)
	}{
		{name: "JSON field order", source: roundTripWorkerJSONVariant, decode: DecodeJSONSubmission},
		{name: "YAML order whitespace and quotes", source: roundTripWorkerYAMLVariant, decode: DecodeYAMLSubmission},
	}
	for _, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			actual := decodeRoundTripSubmission(t, registry, variant.source, variant.decode)
			requireRoundTripGroundTruth(t, expected, actual)
		})
	}

	encodedJSON, err := EncodeJSON(fromJSON.Manifest)
	require.NoError(t, err)
	require.Equal(t, 1, bytes.Count(encodedJSON, []byte{'\n'}))
	require.True(t, bytes.HasSuffix(encodedJSON, []byte{'\n'}))

	encodedYAML, err := EncodeYAML(fromYAML.Manifest)
	require.NoError(t, err)
	requireHumanReadableRoundTripYAML(t, encodedYAML)

	jsonRoundTrip := decodeRoundTripSubmission(t, registry, string(encodedJSON), DecodeJSONSubmission)
	yamlRoundTrip := decodeRoundTripSubmission(t, registry, string(encodedYAML), DecodeYAMLSubmission)
	requireRoundTripGroundTruth(t, expected, jsonRoundTrip)
	requireRoundTripGroundTruth(t, expected, yamlRoundTrip)
}

func roundTripWorkerSchema() Schema {
	return Schema{
		NewSpec: func() any { return &roundTripWorkerTemplateSpec{} },
		Validate: func(metadata Metadata, value any) error {
			spec := value.(*roundTripWorkerTemplateSpec)
			for _, item := range roundTripReferenceFields(*spec) {
				if err := item.reference.ValidateDraft(metadata.Namespace.String()); err != nil {
					return fmt.Errorf("%s: %w", item.name, err)
				}
			}
			if err := spec.CredentialRef.Validate(); err != nil {
				return fmt.Errorf("credentialRef: %w", err)
			}
			return nil
		},
	}
}

func decodeRoundTripSubmission(
	t *testing.T,
	registry *Registry,
	source string,
	decode func([]byte) (Manifest, error),
) roundTripResult {
	t.Helper()
	manifest, err := decode([]byte(source))
	require.NoError(t, err)
	decoded, err := registry.DecodeAndValidate(manifest)
	require.NoError(t, err)
	spec, ok := decoded.(*roundTripWorkerTemplateSpec)
	require.True(t, ok)
	canonicalSpec, err := json.Marshal(spec)
	require.NoError(t, err)
	return roundTripResult{Manifest: manifest, Spec: *spec, CanonicalSpec: canonicalSpec}
}

func expectedRoundTripResult(t *testing.T) roundTripResult {
	t.Helper()
	spec := roundTripWorkerTemplateSpec{
		ModelBindingRef: Reference{
			Kind: "ModelBinding", Name: slugkit.Slug("coding-primary"), Revision: 7,
		},
		PromptRef: Reference{
			APIVersion: APIVersionV1Alpha1, Kind: "Prompt",
			Namespace: slugkit.Slug("team-alpha"), Name: slugkit.Slug("code-review-system"), Revision: 4,
		},
		ToolBindingRef: Reference{
			Kind: "ToolBinding", Name: slugkit.Slug("repository-tools"),
		},
		CredentialRef: SecretReference{
			Name: slugkit.Slug("registry-credentials"), Key: slugkit.Slug("access-token"), Revision: 3,
		},
		Configuration: roundTripWorkerConfiguration{
			Runtime: roundTripRuntimeConfig{
				Mode: "autonomous",
				Limits: roundTripRuntimeLimits{
					MaxParallel: 3, TimeoutSeconds: 900,
				},
			},
			ModelParameters: roundTripModelParameters{
				Temperature: 0.2, MaxOutputTokens: 4096, StopSequences: []string{"DONE", "HALT"},
			},
			Features: []string{"review", "test"},
			Environment: []roundTripEnvironmentVariable{
				{Name: "GOFLAGS", Value: "-count=1"},
				{Name: "CI", Value: "true"},
			},
		},
	}
	canonicalSpec, err := json.Marshal(spec)
	require.NoError(t, err)
	return roundTripResult{
		Manifest: Manifest{
			TypeMeta: TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: "WorkerTemplate"},
			Metadata: Metadata{
				Name: slugkit.Slug("canonical-worker"), Namespace: slugkit.Slug("team-alpha"),
				DisplayName: "Canonical Worker",
				Labels:      map[string]string{"role": "reviewer", "track": "phase-one"},
			},
		},
		Spec:          spec,
		CanonicalSpec: canonicalSpec,
	}
}

func requireRoundTripGroundTruth(t *testing.T, expected, actual roundTripResult) {
	t.Helper()
	require.Equal(t, expected.Manifest.TypeMeta, actual.Manifest.TypeMeta)
	require.Equal(t, expected.Manifest.Metadata, actual.Manifest.Metadata)
	require.Equal(t, expected.CanonicalSpec, actual.CanonicalSpec)
	require.Equal(t, expected.Spec, actual.Spec)

	expectedRefs := roundTripReferenceFields(expected.Spec)
	actualRefs := roundTripReferenceFields(actual.Spec)
	require.Len(t, actualRefs, len(expectedRefs))
	for index := range expectedRefs {
		require.Equal(t, expectedRefs[index].name, actualRefs[index].name)
		require.Equal(t, expectedRefs[index].reference.APIVersion, actualRefs[index].reference.APIVersion)
		require.Equal(t, expectedRefs[index].reference.Kind, actualRefs[index].reference.Kind)
		require.Equal(t, expectedRefs[index].reference.Namespace, actualRefs[index].reference.Namespace)
		require.Equal(t, expectedRefs[index].reference.Name, actualRefs[index].reference.Name)
		require.Equal(t, expectedRefs[index].reference.Revision, actualRefs[index].reference.Revision)
	}
	require.Equal(t, expected.Spec.CredentialRef.Name, actual.Spec.CredentialRef.Name)
	require.Equal(t, expected.Spec.CredentialRef.Key, actual.Spec.CredentialRef.Key)
	require.Equal(t, expected.Spec.CredentialRef.Revision, actual.Spec.CredentialRef.Revision)
	require.Equal(t, "autonomous", actual.Spec.Configuration.Runtime.Mode)
	require.Equal(t, 3, actual.Spec.Configuration.Runtime.Limits.MaxParallel)
	require.Equal(t, 900, actual.Spec.Configuration.Runtime.Limits.TimeoutSeconds)
	require.Equal(t, 0.2, actual.Spec.Configuration.ModelParameters.Temperature)
	require.Equal(t, 4096, actual.Spec.Configuration.ModelParameters.MaxOutputTokens)
	require.Equal(t, []string{"DONE", "HALT"}, actual.Spec.Configuration.ModelParameters.StopSequences)
	require.Equal(t, []string{"review", "test"}, actual.Spec.Configuration.Features)
	require.Equal(t, expected.Spec.Configuration.Environment, actual.Spec.Configuration.Environment)
}

func requireValidRoundTripReferences(
	t *testing.T,
	metadata Metadata,
	spec roundTripWorkerTemplateSpec,
) {
	t.Helper()
	for _, item := range roundTripReferenceFields(spec) {
		require.NoError(t, item.reference.ValidateDraft(metadata.Namespace.String()), item.name)
	}
	require.NoError(t, spec.CredentialRef.Validate())
}

func roundTripReferenceFields(spec roundTripWorkerTemplateSpec) []struct {
	name      string
	reference Reference
} {
	return []struct {
		name      string
		reference Reference
	}{
		{name: "modelBindingRef", reference: spec.ModelBindingRef},
		{name: "promptRef", reference: spec.PromptRef},
		{name: "toolBindingRef", reference: spec.ToolBindingRef},
	}
}

func requireHumanReadableRoundTripYAML(t *testing.T, source []byte) {
	t.Helper()
	text := string(source)
	require.True(t, strings.HasSuffix(text, "\n"))
	require.False(t, strings.HasSuffix(text, "\n\n"))
	require.False(t, strings.HasPrefix(text, "---"))
	require.NotContains(t, text, "\t")
	require.NotContains(t, text, "!!")
	require.Contains(t, text, `value: "true"`)

	var document yaml.Node
	require.NoError(t, yaml.Unmarshal(source, &document))
	require.Len(t, document.Content, 1)
	requireBlockRoundTripYAML(t, document.Content[0])
}

func requireBlockRoundTripYAML(t *testing.T, node *yaml.Node) {
	t.Helper()
	if node.Kind == yaml.MappingNode || node.Kind == yaml.SequenceNode {
		require.Zero(t, node.Style&yaml.FlowStyle)
	}
	if node.Kind == yaml.ScalarNode {
		require.Zero(t, node.Style&yaml.TaggedStyle)
	}
	for _, child := range node.Content {
		requireBlockRoundTripYAML(t, child)
	}
}
