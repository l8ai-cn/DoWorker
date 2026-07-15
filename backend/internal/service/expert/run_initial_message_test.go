package expert

import (
	"context"
	"testing"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunPromptOverrideStillValidatesSnapshot(t *testing.T) {
	snapshotID := int64(42)
	store := newFakeStore()
	require.NoError(t, store.Create(context.Background(), &expertdom.Expert{
		OrganizationID:       7,
		Slug:                 "review",
		Name:                 "Review",
		WorkerSpecSnapshotID: &snapshotID,
	}))
	dispatcher := &fakeDispatcher{}
	service := NewService(Deps{
		Store:       store,
		Dispatch:    dispatcher,
		WorkerSpecs: &expertSnapshotLoader{err: specdomain.ErrNotFound},
	})
	prompt := "Review the security boundary."

	_, err := service.Run(context.Background(), &RunExpertRequest{
		OrganizationID: 7,
		UserID:         5,
		ExpertSlug:     "review",
		PromptOverride: &prompt,
	})

	require.ErrorIs(t, err, ErrWorkerSpecSnapshotMismatch)
	assert.Nil(t, dispatcher.lastReq)
}

func TestResolveRunInitialMessageFallsBackForBlankOverride(t *testing.T) {
	snapshotID := int64(42)
	service := NewService(Deps{
		WorkerSpecs: &expertSnapshotLoader{
			snapshot: expertWorkerSpecSnapshot(snapshotID, 7),
		},
	})
	override := "  "

	text, err := service.resolveRunInitialMessage(
		context.Background(),
		7,
		snapshotID,
		&override,
	)

	require.NoError(t, err)
	assert.Equal(t, "Run checks.", text)
}
