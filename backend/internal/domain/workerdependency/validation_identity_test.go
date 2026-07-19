package workerdependency

import (
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/stretchr/testify/require"
)

func TestValidateRejectsDuplicateDomainIdentities(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Document)
		match  string
	}{
		{
			name: "Skill domain id",
			mutate: func(document *Document) {
				duplicate := document.Skills[0]
				duplicate.Pin = testPin(
					resource.KindSkill,
					"other-skill",
					document.Skills[1].Pin.DomainID,
				)
				document.Skills = append(document.Skills, duplicate)
			},
			match: "duplicate Skill domain id",
		},
		{
			name: "KnowledgeBase domain id",
			mutate: func(document *Document) {
				duplicate := document.KnowledgeBases[0]
				duplicate.Pin = testPin(
					resource.KindKnowledgeBase,
					"other-knowledge",
					document.KnowledgeBases[0].Pin.DomainID,
				)
				document.KnowledgeBases = append(document.KnowledgeBases, duplicate)
			},
			match: "duplicate KnowledgeBase domain id",
		},
		{
			name: "EnvironmentBundle domain id",
			mutate: func(document *Document) {
				duplicate := document.RuntimeBundles[0]
				duplicate.Pin = testPin(
					resource.KindEnvironmentBundle,
					"other-runtime",
					document.RuntimeBundles[1].Pin.DomainID,
				)
				document.RuntimeBundles = append(document.RuntimeBundles, duplicate)
			},
			match: "duplicate EnvironmentBundle domain id",
		},
		{
			name: "config document id",
			mutate: func(document *Document) {
				duplicate := document.RuntimeBundles[2]
				duplicate.Pin = testPin(
					resource.KindEnvironmentBundle,
					"other-config",
					705,
				)
				document.RuntimeBundles = append(document.RuntimeBundles, duplicate)
			},
			match: "duplicate config document id",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := validDocument(t)
			test.mutate(&document)

			err := Validate(document)

			require.ErrorContains(t, err, test.match)
		})
	}
}

func TestConfigBundleRequiresRuntimeJSONDocumentContract(t *testing.T) {
	tests := []struct {
		name  string
		value RuntimeValue
	}{
		{name: "missing reserved key", value: RuntimeValue{Name: "SETTINGS", Value: `{}`}},
		{
			name:  "array body",
			value: RuntimeValue{Name: envbundle.ConfigJSONDataKey, Value: `[]`},
		},
		{
			name:  "invalid body",
			value: RuntimeValue{Name: envbundle.ConfigJSONDataKey, Value: `{`},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := validDocument(t)
			document.RuntimeBundles[2].Values = []RuntimeValue{test.value}
			document.RuntimeBundles[2].ContentDigest = mustRuntimeValuesDigest(
				t,
				document.RuntimeBundles[2].Values,
			)

			err := Validate(document)

			require.ErrorContains(t, err, "config EnvironmentBundle requires")
		})
	}
}

func TestConfigBundleRejectsNonJSONFormat(t *testing.T) {
	document := validDocument(t)
	document.RuntimeBundles[2].ConfigDocument.Format = "yaml"

	err := Validate(document)

	require.ErrorContains(t, err, "config document format must be json")
}

func TestRuntimeImageRequiresNamedDigestReference(t *testing.T) {
	document := validDocument(t)
	document.Placement.RuntimeImage.Reference =
		"@" + document.Placement.RuntimeImage.Digest

	err := Validate(document)

	require.ErrorContains(t, err, "runtime image reference is invalid")
}

func TestRuntimeImageAcceptsLocalDockerDaemonDigestReference(t *testing.T) {
	document := validDocument(t)
	document.Placement.RuntimeImage.Reference =
		"docker-daemon://agentsmesh-runner-e2e-echo:latest@" +
			document.Placement.RuntimeImage.Digest

	require.NoError(t, Validate(document))
}

func TestRuntimeImageRejectsMutableLocalDockerDaemonReference(t *testing.T) {
	document := validDocument(t)
	document.Placement.RuntimeImage.Reference =
		"docker-daemon://agentsmesh-runner-e2e-echo:latest"

	err := Validate(document)

	require.ErrorContains(t, err, "runtime image reference is invalid")
}
