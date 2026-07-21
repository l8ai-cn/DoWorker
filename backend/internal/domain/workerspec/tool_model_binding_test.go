package workerspec

import (
	"testing"

	resourcedomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerSpecPreservesToolModelBindingSnapshot(t *testing.T) {
	spec := validWorkerSpec()
	spec.Runtime.ToolModelBindings = []ToolModelBinding{validSeedanceToolBinding()}

	normalized, err := NormalizeAndValidate(spec)
	require.NoError(t, err)
	summary, err := Summarize(normalized)
	require.NoError(t, err)

	assert.Equal(t, spec.Runtime.ToolModelBindings, normalized.Runtime.ToolModelBindings)
	assert.Equal(t, normalized.Runtime.ToolModelBindings, summary.ToolModelBindings)
}

func TestWorkerSpecRejectsDuplicateToolModelRole(t *testing.T) {
	spec := validWorkerSpec()
	binding := validSeedanceToolBinding()
	spec.Runtime.ToolModelBindings = []ToolModelBinding{binding, binding}

	_, err := NormalizeAndValidate(spec)

	require.Error(t, err)
	assert.ErrorContains(t, err, "duplicate tool model role")
}

func TestWorkerSpecRejectsDuplicateToolModelEnvironmentTarget(t *testing.T) {
	spec := validWorkerSpec()
	first := validSeedanceToolBinding()
	second := validSeedanceToolBinding()
	second.Role = slugkit.MustNewForTest("other-video")
	second.Environment.ModelID = first.Environment.APIKey
	spec.Runtime.ToolModelBindings = []ToolModelBinding{first, second}

	_, err := NormalizeAndValidate(spec)

	require.Error(t, err)
	assert.ErrorContains(t, err, "duplicate tool model environment target")
}

func validSeedanceToolBinding() ToolModelBinding {
	binding := validModelBinding()
	binding.ResourceID = 2001
	binding.ProviderKey = slugkit.MustNewForTest("doubao")
	binding.ModelID = "doubao-seedance-2-0-260128"
	return ToolModelBinding{
		Role:         slugkit.MustNewForTest("seedance-video"),
		ModelBinding: binding,
		Modality:     resourcedomain.ModalityVideo,
		Capability:   resourcedomain.CapabilityVideoGeneration,
		Environment: ToolModelEnvironment{
			APIKey: "SEEDANCE_API_KEY", BaseURL: "SEEDANCE_BASE_URL",
			ModelID: "SEEDANCE_MODEL",
		},
	}
}
