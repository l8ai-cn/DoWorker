package workercreation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/agentfile/parser"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompilerBuildsDeterministicAgentfileLayerFromResolvedSpec(t *testing.T) {
	fixture := newWorkspaceFixture()
	compiler := newCompiler(newWorkspaceResolver(fixture.deps()))
	spec := validCompiledWorkerSpec()

	layer, err := compiler.Compile(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		spec,
	)

	require.NoError(t, err)
	_, parseErrors := parser.Parse(layer)
	require.Empty(t, parseErrors)
	assert.Equal(t, `CONFIG approval_mode = "never"
CONFIG retries = 3
MODE acp
REPO "org/repo"
BRANCH "main"
SKILLS code-review
KNOWLEDGE engineering-docs [rw]
USE_ENV_BUNDLE "runtime-preferences"
USE_ENV_BUNDLE "signing-secrets"
PROMPT "Review before editing.\n\nFix the failing test."
`, layer)
	assert.NotContains(t, layer, "must-not-leak")
	assert.Less(t, strings.Index(layer, "CONFIG approval_mode"), strings.Index(layer, "CONFIG retries"))
}

func TestCompilerOmitsEmptyOptionalDeclarations(t *testing.T) {
	fixture := newWorkspaceFixture()
	compiler := newCompiler(newWorkspaceResolver(fixture.deps()))
	spec := validCompiledWorkerSpec()
	spec.TypeConfig.Values = map[string]any{}
	spec.TypeConfig.SecretRefs = map[string]specdomain.SecretReference{}
	spec.Workspace = specdomain.Workspace{
		SkillIDs:        []int64{},
		KnowledgeMounts: []specdomain.KnowledgeMount{},
		EnvBundleIDs:    []specdomain.RuntimeEnvBundleID{},
		InitialTask:     "Run checks.",
	}

	layer, err := compiler.Compile(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		spec,
	)

	require.NoError(t, err)
	assert.Equal(t, "MODE acp\nPROMPT \"Run checks.\"\n", layer)
}

func validCompiledWorkerSpec() specdomain.Spec {
	return specdomain.NewV1(
		specdomain.Runtime{
			ModelBinding: specdomain.ModelBinding{
				ResourceID:         101,
				ResourceRevision:   7,
				ConnectionID:       201,
				ConnectionRevision: 9,
				ProviderKey:        slugkit.MustNewForTest("openai"),
				ProtocolAdapter:    slugkit.MustNewForTest("openai-compatible"),
				ModelID:            "gpt-5",
			},
			WorkerType: specdomain.WorkerType{
				Slug:           slugkit.MustNewForTest("codex-cli"),
				DefinitionHash: workspaceTestDefinitions()["codex-cli"].DefinitionHash,
			},
			Image: specdomain.RuntimeImage{
				ID:     1,
				Digest: "sha256:" + strings.Repeat("b", 64),
			},
		},
		specdomain.Placement{
			Policy: specdomain.PlacementPolicyExplicit,
			ComputeTarget: specdomain.ComputeTarget{
				ID:   1,
				Kind: specdomain.ComputeTargetKindRunnerPool,
			},
			DeploymentMode: specdomain.DeploymentModePooled,
			ResourceProfile: specdomain.ResourceProfile{
				ID: 1,
				Resources: specdomain.ResourceRequestsLimits{
					CPURequestMilliCPU: 200,
					CPULimitMilliCPU:   1000,
					MemoryRequestBytes: 256 << 20,
					MemoryLimitBytes:   1 << 30,
				},
			},
		},
		specdomain.TypeConfig{
			SchemaVersion: 1,
			Values: map[string]any{
				"retries":       json.Number("3"),
				"approval_mode": "never",
			},
			SecretRefs: map[string]specdomain.SecretReference{
				"SIGNING_KEY": {
					Kind: slugkit.MustNewForTest("env-bundle"),
					ID:   6,
				},
			},
			InteractionMode: specdomain.InteractionModeACP,
			AutomationLevel: specdomain.AutomationLevelAutonomous,
		},
		validWorkspaceDraft(),
		specdomain.Lifecycle{TerminationPolicy: specdomain.TerminationPolicyManual},
		specdomain.Metadata{Alias: "worker"},
	)
}
