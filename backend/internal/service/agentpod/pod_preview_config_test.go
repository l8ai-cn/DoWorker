package agentpod

import (
	"context"
	"testing"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/relay"
	"github.com/stretchr/testify/require"
)

func TestCreatePodPersistsCanonicalPreviewConfig(t *testing.T) {
	db := setupTestDB(t)
	service := newTestPodService(db)

	pod, err := service.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID: 1,
		RunnerID:       1,
		AgentSlug:      "codex-cli",
		CreatedByID:    1,
		PreviewPort:    3000,
		PreviewPath:    "/app//api/",
	})
	require.NoError(t, err)
	require.Equal(t, 3000, pod.PreviewPort)
	require.Equal(t, "/app/api", pod.PreviewPath)
	require.Equal(t, int64(1), pod.Generation)
	require.NotNil(t, pod.ActiveConfigRevisionID)

	var persisted podDomain.Pod
	require.NoError(t, db.First(&persisted, pod.ID).Error)
	require.Equal(t, pod.PreviewPort, persisted.PreviewPort)
	require.Equal(t, pod.PreviewPath, persisted.PreviewPath)
	require.Equal(t, pod.Generation, persisted.Generation)
	require.Equal(t, pod.ActiveConfigRevisionID, persisted.ActiveConfigRevisionID)

	var revision podDomain.PodConfigRevision
	require.NoError(t, db.Where("pod_id = ?", pod.ID).First(&revision).Error)
	require.Equal(t, pod.PreviewPort, revision.PreviewPort)
	require.Equal(t, pod.PreviewPath, revision.PreviewPath)
	require.Equal(t, revision.ID, *pod.ActiveConfigRevisionID)
	require.Equal(t, revision.ID, pod.ActiveConfigRevision.ID)
}

func TestCreatePodDefaultsDisabledPreviewPath(t *testing.T) {
	db := setupTestDB(t)
	service := newTestPodService(db)

	pod, err := service.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID: 1,
		RunnerID:       1,
		CreatedByID:    1,
	})
	require.NoError(t, err)
	require.Zero(t, pod.PreviewPort)
	require.Equal(t, "/", pod.PreviewPath)

	var revision podDomain.PodConfigRevision
	require.NoError(t, db.Where("pod_id = ?", pod.ID).First(&revision).Error)
	require.Zero(t, revision.PreviewPort)
	require.Equal(t, "/", revision.PreviewPath)
}

func TestCreatePodRejectsInvalidPreviewConfigBeforeWrite(t *testing.T) {
	tests := []struct {
		name string
		port int
		path string
		err  error
	}{
		{name: "privileged port", port: 1023, path: "/", err: relay.ErrInvalidPreviewPort},
		{name: "port above maximum", port: 65536, path: "/", err: relay.ErrInvalidPreviewPort},
		{name: "relative path", port: 3000, path: "app", err: relay.ErrInvalidPreviewPath},
		{name: "traversal", port: 3000, path: "/app/%2e%2e/admin", err: relay.ErrInvalidPreviewPath},
		{name: "query", port: 3000, path: "/app?debug=true", err: relay.ErrInvalidPreviewPath},
		{name: "fragment", port: 3000, path: "/app#debug", err: relay.ErrInvalidPreviewPath},
		{name: "invalid escape", port: 3000, path: "/bad%2", err: relay.ErrInvalidPreviewPath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			service := newTestPodService(db)

			_, err := service.CreatePod(context.Background(), &CreatePodRequest{
				OrganizationID: 1,
				RunnerID:       1,
				CreatedByID:    1,
				PreviewPort:    tt.port,
				PreviewPath:    tt.path,
			})
			require.ErrorIs(t, err, tt.err)

			var podCount int64
			require.NoError(t, db.Model(&podDomain.Pod{}).Count(&podCount).Error)
			require.Zero(t, podCount)
			var revisionCount int64
			require.NoError(t, db.Model(&podDomain.PodConfigRevision{}).Count(&revisionCount).Error)
			require.Zero(t, revisionCount)
		})
	}
}

func TestUpdatePreviewConfigCreatesImmutableNextRevision(t *testing.T) {
	db := setupTestDB(t)
	service := newTestPodService(db)
	pod, err := service.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID: 1,
		RunnerID:       1,
		CreatedByID:    1,
		AgentfileLayer: `CONFIG model = "claude-3-7-sonnet-20250219"`,
		PreviewPort:    3000,
		PreviewPath:    "/app",
	})
	require.NoError(t, err)
	originalRevisionID := *pod.ActiveConfigRevisionID

	updated, err := service.UpdatePreviewConfig(
		context.Background(), pod.PodKey, 1, 4321, "/next//api/",
	)
	require.NoError(t, err)
	require.Equal(t, 4321, updated.PreviewPort)
	require.Equal(t, "/next/api", updated.PreviewPath)
	require.Equal(t, int64(2), updated.Generation)
	require.NotNil(t, updated.ActiveConfigRevisionID)
	require.NotEqual(t, originalRevisionID, *updated.ActiveConfigRevisionID)
	require.Equal(t, updated.ActiveConfigRevisionID, &updated.ActiveConfigRevision.ID)

	var revisions []podDomain.PodConfigRevision
	require.NoError(t, db.Where("pod_id = ?", pod.ID).Order("revision").Find(&revisions).Error)
	require.Len(t, revisions, 2)
	require.Equal(t, podDomain.ConfigRevisionStatusSuperseded, revisions[0].Status)
	require.Equal(t, 3000, revisions[0].PreviewPort)
	require.Equal(t, "/app", revisions[0].PreviewPath)
	require.Equal(t, podDomain.ConfigRevisionStatusActive, revisions[1].Status)
	require.Equal(t, 4321, revisions[1].PreviewPort)
	require.Equal(t, "/next/api", revisions[1].PreviewPath)
	require.Equal(t, revisions[0].AgentfileLayer, revisions[1].AgentfileLayer)
	require.JSONEq(t, string(revisions[0].ConfigSummary), string(revisions[1].ConfigSummary))

	persisted, err := service.GetPod(context.Background(), pod.PodKey)
	require.NoError(t, err)
	require.Equal(t, updated.Generation, persisted.Generation)
	require.Equal(t, *updated.ActiveConfigRevisionID, *persisted.ActiveConfigRevisionID)
	require.Equal(t, updated.PreviewPort, persisted.PreviewPort)
	require.Equal(t, updated.PreviewPath, persisted.PreviewPath)
}

func TestUpdatePreviewConfigRejectsInvalidInputWithoutRevision(t *testing.T) {
	db := setupTestDB(t)
	service := newTestPodService(db)
	pod, err := service.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID: 1,
		RunnerID:       1,
		CreatedByID:    1,
	})
	require.NoError(t, err)

	_, err = service.UpdatePreviewConfig(context.Background(), pod.PodKey, 1, 80, "/admin")
	require.ErrorIs(t, err, relay.ErrInvalidPreviewPort)

	var revisionCount int64
	require.NoError(t, db.Model(&podDomain.PodConfigRevision{}).
		Where("pod_id = ?", pod.ID).
		Count(&revisionCount).Error)
	require.Equal(t, int64(1), revisionCount)
}
