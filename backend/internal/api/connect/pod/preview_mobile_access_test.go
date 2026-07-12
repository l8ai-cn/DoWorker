package podconnect

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	agentpodservice "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/service/relay"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
	"gorm.io/gorm"
)

func TestUpdatePodPreviewConfigPersistsNormalizedRevision(t *testing.T) {
	server, service, db, pod := newPreviewMobileServer(t, true)
	response, err := server.UpdatePodPreviewConfig(
		ctxAsUser(42),
		connect.NewRequest(&podv1.UpdatePodPreviewConfigRequest{
			OrgSlug:     "acme",
			PodKey:      pod.PodKey,
			PreviewPort: 4321,
			PreviewPath: "/next//api/",
		}),
	)
	require.NoError(t, err)
	require.Equal(t, int32(4321), response.Msg.Pod.GetPreviewPort())
	require.Equal(t, "/next/api", response.Msg.Pod.GetPreviewPath())

	persisted, err := service.GetPod(context.Background(), pod.PodKey)
	require.NoError(t, err)
	require.Equal(t, int64(2), persisted.Generation)
	require.Equal(t, 4321, persisted.PreviewPort)
	require.Equal(t, "/next/api", persisted.PreviewPath)

	var revisions []agentpod.PodConfigRevision
	require.NoError(t, db.Where("pod_id = ?", pod.ID).Order("revision").Find(&revisions).Error)
	require.Len(t, revisions, 2)
	require.Equal(t, agentpod.ConfigRevisionStatusSuperseded, revisions[0].Status)
	require.Equal(t, agentpod.ConfigRevisionStatusActive, revisions[1].Status)
}

func TestUpdatePodPreviewConfigRejectsUnauthorizedMember(t *testing.T) {
	server, _, db, pod := newPreviewMobileServer(t, true)
	_, err := server.UpdatePodPreviewConfig(
		ctxAsUser(99),
		connect.NewRequest(&podv1.UpdatePodPreviewConfigRequest{
			OrgSlug: "acme", PodKey: pod.PodKey, PreviewPort: 4321, PreviewPath: "/",
		}),
	)
	require.Equal(t, connect.CodePermissionDenied, connectCodeOf(t, err))

	var revisionCount int64
	require.NoError(t, db.Model(&agentpod.PodConfigRevision{}).
		Where("pod_id = ?", pod.ID).
		Count(&revisionCount).Error)
	require.Equal(t, int64(1), revisionCount)
}

func TestGetMobileAccessDescriptorReturnsTokenFreeCanonicalURL(t *testing.T) {
	server, _, _, pod := newPreviewMobileServer(t, true)
	response, err := server.GetMobileAccessDescriptor(
		ctxAsUser(42),
		connect.NewRequest(&podv1.GetMobileAccessDescriptorRequest{
			OrgSlug: "acme", PodKey: pod.PodKey,
		}),
	)
	require.NoError(t, err)
	require.Equal(t, "https://mobile.example/workers/"+pod.PodKey, response.Msg.CanonicalUrl)
	require.NotContains(t, response.Msg.CanonicalUrl, "token")
	require.True(t, response.Msg.ConsoleAvailable)
	require.True(t, response.Msg.PreviewAvailable)
	require.True(t, response.Msg.RelayAvailable)
	require.Equal(t, agentpod.InteractionModePTY, response.Msg.InteractionMode)
}

func TestGetMobileAccessDescriptorReportsUnavailableCapabilities(t *testing.T) {
	server, service, _, pod := newPreviewMobileServer(t, false)
	require.NoError(t, service.UpdatePodStatus(context.Background(), pod.PodKey, agentpod.StatusTerminated))

	response, err := server.GetMobileAccessDescriptor(
		ctxAsUser(42),
		connect.NewRequest(&podv1.GetMobileAccessDescriptorRequest{
			OrgSlug: "acme", PodKey: pod.PodKey,
		}),
	)
	require.NoError(t, err)
	require.False(t, response.Msg.ConsoleAvailable)
	require.False(t, response.Msg.PreviewAvailable)
	require.False(t, response.Msg.RelayAvailable)
}

func TestGetMobileAccessDescriptorRequiresPublicBaseURL(t *testing.T) {
	server := NewServer(nil, &fakeOrgService{role: "member"})
	_, err := server.GetMobileAccessDescriptor(
		ctxAsUser(42),
		connect.NewRequest(&podv1.GetMobileAccessDescriptorRequest{
			OrgSlug: "acme", PodKey: "pod-1",
		}),
	)
	require.Equal(t, connect.CodeUnavailable, connectCodeOf(t, err))
}

func newPreviewMobileServer(
	t *testing.T,
	withRelay bool,
) (*Server, *agentpodservice.PodService, *gorm.DB, *agentpod.Pod) {
	t.Helper()
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(
		"INSERT INTO runners (id, organization_id, node_id, status, current_pods) VALUES (11, 7, 'runner-11', 'online', 0)",
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO users (id, username, name, email) VALUES (42, 'owner-42', 'Owner', 'owner@example.com')",
	).Error)
	service := agentpodservice.NewPodService(infra.NewPodRepository(db))
	pod, err := service.CreatePod(context.Background(), &agentpodservice.CreatePodRequest{
		OrganizationID: 7,
		RunnerID:       11,
		CreatedByID:    42,
		AgentSlug:      "codex-cli",
		PreviewPort:    3000,
		PreviewPath:    "/app",
	})
	require.NoError(t, err)

	options := []Option{WithMobileBaseURL("https://mobile.example/")}
	if withRelay {
		manager := relay.NewManagerWithOptions()
		t.Cleanup(manager.Stop)
		require.NoError(t, manager.Register(&relay.RelayInfo{
			ID: "relay-1", URL: "wss://relay.example", Healthy: true,
		}))
		options = append(options, WithRelayManager(manager))
	}
	return NewServer(service, &fakeOrgService{role: "member"}, options...), service, db, pod
}
