package agentpod

import (
	"context"
	"encoding/json"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/stretchr/testify/require"
)

func TestCreatePodPersistsExactModelResource(t *testing.T) {
	db := setupTestDB(t)
	service := newTestPodService(db)
	resourceID := int64(42)

	pod, err := service.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:  1,
		RunnerID:        1,
		AgentSlug:       "codex-cli",
		CreatedByID:     1,
		ModelResourceID: &resourceID,
	})
	require.NoError(t, err)
	require.Equal(t, &resourceID, pod.ModelResourceID)

	var revision podDomain.PodConfigRevision
	require.NoError(t, db.Where("pod_id = ?", pod.ID).First(&revision).Error)
	require.Equal(t, &resourceID, revision.ModelResourceID)

	var summary configSummary
	require.NoError(t, json.Unmarshal(revision.ConfigSummary, &summary))
	require.Equal(t, resourceID, summary.References["model_resource"].ID)
}

func TestCreatePodReturnsActiveConfigRevisionMetadata(t *testing.T) {
	db := setupTestDB(t)
	service := newTestPodService(db)
	layer := `CONFIG model = "claude-3-7-sonnet-20250219"
CONFIG base_url = "https://api.anthropic.com/v1"`

	pod, err := service.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID: 1,
		RunnerID:       1,
		AgentSlug:      "codex-cli",
		CreatedByID:    1,
		AgentfileLayer: layer,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), pod.Generation)
	require.NotNil(t, pod.ActiveConfigRevisionID)

	var revision podDomain.PodConfigRevision
	require.NoError(t, db.Where("pod_id = ?", pod.ID).First(&revision).Error)
	require.Equal(t, revision.ID, *pod.ActiveConfigRevisionID)
	require.Equal(t, podDomain.ConfigRevisionStatusActive, revision.Status)
	require.Equal(t, layer, revision.AgentfileLayer)

	var persisted podDomain.Pod
	require.NoError(t, db.First(&persisted, pod.ID).Error)
	require.Equal(t, pod.Generation, persisted.Generation)
	require.Equal(t, pod.ActiveConfigRevisionID, persisted.ActiveConfigRevisionID)
}
