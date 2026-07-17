package sessionapi

import (
	"context"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	itemDomain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEnsureTurnAutoStarts(t *testing.T) {
	hub := NewSessionHub()
	p := &SessionStreamPublisher{Hub: hub}
	id := p.ensureTurn("conv_test", "")
	require.NotEmpty(t, id)
	got, ok := hub.ActiveResponse("conv_test")
	require.True(t, ok)
	require.Equal(t, id, got)
}

func TestMapPodSessionStatus(t *testing.T) {
	require.Equal(t, "launching", mapPodSessionStatus(podDomain.StatusInitializing, ""))
	require.Equal(t, "launching", mapPodSessionStatus(podDomain.StatusQueued, ""))
	require.Equal(t, "idle", mapPodSessionStatus(podDomain.StatusRunning, podDomain.AgentStatusIdle))
	require.Equal(t, "running", mapPodSessionStatus(podDomain.StatusRunning, podDomain.AgentStatusExecuting))
	require.Equal(t, "idle", mapPodSessionStatus(podDomain.StatusCompleted, ""))
	require.Equal(t, "failed", mapPodSessionStatus(podDomain.StatusOrphaned, ""))
	require.Equal(t, "idle", mapPodSessionStatus(podDomain.StatusDisconnected, podDomain.AgentStatusIdle))
}

func TestMapSessionStatusDelegates(t *testing.T) {
	require.Equal(t, "idle", mapSessionStatus(nil))
	require.Equal(t, "launching", mapSessionStatus(&podDomain.Pod{Status: podDomain.StatusQueued}))
}

func TestHandleAcpSessionIdleFinalizesBuffer(t *testing.T) {
	hub := NewSessionHub()
	p := &SessionStreamPublisher{Hub: hub, Sessions: nil}
	p.HandleAcpSession(context.Background(), "missing-pod", "sessionState", `{"state":"idle"}`)
}

func TestHandleAcpSessionCreatesMissingSessionBeforePersistingResult(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/worker-session.db"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&sessionDomain.Session{}, &itemDomain.Item{}))
	pod := &podDomain.Pod{
		PodKey: "7-standalone-caa70224", OrganizationID: 3, CreatedByID: 7,
		AgentSlug: "seedance-expert", InteractionMode: podDomain.InteractionModeACP,
	}
	publisher := NewSessionStreamPublisher(
		NewSessionHub(),
		itemsvc.NewService(db),
		sessionsvc.NewService(db),
		nil,
	)
	publisher.Pods = streamPodStore{pod: pod}

	publisher.HandleAcpSession(
		context.Background(),
		pod.PodKey,
		"contentChunk",
		`{"role":"assistant","text":"credentials verified"}`,
	)
	publisher.HandleAcpSession(
		context.Background(),
		pod.PodKey,
		"sessionState",
		`{"state":"idle"}`,
	)

	session, err := publisher.Sessions.GetByPodKey(context.Background(), pod.PodKey)
	require.NoError(t, err)
	require.Equal(t, "seedance-expert", session.AgentSlug)
	page, err := publisher.Items.ListPage(context.Background(), session.ID, 20, "", false)
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Contains(t, string(page.Items[0].Payload), "credentials verified")
}

type streamPodStore struct {
	pod *podDomain.Pod
}

func (store streamPodStore) GetByKey(
	_ context.Context,
	_ string,
) (*podDomain.Pod, error) {
	return store.pod, nil
}

func (streamPodStore) UpdateExternalSessionID(
	context.Context,
	string,
	string,
) error {
	return nil
}
