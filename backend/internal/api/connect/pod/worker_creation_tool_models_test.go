package podconnect

import (
	"testing"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerDraftFromProtoPreservesToolModelSelections(t *testing.T) {
	message := completeWorkerDraftProto()
	message.ToolModelResourceIds = map[string]int64{"seedance-video": 202}

	draft, err := workerDraftFromProto(message)

	require.NoError(t, err)
	assert.Equal(t, map[string]int64{"seedance-video": 202}, draft.WorkerSpec.ToolModelResourceIDs)
}

func TestWorkerCreateOptionsExposeToolModelRequirements(t *testing.T) {
	options := workercreation.CreateOptions{
		Revision: "revision",
		WorkerTypes: []workercreation.WorkerTypeOption{
			{
				Slug: "seedance-expert", Name: "Seedance Expert",
				ToolModelRequirements: []specdomain.ToolModelRequirement{
					{
						Role:             slugkit.MustNewForTest("seedance-video"),
						ProviderKeys:     []slugkit.Slug{slugkit.MustNewForTest("doubao")},
						ProtocolAdapters: []slugkit.Slug{slugkit.MustNewForTest("openai-compatible")},
						Modality:         resourcedomain.ModalityVideo,
						Capability:       resourcedomain.CapabilityVideoGeneration,
					},
				},
			},
		},
	}

	response, err := workerCreateOptionsToProto(options)

	require.NoError(t, err)
	require.Len(t, response.WorkerTypes, 1)
	require.Len(t, response.WorkerTypes[0].ToolModelRequirements, 1)
	assert.Equal(t, "seedance-video", response.WorkerTypes[0].ToolModelRequirements[0].Role)
	assert.Equal(t, []string{"doubao"}, response.WorkerTypes[0].ToolModelRequirements[0].ProviderKeys)
}

var _ = podv1.WorkerSpecDraft{}
