package workerdependency

import (
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

func TestEncodeRejectsIncompleteOrUnsafeDependencies(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Document)
		message string
	}{
		{"agentfile digest mismatch", func(document *Document) {
			document.Worker.AgentfileSourceDigest = TextDigest("different")
		}, "agentfile source digest does not match"},
		{"agentfile syntax invalid", func(document *Document) {
			document.Worker.AgentfileSource = "AGENT"
			document.Worker.AgentfileSourceDigest = TextDigest("AGENT")
		}, "agentfile source is invalid"},
		{"agentfile model literal", func(document *Document) {
			document.Worker.AgentfileSource =
				"AGENT \"codex\"\nENV OPENAI_API_KEY = \"must-not-leak\"\n"
			document.Worker.AgentfileSourceDigest = TextDigest(
				document.Worker.AgentfileSource,
			)
		}, "must use a live reference"},
		{"agentfile model config literal", func(document *Document) {
			document.Worker.AgentfileSource =
				"AGENT codex\nCONFIG model STRING = \"current-model\"\n"
			document.Worker.AgentfileSourceDigest = TextDigest(
				document.Worker.AgentfileSource,
			)
		}, "must use a live reference"},
		{"worker type missing", func(document *Document) {
			document.Worker.WorkerType = ""
		}, "worker type"},
		{"runtime bundle overrides primary model field", func(document *Document) {
			document.RuntimeBundles[0].Values[0].Name = "OPENAI_BASE_URL"
			document.RuntimeBundles[0].ContentDigest = mustRuntimeValuesDigest(
				t,
				document.RuntimeBundles[0].Values,
			)
		}, "managed by a model resource"},
		{"runtime bundle overrides tool model field", func(document *Document) {
			document.RuntimeBundles[0].Values[0].Name = "SEEDANCE_BASE_URL"
			document.RuntimeBundles[0].ContentDigest = mustRuntimeValuesDigest(
				t,
				document.RuntimeBundles[0].Values,
			)
		}, "managed by a model resource"},
		{"secret target is not credential-owned", func(document *Document) {
			document.SecretReferences[0].Field = "UNDECLARED_SECRET"
		}, "not owned by a credential bundle"},
		{"secret target overrides tool model field", func(document *Document) {
			document.Worker.CredentialBundleFields = append(
				document.Worker.CredentialBundleFields,
				"SEEDANCE_API_KEY",
			)
			document.SecretReferences[0].Field = "SEEDANCE_API_KEY"
		}, "managed by a model resource"},
		{"model base url missing", func(document *Document) {
			document.Models.Primary.BaseURL = ""
		}, "model base url"},
		{"duplicate model modality", func(document *Document) {
			document.Models.Primary.Modalities = append(
				document.Models.Primary.Modalities,
				document.Models.Primary.Modalities[0],
			)
		}, "duplicate model modality"},
		{"repository commit missing", func(document *Document) {
			document.Repository.CommitSHA = ""
		}, "repository commit"},
		{"repository credential id missing", func(document *Document) {
			document.Repository.Credential = patCredential(0, 7)
		}, "repository credential id"},
		{"repository credential owner missing", func(document *Document) {
			document.Repository.Credential = patCredential(9, 0)
		}, "credential owner user id"},
		{"unauthenticated clone carries id", func(document *Document) {
			id := int64(9)
			document.Repository.Credential.CredentialID = &id
		}, "unauthenticated repository clone"},
		{"runner-local credential is not snapshot safe", func(document *Document) {
			document.Repository.Credential = RepositoryCredential{
				Type: user.CredentialTypeRunnerLocal,
			}
		}, "exact Runner secret reference"},
		{"Skill digest missing", func(document *Document) {
			document.Skills[0].ContentDigest = ""
		}, "Skill content digest"},
		{"KnowledgeBase commit missing", func(document *Document) {
			document.KnowledgeBases[0].CommitSHA = ""
		}, "KnowledgeBase commit"},
		{"credential bundle materialized", func(document *Document) {
			document.RuntimeBundles[0].Kind = envbundle.KindCredential
		}, "materialized EnvironmentBundle kind"},
		{"Secret bundle also materialized", func(document *Document) {
			document.RuntimeBundles[0].Pin.DomainID =
				document.SecretReferences[0].Pin.DomainID
		}, "Secret EnvironmentBundle cannot be materialized"},
		{"Secret org mismatch", func(document *Document) {
			document.SecretReferences[0].OwnerID = 99
		}, "Secret owner does not match"},
		{"cross namespace pin", func(document *Document) {
			document.Repository.Pin.Reference.Namespace =
				slugkit.MustNewForTest("other-team")
		}, "cross-namespace reference"},
		{"mutable image reference", func(document *Document) {
			document.Placement.RuntimeImage.Reference =
				"registry.example.com/worker:latest"
		}, "runtime image reference must match"},
		{"compute target mismatch", func(document *Document) {
			document.Placement.ComputeTarget.DomainID = 2
		}, "compute target id does not match"},
		{"config bundle lacks document metadata", func(document *Document) {
			for index := range document.RuntimeBundles {
				if document.RuntimeBundles[index].Kind == envbundle.KindConfig {
					document.RuntimeBundles[index].ConfigDocument = nil
					return
				}
			}
		}, "requires document metadata"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := clonedDocument(t, validDocument(t))
			test.mutate(&document)
			_, err := Encode(document)
			require.ErrorContains(t, err, test.message)
		})
	}
}

func TestEncodeRejectsDuplicateResourceEntries(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Document)
		message string
	}{
		{"Skill", func(document *Document) {
			document.Skills = append(document.Skills, document.Skills[0])
		}, "duplicate Skill"},
		{"KnowledgeBase", func(document *Document) {
			document.KnowledgeBases = append(
				document.KnowledgeBases,
				document.KnowledgeBases[0],
			)
		}, "duplicate KnowledgeBase"},
		{"EnvironmentBundle", func(document *Document) {
			document.RuntimeBundles = append(
				document.RuntimeBundles,
				document.RuntimeBundles[0],
			)
		}, "duplicate EnvironmentBundle"},
		{"Secret field", func(document *Document) {
			copy := document.SecretReferences[0]
			copy.Pin = testPin(
				copy.Pin.Reference.Kind,
				"other-credentials",
				705,
			)
			document.SecretReferences = append(document.SecretReferences, copy)
		}, "duplicate Secret target field"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := clonedDocument(t, validDocument(t))
			test.mutate(&document)
			_, err := Encode(document)
			require.ErrorContains(t, err, test.message)
		})
	}
}

func TestPreparationScriptMustCarryExactDigestAndTimeout(t *testing.T) {
	document := clonedDocument(t, validDocument(t))
	document.Repository.PreparationScript = "make prepare"
	document.Repository.PreparationScriptDigest = TextDigest("make something else")
	_, err := Encode(document)
	require.ErrorContains(t, err, "preparation script digest does not match")

	document = clonedDocument(t, validDocument(t))
	document.Repository.PreparationScript = ""
	document.Repository.PreparationScriptDigest = ""
	document.Repository.PreparationTimeoutSeconds = 0
	_, err = Encode(document)
	require.NoError(t, err)

	document.Repository.PreparationTimeoutSeconds = 300
	_, err = Encode(document)
	require.ErrorContains(t, err, "empty repository preparation")
}

func TestGitCommitAcceptsSHA1AndSHA256Only(t *testing.T) {
	document := clonedDocument(t, validDocument(t))
	document.Repository.CommitSHA = strings.Repeat("e", 64)
	_, err := Encode(document)
	require.NoError(t, err)

	document.Repository.CommitSHA = strings.Repeat("e", 39)
	_, err = Encode(document)
	require.ErrorContains(t, err, "Git commit SHA")
}
