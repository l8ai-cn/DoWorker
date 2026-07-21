package orchestrationcontrol

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/stretchr/testify/require"
)

func TestCanonicalJSONIgnoresMapOrderAndWhitespace(t *testing.T) {
	first, err := CanonicalJSON(json.RawMessage(`{
		"nested": {"z": 3, "a": 1},
		"list": [3, 2, 1]
	}`))
	require.NoError(t, err)
	second, err := CanonicalJSON(map[string]any{
		"list": []any{3, 2, 1},
		"nested": map[string]any{
			"a": 1,
			"z": 3,
		},
	})
	require.NoError(t, err)

	require.JSONEq(t, string(first), string(second))
	require.Equal(t, first, second)
	require.Equal(t, `{"list":[3,2,1],"nested":{"a":1,"z":3}}`, string(first))
}

func TestCanonicalJSONEnforcesContainerShapeAndRejectsUnsupportedValues(t *testing.T) {
	_, err := CanonicalJSONObject([]string{"wrong"})
	require.ErrorIs(t, err, ErrInvalid)
	_, err = CanonicalJSONArray(map[string]string{"wrong": "shape"})
	require.ErrorIs(t, err, ErrInvalid)

	for _, value := range []any{
		math.NaN(),
		math.Inf(1),
		func() {},
		make(chan int),
		json.RawMessage(`{"value":NaN}`),
		"scalar",
		nil,
	} {
		_, err := CanonicalJSON(value)
		require.ErrorIs(t, err, ErrInvalid)
	}
}

func TestCanonicalJSONRejectsDuplicateKeysWithoutEchoingThem(t *testing.T) {
	const duplicate = "sk-live-duplicate-key"
	_, err := CanonicalJSON(json.RawMessage(
		`{"` + duplicate + `":1,"` + duplicate + `":2}`,
	))

	require.ErrorIs(t, err, ErrInvalid)
	require.NotContains(t, err.Error(), duplicate)
}

func TestDigestCanonicalJSONUsesLowercaseSHA256Grammar(t *testing.T) {
	first, err := DigestCanonicalJSON(map[string]any{"b": 2, "a": 1})
	require.NoError(t, err)
	second, err := DigestCanonicalJSON(json.RawMessage("{\n \"a\":1, \"b\":2\n}"))
	require.NoError(t, err)

	require.Equal(t, first, second)
	require.Regexp(t, `^sha256:[0-9a-f]{64}$`, first)
	require.NotContains(t, first, "A")
}

func TestComputePlanHashIsStableAcrossReferenceOrder(t *testing.T) {
	firstRef := validResolvedReferenceForControl()
	secondRef := firstRef
	secondRef.Kind = "Prompt"
	secondRef.Name = "system-prompt"
	secondRef.UID = "44444444-4444-4444-8444-444444444444"
	secondRef.Revision = 5
	secondRef.Digest = "sha256:" + strings.Repeat("b", 64)

	input := validPlanHashInput()
	input.ResolvedReferences = []ResolvedReference{firstRef, secondRef}
	first, err := ComputePlanHash(input)
	require.NoError(t, err)

	input.ResolvedReferences = []ResolvedReference{secondRef, firstRef}
	second, err := ComputePlanHash(input)
	require.NoError(t, err)

	require.Equal(t, first, second)
	require.Regexp(t, `^sha256:[0-9a-f]{64}$`, first)
}

func TestComputePlanHashRejectsDuplicateReferenceIdentityRevision(t *testing.T) {
	input := validPlanHashInput()
	duplicate := validResolvedReferenceForControl()
	duplicate.Digest = "sha256:" + strings.Repeat("f", 64)
	input.ResolvedReferences = []ResolvedReference{
		validResolvedReferenceForControl(),
		duplicate,
	}

	_, err := ComputePlanHash(input)
	require.ErrorIs(t, err, ErrInvalid)
}

func TestComputePlanHashBindsOpaqueOptionsRevision(t *testing.T) {
	input := validPlanHashInput()
	first, err := ComputePlanHash(input)
	require.NoError(t, err)

	input.OptionsRevision = "runtime-catalog-5"
	second, err := ComputePlanHash(input)
	require.NoError(t, err)
	require.NotEqual(t, first, second)

	for _, invalidRevision := range []string{
		"",
		" runtime-catalog-5",
		"runtime-catalog-5\n",
		strings.Repeat("x", maxOptionsRevisionRunes+1),
	} {
		input.OptionsRevision = invalidRevision
		_, err := ComputePlanHash(input)
		require.ErrorIs(t, err, ErrInvalid)
	}
}

func TestComputePlanHashDistinguishesCreateAndUpdateBaseState(t *testing.T) {
	create := validPlanHashInput()
	createHash, err := ComputePlanHash(create)
	require.NoError(t, err)

	update := create
	update.Operation = PlanOperationUpdate
	update.BaseUID = testTargetID
	update.BaseResourceVersion = 8
	updateHash, err := ComputePlanHash(update)
	require.NoError(t, err)
	require.NotEqual(t, createHash, updateHash)

	update.BaseUID = ""
	_, err = ComputePlanHash(update)
	require.ErrorIs(t, err, ErrInvalid)

	create.BaseUID = testTargetID
	create.BaseResourceVersion = 1
	_, err = ComputePlanHash(create)
	require.ErrorIs(t, err, ErrInvalid)
}

func TestCanonicalDraftHashIgnoresJSONAndYAMLSourceFormatting(t *testing.T) {
	jsonSource := []byte(`{
		"apiVersion":"agentcloud.io/v1alpha1",
		"kind":"WorkerTemplate",
		"metadata":{"name":"worker-one","namespace":"team-alpha"},
		"spec":{"settings":{"b":2,"a":1}}
	}`)
	yamlSource := []byte(`
kind: WorkerTemplate
metadata:
  namespace: team-alpha
  name: worker-one
spec:
  settings: {a: 1, b: 2}
apiVersion: agentcloud.io/v1alpha1
`)
	fromJSON, err := orchestrationresource.DecodeJSONSubmission(jsonSource)
	require.NoError(t, err)
	fromYAML, err := orchestrationresource.DecodeYAMLSubmission(yamlSource)
	require.NoError(t, err)

	jsonHash, err := DigestCanonicalJSON(fromJSON)
	require.NoError(t, err)
	yamlHash, err := DigestCanonicalJSON(fromYAML)
	require.NoError(t, err)
	require.Equal(t, jsonHash, yamlHash)

	first := validPlanHashInput()
	first.DraftHash = jsonHash
	second := first
	second.DraftHash = yamlHash
	firstHash, err := ComputePlanHash(first)
	require.NoError(t, err)
	secondHash, err := ComputePlanHash(second)
	require.NoError(t, err)
	require.Equal(t, firstHash, secondHash)
}

func TestCanonicalJSONNormalizesEquivalentNumberSpellings(t *testing.T) {
	var canonical []byte
	for _, raw := range []json.RawMessage{
		json.RawMessage(`{"value":1}`),
		json.RawMessage(`{"value":1.0}`),
		json.RawMessage(`{"value":1e0}`),
	} {
		encoded, err := CanonicalJSONObject(raw)
		require.NoError(t, err)
		if canonical == nil {
			canonical = encoded
			continue
		}
		require.Equal(t, canonical, encoded)
	}
}

func validPlanHashInput() PlanHashInput {
	return PlanHashInput{
		Operation:          PlanOperationCreate,
		Scope:              validScope(),
		Target:             validTarget(),
		DraftHash:          "sha256:" + strings.Repeat("c", 64),
		ResolvedReferences: []ResolvedReference{validResolvedReferenceForControl()},
		ArtifactDigest:     "sha256:" + strings.Repeat("d", 64),
		OptionsRevision:    "runtime-catalog-4",
	}
}
