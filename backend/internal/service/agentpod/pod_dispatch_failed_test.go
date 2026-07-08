package agentpod

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkDispatchFailed_FromInitializing(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestPodService(db)
	ctx := context.Background()

	pod, err := svc.CreatePod(ctx, &CreatePodRequest{OrganizationID: 1, RunnerID: 1, CreatedByID: 1})
	require.NoError(t, err)

	require.NoError(t, svc.MarkDispatchFailed(ctx, pod.PodKey, "RUNNER_UNREACHABLE", "dispatch failed"))

	updated, err := svc.GetPod(ctx, pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, agentpod.StatusError, updated.Status)
	require.NotNil(t, updated.ErrorCode)
	assert.Equal(t, "RUNNER_UNREACHABLE", *updated.ErrorCode)
}

func TestMarkDispatchFailed_FromQueued(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestPodService(db)
	ctx := context.Background()

	pod, err := svc.CreatePod(ctx, &CreatePodRequest{OrganizationID: 1, RunnerID: 1, CreatedByID: 1})
	require.NoError(t, err)
	require.NoError(t, svc.UpdatePodStatus(ctx, pod.PodKey, agentpod.StatusQueued))

	require.NoError(t, svc.MarkDispatchFailed(ctx, pod.PodKey, "QUEUE_FULL", "queue full"))

	updated, err := svc.GetPod(ctx, pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, agentpod.StatusError, updated.Status)
	require.NotNil(t, updated.ErrorCode)
	assert.Equal(t, "QUEUE_FULL", *updated.ErrorCode)
}

func TestMarkDispatchFailed_SkipsRunningPod(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestPodService(db)
	ctx := context.Background()

	pod, err := svc.CreatePod(ctx, &CreatePodRequest{OrganizationID: 1, RunnerID: 1, CreatedByID: 1})
	require.NoError(t, err)
	require.NoError(t, svc.UpdatePodStatus(ctx, pod.PodKey, agentpod.StatusRunning))

	require.NoError(t, svc.MarkDispatchFailed(ctx, pod.PodKey, "RUNNER_UNREACHABLE", "late failure"))

	updated, err := svc.GetPod(ctx, pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, agentpod.StatusRunning, updated.Status)
}
