package workerdependency

import (
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func validDocument(t *testing.T) Document {
	t.Helper()
	runtimeValues := []RuntimeValue{
		{Name: "LOG_LEVEL", Value: "info"},
		{Name: "ANTHROPIC_BASE_URL", Value: "https://proxy.example.com"},
	}
	configValues := []RuntimeValue{
		{Name: envbundle.ConfigJSONDataKey, Value: `{"approval_mode":"never"}`},
	}
	overlayValues := []RuntimeValue{
		{Name: "LOG_LEVEL", Value: "debug"},
	}
	document := Document{
		Version:        VersionV1,
		OrganizationID: 7,
		Namespace:      slugkit.MustNewForTest("team-alpha"),
		Worker: Worker{
			WorkerType:             slugkit.MustNewForTest("codex-cli"),
			AdapterID:              slugkit.MustNewForTest("codex-app-server"),
			SpecVersion:            workerspec.VersionV1,
			SpecDigest:             TextDigest("workerspec"),
			DefinitionHash:         strings.Repeat("1", 64),
			ModelManagedFields:     []string{"model", "OPENAI_MODEL", "OPENAI_API_KEY", "OPENAI_BASE_URL"},
			CredentialBundleFields: []string{"CURSOR_API_KEY"},
			AgentfileSource: "AGENT codex\nCONFIG model STRING = \"\"\n" +
				"ENV OPENAI_API_KEY SECRET OPTIONAL\n" +
				"ENV CURSOR_API_KEY SECRET OPTIONAL\nMODE pty\n",
		},
		Models: Models{
			Primary: &Model{
				Pin:                testPin(resource.KindModelBinding, "coding-model", 101),
				ResourceRevision:   7,
				ConnectionID:       201,
				ConnectionRevision: 9,
				ProviderKey:        slugkit.MustNewForTest("openai"),
				ProtocolAdapter:    slugkit.MustNewForTest("openai-compatible"),
				ModelID:            "gpt-5",
				BaseURL:            "https://api.example.com/v1",
				Modalities: []airesource.Modality{
					airesource.ModalityMultimodal,
					airesource.ModalityChat,
				},
				Capabilities: []airesource.Capability{
					airesource.CapabilityVisionInput,
					airesource.CapabilityTextGeneration,
				},
			},
			Tools: []ToolModel{{
				Binding: testPin(
					resource.KindToolBinding,
					"video-tool",
					501,
				).Reference,
				Role: slugkit.MustNewForTest("video-generation"),
				Model: Model{
					Pin:                testPin(resource.KindModelBinding, "video-model", 102),
					ResourceRevision:   3,
					ConnectionID:       202,
					ConnectionRevision: 4,
					ProviderKey:        slugkit.MustNewForTest("doubao"),
					ProtocolAdapter:    slugkit.MustNewForTest("openai-compatible"),
					ModelID:            "seedance-1-0-pro",
					BaseURL:            "https://ark.example.com/api/v3",
					Modalities:         []airesource.Modality{airesource.ModalityVideo},
					Capabilities: []airesource.Capability{
						airesource.CapabilityVideoGeneration,
					},
				},
				Modality:   airesource.ModalityVideo,
				Capability: airesource.CapabilityVideoGeneration,
				Environment: ToolModelEnvironment{
					APIKeyTarget:  "SEEDANCE_API_KEY",
					BaseURLTarget: "SEEDANCE_BASE_URL",
					ModelIDTarget: "SEEDANCE_MODEL_ID",
				},
			}},
		},
		Repository: &Repository{
			Pin:                       testPin(resource.KindRepository, "agents-mesh", 301),
			HTTPCloneURL:              "https://git.example.com/agents-mesh.git",
			SSHCloneURL:               "git@git.example.com:agents-mesh.git",
			Branch:                    "main",
			CommitSHA:                 strings.Repeat("a", 40),
			Credential:                unauthenticatedCredential(),
			PreparationScript:         "pnpm install --frozen-lockfile",
			PreparationScriptDigest:   TextDigest("pnpm install --frozen-lockfile"),
			PreparationTimeoutSeconds: 300,
		},
		Skills: []Skill{
			{
				Pin:           testPin(resource.KindSkill, "review", 402),
				Slug:          slugkit.MustNewForTest("review"),
				Version:       2,
				ContentDigest: TextDigest("review-package"),
				StorageKey:    "skills/review.tgz",
				PackageSize:   2048,
			},
			{
				Pin:           testPin(resource.KindSkill, "build", 401),
				Slug:          slugkit.MustNewForTest("build"),
				Version:       1,
				ContentDigest: TextDigest("build-package"),
				StorageKey:    "skills/build.tgz",
				PackageSize:   1024,
			},
		},
		KnowledgeBases: []KnowledgeBase{{
			Pin:          testPin(resource.KindKnowledgeBase, "engineering-docs", 601),
			Slug:         slugkit.MustNewForTest("engineering-docs"),
			HTTPCloneURL: "https://git.example.com/knowledge/engineering-docs.git",
			Branch:       "main",
			CommitSHA:    strings.Repeat("b", 40),
			Mode:         workerspec.KnowledgeMountReadOnly,
		}},
		RuntimeBundles: []RuntimeBundle{
			{
				Pin:           testPin(resource.KindEnvironmentBundle, "runtime", 701),
				Kind:          envbundle.KindRuntime,
				ContentDigest: mustRuntimeValuesDigest(t, runtimeValues),
				Values:        runtimeValues,
			},
			{
				Pin:           testPin(resource.KindEnvironmentBundle, "runtime-overlay", 704),
				Kind:          envbundle.KindShared,
				ContentDigest: mustRuntimeValuesDigest(t, overlayValues),
				Values:        overlayValues,
			},
			{
				Pin:           testPin(resource.KindEnvironmentBundle, "config", 702),
				Kind:          envbundle.KindConfig,
				ContentDigest: mustRuntimeValuesDigest(t, configValues),
				Values:        configValues,
				ConfigDocument: &ConfigDocument{
					ID: "settings", Format: "json", TargetPath: ".codex/settings.json",
				},
			},
		},
		SecretReferences: []SecretReference{{
			Pin:        testPin(resource.KindEnvironmentBundle, "credentials", 703),
			Field:      "CURSOR_API_KEY",
			BundleKey:  "CURSOR_API_KEY",
			OwnerScope: envbundle.OwnerScopeOrg,
			OwnerID:    7,
		}},
		Placement: validPlacement(),
	}
	document.Worker.AgentfileSourceDigest = TextDigest(document.Worker.AgentfileSource)
	return document
}

func unauthenticatedCredential() RepositoryCredential {
	return RepositoryCredential{Type: RepositoryCredentialTypeNone}
}

func patCredential(id, ownerUserID int64) RepositoryCredential {
	return RepositoryCredential{
		Type: user.CredentialTypePAT, CredentialID: &id, OwnerUserID: ownerUserID,
	}
}

func validPlacement() Placement {
	imageDigest := TextDigest("worker-image")
	profile := testPin(resource.KindResourceProfile, "standard", 1)
	return Placement{
		CatalogRevision: "catalog-v1",
		RuntimeImage: RuntimeImage{
			ID:        1,
			Reference: "registry.example.com/worker@" + imageDigest,
			Digest:    imageDigest,
		},
		ComputeTarget:   testPin(resource.KindComputeTarget, "runner-pool", 1),
		ResourceProfile: &profile,
		Spec: workerspec.Placement{
			Policy:         workerspec.PlacementPolicyExplicit,
			ComputeTarget:  workerspec.ComputeTarget{ID: 1, Kind: workerspec.ComputeTargetKindRunnerPool},
			DeploymentMode: workerspec.DeploymentModePooled,
			ResourceProfile: workerspec.ResourceProfile{
				ID: 1,
				Resources: workerspec.ResourceRequestsLimits{
					CPURequestMilliCPU: 200, CPULimitMilliCPU: 1000,
					MemoryRequestBytes: 256 << 20, MemoryLimitBytes: 1 << 30,
					StorageRequestBytes: 10 << 30, StorageLimitBytes: 10 << 30,
				},
			},
		},
	}
}

func testPin(kind, name string, domainID int64) ResourcePin {
	identity := kind + "\x00" + name
	return ResourcePin{
		DomainID: domainID,
		Reference: resource.Reference{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       kind,
			Namespace:  slugkit.MustNewForTest("team-alpha"),
			Name:       slugkit.MustNewForTest(name),
			UID:        uuid.NewSHA1(uuid.NameSpaceOID, []byte(identity)).String(),
			Revision:   1,
			Digest:     TextDigest(identity),
		},
	}
}

func mustRuntimeValuesDigest(t *testing.T, values []RuntimeValue) string {
	t.Helper()
	digest, err := DigestRuntimeValues(values)
	require.NoError(t, err)
	return digest
}

func clonedDocument(t *testing.T, document Document) Document {
	t.Helper()
	encoded, err := Encode(document)
	require.NoError(t, err)
	decoded, err := Decode(encoded)
	require.NoError(t, err)
	return decoded
}
