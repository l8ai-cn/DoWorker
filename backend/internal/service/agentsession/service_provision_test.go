package agentsession_test

import (
	"context"
	"encoding/json"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	itemDomain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	svc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	"github.com/stretchr/testify/require"
)

func TestProvisionForPodCreatesSessionFromTrustedPod(t *testing.T) {
	service := svc.NewService(testDB(t))
	title := "Seedance storyboard"
	pod := &podDomain.Pod{
		OrganizationID: 3,
		CreatedByID:    7,
		PodKey:         "7-standalone-caa70224",
		AgentSlug:      "seedance-expert",
	}

	receipt, err := service.PrepareForPod(context.Background(), pod, sessionDomain.ProvisionSpec{
		ID:    "conv_seedance",
		Title: &title,
	})

	require.NoError(t, err)
	row := receipt.Session
	require.True(t, receipt.Created)
	require.Equal(t, "conv_seedance", row.ID)
	require.Equal(t, pod.PodKey, row.PodKey)
	require.Equal(t, pod.OrganizationID, row.OrganizationID)
	require.Equal(t, pod.CreatedByID, row.UserID)
	require.Equal(t, pod.AgentSlug, row.AgentSlug)
	require.Equal(t, title, *row.Title)

	persisted, err := service.GetByPodKey(context.Background(), pod.PodKey)
	require.NoError(t, err)
	require.Equal(t, row.ID, persisted.ID)
}

func TestProvisionForPodRebindsExistingSession(t *testing.T) {
	service := svc.NewService(testDB(t))
	require.NoError(t, service.Create(context.Background(), &sessionDomain.Session{
		ID:             "conv_seedance",
		OrganizationID: 3,
		UserID:         7,
		PodKey:         "7-standalone-old",
		AgentSlug:      "codex-cli",
		Status:         "idle",
	}))
	pod := &podDomain.Pod{
		OrganizationID: 3,
		CreatedByID:    7,
		PodKey:         "7-standalone-new",
		AgentSlug:      "seedance-expert",
	}

	receipt, err := service.PrepareForPod(context.Background(), pod, sessionDomain.ProvisionSpec{
		ID:             "conv_seedance",
		ExpectedPodKey: "7-standalone-old",
		UpdateExisting: true,
	})

	require.NoError(t, err)
	row := receipt.Session
	require.False(t, receipt.Created)
	require.Equal(t, "7-standalone-old", receipt.PreviousPodKey)
	require.Equal(t, "codex-cli", receipt.PreviousAgentSlug)
	require.Equal(t, pod.PodKey, row.PodKey)
	require.Equal(t, pod.AgentSlug, row.AgentSlug)
	persisted, err := service.Get(context.Background(), row.ID)
	require.NoError(t, err)
	require.Equal(t, pod.PodKey, persisted.PodKey)
	require.Equal(t, pod.AgentSlug, persisted.AgentSlug)
}

func TestPrepareForPodRejectsStaleExpectedPodKey(t *testing.T) {
	service := svc.NewService(testDB(t))
	require.NoError(t, service.Create(context.Background(), &sessionDomain.Session{
		ID:             "conv_seedance",
		OrganizationID: 3,
		UserID:         7,
		PodKey:         "7-standalone-old",
		AgentSlug:      "codex-cli",
		Status:         "idle",
	}))
	firstPod := &podDomain.Pod{
		OrganizationID: 3, CreatedByID: 7,
		PodKey: "7-standalone-first", AgentSlug: "seedance-expert",
	}
	secondPod := &podDomain.Pod{
		OrganizationID: 3, CreatedByID: 7,
		PodKey: "7-standalone-second", AgentSlug: "seedance-expert",
	}
	spec := sessionDomain.ProvisionSpec{
		ID: "conv_seedance", ExpectedPodKey: "7-standalone-old", UpdateExisting: true,
	}

	_, err := service.PrepareForPod(context.Background(), firstPod, spec)
	require.NoError(t, err)
	_, err = service.PrepareForPod(context.Background(), secondPod, spec)
	require.ErrorIs(t, err, svc.ErrSessionBindingChanged)

	persisted, err := service.Get(context.Background(), "conv_seedance")
	require.NoError(t, err)
	require.Equal(t, firstPod.PodKey, persisted.PodKey)
}

func TestRollbackProvisionDeletesCreatedSessionAndItems(t *testing.T) {
	db := testDB(t)
	service := svc.NewService(db)
	pod := &podDomain.Pod{
		OrganizationID: 3, CreatedByID: 7,
		PodKey: "7-standalone-new", AgentSlug: "seedance-expert",
	}
	receipt, err := service.PrepareForPod(context.Background(), pod, sessionDomain.ProvisionSpec{
		ID: "conv_seedance",
	})
	require.NoError(t, err)
	require.NoError(t, db.Create(&itemDomain.Item{
		ID: "item_seedance", SessionID: receipt.Session.ID, ItemType: "message",
		ResponseID: "resp_seedance", Status: "completed", Position: 1,
		Payload: json.RawMessage(`{"role":"user"}`),
	}).Error)

	require.NoError(t, service.RollbackProvision(context.Background(), receipt))

	_, err = service.Get(context.Background(), receipt.Session.ID)
	require.ErrorIs(t, err, svc.ErrNotFound)
	var count int64
	require.NoError(t, db.Model(&itemDomain.Item{}).
		Where("session_id = ?", receipt.Session.ID).Count(&count).Error)
	require.Zero(t, count)
}

func TestRollbackProvisionRestoresPreviousBinding(t *testing.T) {
	service := svc.NewService(testDB(t))
	require.NoError(t, service.Create(context.Background(), &sessionDomain.Session{
		ID:             "conv_seedance",
		OrganizationID: 3,
		UserID:         7,
		PodKey:         "7-standalone-old",
		AgentSlug:      "codex-cli",
		Status:         "idle",
	}))
	pod := &podDomain.Pod{
		OrganizationID: 3, CreatedByID: 7,
		PodKey: "7-standalone-new", AgentSlug: "seedance-expert",
	}
	receipt, err := service.PrepareForPod(context.Background(), pod, sessionDomain.ProvisionSpec{
		ID: "conv_seedance", ExpectedPodKey: "7-standalone-old", UpdateExisting: true,
	})
	require.NoError(t, err)

	require.NoError(t, service.RollbackProvision(context.Background(), receipt))

	persisted, err := service.Get(context.Background(), "conv_seedance")
	require.NoError(t, err)
	require.Equal(t, "7-standalone-old", persisted.PodKey)
	require.Equal(t, "codex-cli", persisted.AgentSlug)
}
