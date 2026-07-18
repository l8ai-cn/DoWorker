package workerdependency

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeCanonicalizesDependencyOrder(t *testing.T) {
	left := validDocument(t)
	right := validDocument(t)
	right.Skills[0], right.Skills[1] = right.Skills[1], right.Skills[0]
	right.Models.Primary.Modalities[0], right.Models.Primary.Modalities[1] =
		right.Models.Primary.Modalities[1], right.Models.Primary.Modalities[0]
	right.RuntimeBundles[0].Values[0], right.RuntimeBundles[0].Values[1] =
		right.RuntimeBundles[0].Values[1], right.RuntimeBundles[0].Values[0]

	leftJSON, err := Encode(left)
	require.NoError(t, err)
	rightJSON, err := Encode(right)
	require.NoError(t, err)
	assert.Equal(t, leftJSON, rightJSON)

	leftDigest, err := Digest(left)
	require.NoError(t, err)
	rightDigest, err := Digest(right)
	require.NoError(t, err)
	assert.Equal(t, leftDigest, rightDigest)

	decoded, err := Decode(leftJSON)
	require.NoError(t, err)
	assert.Equal(t, Normalize(left), decoded)
}

func TestRuntimeBundleOrderRemainsExecutionSignificant(t *testing.T) {
	first := validDocument(t)
	second := validDocument(t)
	second.RuntimeBundles[0], second.RuntimeBundles[1] =
		second.RuntimeBundles[1], second.RuntimeBundles[0]

	firstDigest, err := Digest(first)
	require.NoError(t, err)
	secondDigest, err := Digest(second)
	require.NoError(t, err)
	assert.NotEqual(t, firstDigest, secondDigest)

	encoded, err := Encode(second)
	require.NoError(t, err)
	decoded, err := Decode(encoded)
	require.NoError(t, err)
	assert.Equal(
		t,
		second.RuntimeBundles[0].Pin.DomainID,
		decoded.RuntimeBundles[0].Pin.DomainID,
	)
}

func TestDecodeRejectsUnknownVersionFieldsAndTrailingData(t *testing.T) {
	encoded, err := Encode(validDocument(t))
	require.NoError(t, err)

	var versioned map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(encoded, &versioned))
	versioned["version"] = json.RawMessage(`2`)
	versionTwo, err := json.Marshal(versioned)
	require.NoError(t, err)
	_, err = Decode(versionTwo)
	require.ErrorIs(t, err, ErrUnsupportedVersion)

	var root map[string]any
	require.NoError(t, json.Unmarshal(encoded, &root))
	root["credentials"] = map[string]string{"api_key": "must-not-leak"}
	unknownRoot, err := json.Marshal(root)
	require.NoError(t, err)
	_, err = Decode(unknownRoot)
	require.ErrorContains(t, err, "unknown field")

	_, err = Decode(append(encoded, []byte(`{}`)...))
	require.ErrorContains(t, err, "must contain one valid JSON value")
}

func TestDecodeRequiresCompleteBoundedDocumentShape(t *testing.T) {
	encoded, err := Encode(validDocument(t))
	require.NoError(t, err)
	var root map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(encoded, &root))

	delete(root, "skills")
	missing, err := json.Marshal(root)
	require.NoError(t, err)
	_, err = Decode(missing)
	require.ErrorContains(t, err, `field "skills" is required`)

	require.NoError(t, json.Unmarshal(encoded, &root))
	root["skills"] = json.RawMessage(`null`)
	nullCollection, err := json.Marshal(root)
	require.NoError(t, err)
	_, err = Decode(nullCollection)
	require.ErrorContains(t, err, "collections must be arrays")

	_, err = Decode(make([]byte, MaxDocumentBytes+1))
	require.ErrorIs(t, err, ErrDocumentTooLarge)
}

func TestEncodeRejectsOversizedDocumentBeforeAgentfileParsing(t *testing.T) {
	document := validDocument(t)
	document.Worker.AgentfileSource = strings.Repeat("x", MaxDocumentBytes+1)
	document.Worker.AgentfileSourceDigest = TextDigest(
		document.Worker.AgentfileSource,
	)

	_, err := Encode(document)

	require.ErrorIs(t, err, ErrDocumentTooLarge)
}

func TestDecodeRejectsDuplicateJSONKeys(t *testing.T) {
	encoded, err := Encode(validDocument(t))
	require.NoError(t, err)
	duplicated := `{"version":1,` + string(encoded[1:])
	_, err = Decode([]byte(duplicated))
	require.Error(t, err)
}

func TestDecodeRejectsSecretValuesInsideLiveReference(t *testing.T) {
	encoded, err := Encode(validDocument(t))
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "must-not-leak")

	var root map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(encoded, &root))
	var references []map[string]any
	require.NoError(t, json.Unmarshal(root["secret_refs"], &references))
	references[0]["value"] = "must-not-leak"
	root["secret_refs"], err = json.Marshal(references)
	require.NoError(t, err)
	injected, err := json.Marshal(root)
	require.NoError(t, err)

	_, err = Decode(injected)
	require.ErrorContains(t, err, "unknown field")

	require.NoError(t, json.Unmarshal(encoded, &root))
	var repository map[string]any
	require.NoError(t, json.Unmarshal(root["repository"], &repository))
	credential := repository["credential_ref"].(map[string]any)
	credential["token"] = "must-not-leak"
	root["repository"], err = json.Marshal(repository)
	require.NoError(t, err)
	injected, err = json.Marshal(root)
	require.NoError(t, err)
	_, err = Decode(injected)
	require.ErrorContains(t, err, "unknown field")
}

func TestDigestChangesForEveryMaterializedRuntimeFact(t *testing.T) {
	baseline := clonedDocument(t, validDocument(t))
	baselineDigest, err := Digest(baseline)
	require.NoError(t, err)
	tests := []struct {
		name   string
		mutate func(*Document)
	}{
		{"model base url", func(document *Document) {
			document.Models.Primary.BaseURL = "https://new.example.com/v1"
		}},
		{"worker type", func(document *Document) {
			document.Worker.WorkerType = slugkit.MustNewForTest("cursor-cli")
		}},
		{"adapter id", func(document *Document) {
			document.Worker.AdapterID = slugkit.MustNewForTest("cursor-acp")
		}},
		{"model id", func(document *Document) {
			document.Models.Primary.ModelID = "gpt-5.1"
		}},
		{"repository commit", func(document *Document) {
			document.Repository.CommitSHA = strings.Repeat("c", 40)
		}},
		{"repository credential", func(document *Document) {
			document.Repository.Credential = patCredential(44, 7)
		}},
		{"Skill content", func(document *Document) {
			document.Skills[0].ContentDigest = TextDigest("new-skill")
		}},
		{"KnowledgeBase commit", func(document *Document) {
			document.KnowledgeBases[0].CommitSHA = strings.Repeat("d", 40)
		}},
		{"runtime bundle value", func(document *Document) {
			document.RuntimeBundles[0].Values[0].Value = "debug"
			document.RuntimeBundles[0].ContentDigest = mustRuntimeValuesDigest(
				t,
				document.RuntimeBundles[0].Values,
			)
		}},
		{"runtime image", func(document *Document) {
			digest := TextDigest("new-image")
			document.Placement.RuntimeImage.Digest = digest
			document.Placement.RuntimeImage.Reference = "registry.example.com/worker@" + digest
		}},
		{"Agentfile source", func(document *Document) {
			document.Worker.AgentfileSource = "AGENT \"codex\"\nMODE acp\n"
			document.Worker.AgentfileSourceDigest = TextDigest(
				document.Worker.AgentfileSource,
			)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changed := clonedDocument(t, baseline)
			test.mutate(&changed)
			digest, err := Digest(changed)
			require.NoError(t, err)
			assert.NotEqual(t, baselineDigest, digest)
		})
	}
}

func TestEquivalentUUIDFormsProduceOneCanonicalDigest(t *testing.T) {
	canonical := validDocument(t)
	alternate := validDocument(t)
	alternate.Repository.Pin.Reference.UID = strings.ReplaceAll(
		alternate.Repository.Pin.Reference.UID,
		"-",
		"",
	)

	canonicalDigest, err := Digest(canonical)
	require.NoError(t, err)
	alternateDigest, err := Digest(alternate)
	require.NoError(t, err)
	assert.Equal(t, canonicalDigest, alternateDigest)

	encoded, err := Encode(alternate)
	require.NoError(t, err)
	decoded, err := Decode(encoded)
	require.NoError(t, err)
	assert.Equal(
		t,
		canonical.Repository.Pin.Reference.UID,
		decoded.Repository.Pin.Reference.UID,
	)
}

func TestToolBindingUsesResourceIdentityWithoutSyntheticDomainID(t *testing.T) {
	encoded, err := Encode(validDocument(t))
	require.NoError(t, err)

	var root map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(encoded, &root))
	var models struct {
		Tools []struct {
			Binding map[string]json.RawMessage `json:"binding"`
		} `json:"tools"`
	}
	require.NoError(t, json.Unmarshal(root["models"], &models))
	require.Len(t, models.Tools, 1)
	assert.NotContains(t, models.Tools[0].Binding, "domain_id")
	assert.Contains(t, models.Tools[0].Binding, "uid")
	assert.Contains(t, models.Tools[0].Binding, "revision")
	assert.Contains(t, models.Tools[0].Binding, "digest")
}

func TestOptionalPrimaryModelAndRepositoryRemainExplicitNulls(t *testing.T) {
	document := validDocument(t)
	document.Models.Primary = nil
	document.Repository = nil
	encoded, err := Encode(document)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), `"primary":null`)
	assert.Contains(t, string(encoded), `"repository":null`)

	decoded, err := Decode(encoded)
	require.NoError(t, err)
	assert.Nil(t, decoded.Models.Primary)
	assert.Nil(t, decoded.Repository)
}

func TestUnsupportedVersionWrapsStableSentinel(t *testing.T) {
	_, err := Decode([]byte(`{"version":9}`))
	require.True(t, errors.Is(err, ErrUnsupportedVersion))
}
