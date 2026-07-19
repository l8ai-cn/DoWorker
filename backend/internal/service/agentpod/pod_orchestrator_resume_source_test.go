package agentpod

import (
	"context"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createImmutableResumeSource(
	t *testing.T,
	orchestrator *PodOrchestrator,
	podService *PodService,
	db *gorm.DB,
	req *OrchestrateCreatePodRequest,
) *podDomain.Pod {
	t.Helper()
	if req.RunnerID == 0 {
		req.RunnerID = 1
	}
	result, err := createPodWithPlanSourceForTest(
		t,
		orchestrator,
		context.Background(),
		req,
	)
	require.NoError(t, err)
	require.NotNil(t, result.Pod.WorkerSpecSnapshotID)
	require.Positive(t, *result.Pod.WorkerSpecSnapshotID)
	require.NoError(
		t,
		db.Model(&podDomain.Pod{}).
			Where("pod_key = ?", result.Pod.PodKey).
			Update("status", podDomain.StatusTerminated).Error,
	)
	source, err := podService.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	return source
}

func updateResumeSource(
	t *testing.T,
	podService *PodService,
	db *gorm.DB,
	podKey string,
	updates map[string]interface{},
) *podDomain.Pod {
	t.Helper()
	require.NoError(
		t,
		db.Model(&podDomain.Pod{}).
			Where("pod_key = ?", podKey).
			Updates(updates).Error,
	)
	pod, err := podService.GetPod(context.Background(), podKey)
	require.NoError(t, err)
	return pod
}
