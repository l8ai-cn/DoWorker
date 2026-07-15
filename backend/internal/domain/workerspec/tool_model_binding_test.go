package workerspec

import (
	"testing"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerSpecPreservesToolModelBindingSnapshot(t *testing.T) {
	spec := validWorkerSpec()
	spec.Runtime.ToolModelBindings = []ToolModelBinding{validVideoToolBinding()}

	normalized, err := NormalizeAndValidate(spec)
	require.NoError(t, err)
	summary, err := Summarize(normalized)
	require.NoError(t, err)

	assert.Equal(t, spec.Runtime.ToolModelBindings, normalized.Runtime.ToolModelBindings)
	assert.Equal(t, normalized.Runtime.ToolModelBindings, summary.ToolModelBindings)
}

func TestWorkerSpecRejectsDuplicateToolModelRole(t *testing.T) {
	spec := validWorkerSpec()
	binding := validVideoToolBinding()
	spec.Runtime.ToolModelBindings = []ToolModelBinding{binding, binding}

	_, err := NormalizeAndValidate(spec)

	require.Error(t, err)
	assert.ErrorContains(t, err, "duplicate tool model role")
}

func TestWorkerSpecRejectsDuplicateToolModelEnvironmentTarget(t *testing.T) {
	spec := validWorkerSpec()
	first := validVideoToolBinding()
	second := validVideoToolBinding()
	second.Role = slugkit.MustNewForTest("other-video")
	second.Environment.ModelID = first.Environment.APIKey
	spec.Runtime.ToolModelBindings = []ToolModelBinding{first, second}

	_, err := NormalizeAndValidate(spec)

	require.Error(t, err)
	assert.ErrorContains(t, err, "duplicate tool model environment target")
}

func TestWorkerSpecRejectsEmptyToolModelBinding(t *testing.T) {
	spec := validWorkerSpec()
	binding := validVideoToolBinding()
	binding.ModelBinding = ModelBinding{}
	spec.Runtime.ToolModelBindings = []ToolModelBinding{binding}

	_, err := NormalizeAndValidate(spec)

	assert.ErrorContains(t, err, "binding is required")
}

func validVideoToolBinding() ToolModelBinding {
	binding := validModelBinding()
	binding.ResourceID = 2001
	binding.ProviderKey = slugkit.MustNewForTest("openai")
	binding.ModelID = "video-model-v1"
	return ToolModelBinding{
		Role:         slugkit.MustNewForTest("video-generator"),
		ModelBinding: binding,
		Modality:     resourcedomain.ModalityVideo,
		Capability:   resourcedomain.CapabilityVideoGeneration,
		Environment: ToolModelEnvironment{
			APIKey: "VIDEO_MODEL_API_KEY", BaseURL: "VIDEO_MODEL_BASE_URL",
			ModelID: "VIDEO_MODEL_ID",
		},
	}
}
