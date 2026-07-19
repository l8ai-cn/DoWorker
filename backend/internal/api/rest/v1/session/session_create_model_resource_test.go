package sessionapi

import (
	"context"
	"encoding/json"
	"testing"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSessionBodyAcceptsModelResourceID(t *testing.T) {
	var body createSessionBody

	err := json.Unmarshal([]byte(`{"agent_id":"do-agent","model_resource_id":42}`), &body)

	require.NoError(t, err)
	require.NotNil(t, body.ModelResourceID)
	assert.Equal(t, int64(42), *body.ModelResourceID)
}

func TestLegacySessionCreateModelFieldsAreRejected(t *testing.T) {
	for _, field := range []string{
		"credential" + "_profile_id",
		"model",
		"model" + "_config_id",
		"virtual_api" + "_key_id",
	} {
		t.Run(field, func(t *testing.T) {
			got, ok := legacySessionCreateModelField([]byte(`{"agent_id":"do-agent","` + field + `":99}`))

			require.True(t, ok)
			assert.Equal(t, field, got)
		})
	}
}

func TestSessionCreatePodRequestBuildsPlanSource(t *testing.T) {
	resourceID := int64(42)
	layer := "MODE acp"
	factory := &recordingSessionWorkerDraftFactory{
		draft: workercreation.Draft{
			OptionsRevision: "rev-1",
			WorkerSpec: specservice.Draft{
				WorkerTypeSlug:  testSlug(t, "do-agent"),
				ModelResourceID: resourceID,
			},
		},
	}
	deps := &Deps{WorkerCreation: factory}

	req, err := deps.sessionCreatePodRequest(context.Background(), 11, 21, "dev-org", createSessionBody{
		AgentID:         "do-agent",
		ModelResourceID: &resourceID,
		AutomationLevel: "autonomous",
		WorkerSpec: &sessionWorkerSpecBody{
			OptionsRevision:   "rev-1",
			RuntimeImageID:    4,
			PlacementPolicy:   "explicit",
			ComputeTargetID:   8,
			DeploymentMode:    "pooled",
			ResourceProfileID: 9,
		},
	}, &layer, "/tmp/workspace")

	require.NoError(t, err)
	assert.Equal(t, int64(11), req.UserID)
	assert.Equal(t, int64(21), req.OrganizationID)
	assert.Empty(t, req.AgentSlug)
	assert.Nil(t, req.ModelResourceID)
	require.NotNil(t, req.WorkerSpecDraft)
	assert.Nil(t, req.SessionConfigBundles)
	assert.Empty(t, req.ModelResourceEnv)
	assert.Equal(t, "/tmp/workspace", req.LocalPath)
	assert.Equal(t, "do-agent", factory.input.WorkerTypeSlug)
	assert.Equal(t, int64(4), factory.input.Runtime.RuntimeImageID)
	assert.Equal(t, specdomain.DeploymentModePooled, factory.input.Runtime.DeploymentMode)
	assert.Equal(t, specdomain.AutomationLevelAutonomous, factory.input.AutomationLevel)
	assert.Equal(t, "dev-org", factory.input.OrganizationSlug)
}

func TestSessionCreatePodRequestRequiresWorkerSpec(t *testing.T) {
	req, err := (&Deps{WorkerCreation: &recordingSessionWorkerDraftFactory{}}).
		sessionCreatePodRequest(context.Background(), 11, 21, "dev-org", createSessionBody{
			AgentID: "do-agent",
		}, nil, "")

	require.Error(t, err)
	assert.Nil(t, req)
	assert.Equal(t, "worker_spec", specservice.InvalidDraftField(err))
}

func TestSessionCreatePodRequestRequiresExplicitAutomationLevel(t *testing.T) {
	req, err := (&Deps{WorkerCreation: &recordingSessionWorkerDraftFactory{}}).
		sessionCreatePodRequest(context.Background(), 11, 21, "dev-org", createSessionBody{
			AgentID:    "do-agent",
			WorkerSpec: validSessionWorkerSpecBody(),
		}, nil, "")

	require.Error(t, err)
	assert.Nil(t, req)
	assert.Equal(t, "automation_level", specservice.InvalidDraftField(err))
}

func validSessionWorkerSpecBody() *sessionWorkerSpecBody {
	return &sessionWorkerSpecBody{
		OptionsRevision:   "rev-1",
		RuntimeImageID:    4,
		PlacementPolicy:   "explicit",
		ComputeTargetID:   8,
		DeploymentMode:    "pooled",
		ResourceProfileID: 9,
	}
}

func sessionTestWorkerDraft(t *testing.T, workerType string) workercreation.Draft {
	t.Helper()
	return workercreation.Draft{
		OptionsRevision:  "rev-1",
		OrganizationSlug: testSlug(t, "dev-org"),
		WorkerSpec: specservice.Draft{
			WorkerTypeSlug: testSlug(t, workerType),
		},
	}
}

type recordingSessionWorkerDraftFactory struct {
	input workercreation.FreshPodDraftInput
	draft workercreation.Draft
	err   error
}

func (factory *recordingSessionWorkerDraftFactory) NewFreshPodDraft(
	_ context.Context,
	_ specservice.Scope,
	input workercreation.FreshPodDraftInput,
) (workercreation.Draft, error) {
	factory.input = input
	return factory.draft, factory.err
}

func testSlug(t *testing.T, value string) slugkit.Slug {
	t.Helper()
	slug, err := slugkit.NewFromTrusted(value)
	require.NoError(t, err)
	return slug
}
