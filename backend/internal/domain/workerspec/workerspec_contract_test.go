package workerspec

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerSpecV1CanonicalContract(t *testing.T) {
	spec := validWorkerSpec()
	spec.Runtime.Image.Digest = "  " + spec.Runtime.Image.Digest + "  "
	spec.Placement.ResourceProfile.Resources.GPURequest = nil
	spec.Placement.ResourceProfile.Resources.GPULimit = nil
	spec.Workspace.SkillIDs = []int64{9, 3, 7}
	spec.Workspace.KnowledgeMounts = []KnowledgeMount{
		{KnowledgeBaseID: 10, Mode: KnowledgeMountReadOnly},
		{KnowledgeBaseID: 4, Mode: KnowledgeMountReadWrite},
	}
	spec.Workspace.Instructions = "  Keep the review strict.  "
	spec.Workspace.InitialTask = "\nRun the focused tests.\n"
	spec.Metadata.Alias = "  Codex worker  "

	normalized, err := NormalizeAndValidate(spec)
	require.NoError(t, err)

	assert.Equal(t, VersionV1, normalized.Version)
	assert.Equal(t, validModelBinding(), normalized.Runtime.ModelBinding)
	assert.Equal(t, "codex-cli", normalized.Runtime.WorkerType.Slug.String())
	assert.Equal(t, validWorkerImage().Digest, normalized.Runtime.Image.Digest)
	assert.Equal(t, InteractionModeACP, normalized.TypeConfig.InteractionMode)
	assert.Equal(t, AutomationLevelAutonomous, normalized.TypeConfig.AutomationLevel)
	assert.Equal(t, []int64{3, 7, 9}, normalized.Workspace.SkillIDs)
	assert.Equal(t, []KnowledgeMount{
		{KnowledgeBaseID: 4, Mode: KnowledgeMountReadWrite},
		{KnowledgeBaseID: 10, Mode: KnowledgeMountReadOnly},
	}, normalized.Workspace.KnowledgeMounts)
	assert.Equal(t, "Keep the review strict.", normalized.Workspace.Instructions)
	assert.Equal(t, "Run the focused tests.", normalized.Workspace.InitialTask)
	assert.Equal(t, "Codex worker", normalized.Metadata.Alias)

	encoded, err := EncodeSpec(normalized)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), `"resource_id":1001`)
	assert.Contains(t, string(encoded), `"resource_revision":7`)
	assert.NotContains(t, string(encoded), "secret_value")

	summary, err := Summarize(normalized)
	require.NoError(t, err)
	assert.Equal(t, validModelBinding(), summary.ModelBinding)
	summaryJSON, err := EncodeSummary(summary)
	require.NoError(t, err)
	assert.Contains(t, string(summaryJSON), `"resource_id":1001`)
	assert.NotContains(t, string(summaryJSON), "type_config")
	assert.NotContains(t, string(summaryJSON), "secret_refs")
	assert.NotContains(t, string(summaryJSON), "instructions")
	assert.NotContains(t, string(summaryJSON), "initial_task")

	decoded, err := DecodeSpec(encoded)
	require.NoError(t, err)
	assert.Equal(t, normalized, decoded)
}

func TestNewV1IncludesResolvedPlacement(t *testing.T) {
	spec := NewV1(
		Runtime{},
		validWorkerPlacement(),
		TypeConfig{},
		Workspace{},
		Lifecycle{},
		Metadata{},
	)

	assert.Equal(t, validWorkerPlacement(), spec.Placement)
}

func TestWorkerSpecAcceptsRunnerPoolPlacement(t *testing.T) {
	spec := validWorkerSpec()
	spec.Placement.ComputeTarget.Kind = ComputeTargetKindRunnerPool
	spec.Placement.DeploymentMode = DeploymentModePooled

	_, err := NormalizeAndValidate(spec)

	require.NoError(t, err)
}

func TestWorkerSpecRejectsInvalidRuntimePlacement(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Spec)
		match  string
	}{
		{"mutable image digest", func(spec *Spec) {
			spec.Runtime.Image.Digest = "ghcr.io/example/worker:latest"
		}, "runtime image digest"},
		{"missing model resource", func(spec *Spec) {
			spec.Runtime.ModelBinding.ResourceID = 0
		}, "model binding resource"},
		{"unknown placement policy", func(spec *Spec) {
			spec.Placement.Policy = "fallback"
		}, "placement policy"},
		{"unsupported compute target", func(spec *Spec) {
			spec.Placement.ComputeTarget.Kind = "bare-metal"
		}, "compute target kind"},
		{"unknown deployment mode", func(spec *Spec) {
			spec.Placement.DeploymentMode = "serverless"
		}, "deployment mode"},
		{"oversized resource request", func(spec *Spec) {
			spec.Placement.ResourceProfile.Resources.CPURequestMilliCPU = 2000
		}, "cpu request"},
		{"partial gpu pair", func(spec *Spec) {
			spec.Placement.ResourceProfile.Resources.GPULimit = nil
		}, "gpu request and limit"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := validWorkerSpec()
			test.mutate(&spec)

			_, err := NormalizeAndValidate(spec)
			require.Error(t, err)
			assert.ErrorContains(t, err, test.match)
		})
	}
}

func TestWorkerSpecRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Spec)
		match  string
	}{
		{"values", func(spec *Spec) {
			spec.TypeConfig.Values = nil
		}, "values must be an object"},
		{"secret refs", func(spec *Spec) {
			spec.TypeConfig.SecretRefs = nil
		}, "secret refs must be an object"},
		{"interaction mode", func(spec *Spec) {
			spec.TypeConfig.InteractionMode = ""
		}, "interaction mode"},
		{"automation level", func(spec *Spec) {
			spec.TypeConfig.AutomationLevel = ""
		}, "automation level"},
		{"knowledge mount mode", func(spec *Spec) {
			spec.Workspace.KnowledgeMounts[0].Mode = ""
		}, "invalid mode"},
		{"termination policy", func(spec *Spec) {
			spec.Lifecycle.TerminationPolicy = ""
		}, "termination policy"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := validWorkerSpec()
			test.mutate(&spec)

			_, err := NormalizeAndValidate(spec)
			require.Error(t, err)
			assert.ErrorContains(t, err, test.match)
		})
	}
}

func TestWorkerSpecSnapshotIsOrganizationScopedAndSummarized(t *testing.T) {
	snapshot, err := NewSnapshot(42, validWorkerSpec())
	require.NoError(t, err)

	assert.Equal(t, int64(42), snapshot.OrganizationID)
	assert.Equal(t, VersionV1, snapshot.Summary.Version)
	assert.Equal(t, snapshot.Spec.Runtime.WorkerType, snapshot.Summary.WorkerType)
	assert.Equal(t, snapshot.Spec.Runtime.Image, snapshot.Summary.RuntimeImage)
	assert.Equal(t, snapshot.Spec.Placement, snapshot.Summary.Placement)

	_, err = NewSnapshot(0, validWorkerSpec())
	assert.True(t, errors.Is(err, ErrInvalidOrganizationID))
}

func TestWorkerSpecCodecRejectsUnknownAndUnsupportedDocuments(t *testing.T) {
	_, err := DecodeSpec([]byte(`{"version":2}`))
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedVersion))

	encoded, err := EncodeSpec(validWorkerSpec())
	require.NoError(t, err)

	var document map[string]any
	require.NoError(t, json.Unmarshal(encoded, &document))
	document["fallback"] = true
	encoded, err = json.Marshal(document)
	require.NoError(t, err)

	_, err = DecodeSpec(encoded)
	require.Error(t, err)
	assert.ErrorContains(t, err, "unknown field")
}

func validWorkerSpec() Spec {
	spec := NewV1(
		Runtime{
			ModelBinding: validModelBinding(),
			WorkerType: WorkerType{
				Slug:           slugkit.MustNewForTest("codex-cli"),
				DefinitionHash: strings.Repeat("a", 64),
			},
			Image: validWorkerImage(),
		},
		validWorkerPlacement(),
		TypeConfig{
			SchemaVersion: 1,
			Values: map[string]any{
				"reasoning_effort": "high",
			},
			SecretRefs: map[string]SecretReference{
				"api_token": {
					Kind: slugkit.MustNewForTest("vault-secret"),
					ID:   91,
				},
			},
			InteractionMode: InteractionModeACP,
			AutomationLevel: AutomationLevelAutonomous,
		},
		Workspace{
			RepositoryID: int64PointerForWorkerSpecTest(22),
			Branch:       "main",
			SkillIDs:     []int64{5, 3},
			KnowledgeMounts: []KnowledgeMount{
				{KnowledgeBaseID: 7, Mode: KnowledgeMountReadOnly},
			},
			EnvBundleIDs: []RuntimeEnvBundleID{10, 11},
			Instructions: "Review before editing.",
			InitialTask:  "Fix the failing test.",
		},
		Lifecycle{
			TerminationPolicy:  TerminationPolicyOnIdle,
			IdleTimeoutMinutes: 120,
		},
		Metadata{
			Alias:          "API worker",
			SourceExpertID: int64PointerForWorkerSpecTest(31),
		},
	)
	return spec
}

func validModelBinding() ModelBinding {
	return ModelBinding{
		ResourceID:         1001,
		ResourceRevision:   7,
		ConnectionID:       2001,
		ConnectionRevision: 9,
		ProviderKey:        slugkit.MustNewForTest("openai"),
		ProtocolAdapter:    slugkit.MustNewForTest("openai-compatible"),
		ModelID:            "gpt-5",
	}
}

func validWorkerImage() RuntimeImage {
	return RuntimeImage{
		ID:     41,
		Digest: "sha256:" + strings.Repeat("a", 64),
	}
}

func validWorkerPlacement() Placement {
	gpuRequest := uint32(1)
	gpuLimit := uint32(2)
	return Placement{
		Policy: PlacementPolicyAutomatic,
		ComputeTarget: ComputeTarget{
			ID:   52,
			Kind: ComputeTargetKindKubernetes,
		},
		DeploymentMode: DeploymentModeDedicated,
		ResourceProfile: ResourceProfile{
			ID: 63,
			Resources: ResourceRequestsLimits{
				CPURequestMilliCPU: 500,
				CPULimitMilliCPU:   1000,
				MemoryRequestBytes: 536870912,
				MemoryLimitBytes:   1073741824,
				GPURequest:         &gpuRequest,
				GPULimit:           &gpuLimit,
			},
		},
	}
}

func int64PointerForWorkerSpecTest(value int64) *int64 {
	return &value
}
