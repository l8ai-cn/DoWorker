package runner

import (
	"context"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func TestTerminateQueuedPod_UsesCompletedStatus(t *testing.T) {
	logger := newTestLogger()
	db, cm, tr, hb, podStore, runnerRepo := setupPodCoordinatorDeps(t)
	orgID := int64(1)
	r := createOnlineRunner(t, db, runnerRepo, 0, 5)
	require.NoError(t, seedQueuedPod(t, db, orgID, r.ID, "q-cancel"))

	pc := NewPodCoordinator(podStore, runnerRepo, cm, tr, hb, logger)
	pc.SetPendingQueue(NewPendingCommandQueue(&memPendingRepo{}, nil, 5, time.Minute, true, logger))

	err := pc.TerminatePod(context.Background(), "q-cancel")
	require.NoError(t, err)

	pod, err := podStore.GetByKey(context.Background(), "q-cancel")
	require.NoError(t, err)
	assert.Equal(t, agentpod.StatusCompleted, pod.Status)
}

func TestTerminatePod_CancelsPendingSendPrompt(t *testing.T) {
	logger := newTestLogger()
	db, cm, tr, hb, podStore, runnerRepo := setupPodCoordinatorDeps(t)
	orgID := int64(1)
	r := createOnlineRunner(t, db, runnerRepo, 1, 5)
	podKey := "run-cancel-pending"
	require.NoError(t, db.Exec(
		`INSERT INTO pods (organization_id, pod_key, runner_id, created_by_id, status) VALUES (?, ?, ?, 1, ?)`,
		orgID, podKey, r.ID, agentpod.StatusRunning,
	).Error)

	repo := &memPendingRepo{}
	queue := NewPendingCommandQueue(repo, nil, 5, time.Minute, true, logger)
	payload, err := proto.Marshal(&runnerv1.ServerMessage{
		Payload: &runnerv1.ServerMessage_SendPrompt{
			SendPrompt: &runnerv1.SendPromptCommand{PodKey: podKey, Prompt: "hello"},
		},
	})
	require.NoError(t, err)
	require.NoError(t, repo.Enqueue(context.Background(), &agentpod.PendingCommand{
		OrganizationID: orgID,
		RunnerID:       r.ID,
		PodKey:         podKey,
		CommandType:    agentpod.CommandTypeSendPrompt,
		CommandID:      "chmsg-1",
		Payload:        payload,
		ExpiresAt:      time.Now().Add(time.Minute),
	}))

	pc := NewPodCoordinator(podStore, runnerRepo, cm, tr, hb, logger)
	pc.SetPendingQueue(queue)
	pc.SetCommandSender(&MockCommandSender{})

	err = pc.TerminatePod(context.Background(), podKey)
	if err != nil {
		t.Logf("TerminatePod err (SQLite GREATEST in DecrementPods): %v", err)
	}
	assert.Empty(t, repo.rows)
}
