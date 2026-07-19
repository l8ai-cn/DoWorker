package expert

import (
	"context"
	"testing"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunResourceManagedExpertDoesNotDispatchDefinitionPromptByDefault(t *testing.T) {
	snapshotID := int64(42)
	resourceID := int64(90)
	resourceRevision := int64(3)
	prompt := "Review carefully"
	store := newFakeStore()
	require.NoError(t, store.Create(context.Background(), &expertdom.Expert{
		OrganizationID:                7,
		Slug:                          "review",
		Name:                          "Review",
		AgentSlug:                     "resource-native",
		Prompt:                        &prompt,
		WorkerSpecSnapshotID:          &snapshotID,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
	}))
	dispatcher := &fakeDispatcher{}
	service := NewService(Deps{
		Store: store, Dispatch: dispatcher,
		WorkerSpecs: &expertSnapshotLoader{
			snapshot: expertWorkerSpecSnapshot(snapshotID, 7),
		},
	})

	_, err := service.Run(context.Background(), &RunExpertRequest{
		OrganizationID: 7,
		UserID:         5,
		ExpertSlug:     "review",
	})

	require.NoError(t, err)
	require.NotNil(t, dispatcher.lastReq)
	assert.Nil(t, dispatcher.lastReq.WorkerSpecPromptOverride)
}

func TestRunResourceManagedExpertPrefersRequestPromptOverride(t *testing.T) {
	snapshotID := int64(42)
	resourceID := int64(90)
	resourceRevision := int64(3)
	prompt := "Review carefully"
	override := "Only check authorization."
	store := newFakeStore()
	require.NoError(t, store.Create(context.Background(), &expertdom.Expert{
		OrganizationID:                7,
		Slug:                          "review",
		Name:                          "Review",
		AgentSlug:                     "resource-native",
		Prompt:                        &prompt,
		WorkerSpecSnapshotID:          &snapshotID,
		OrchestrationResourceID:       &resourceID,
		OrchestrationResourceRevision: &resourceRevision,
	}))
	dispatcher := &fakeDispatcher{}
	service := NewService(Deps{
		Store: store, Dispatch: dispatcher,
		WorkerSpecs: &expertSnapshotLoader{
			snapshot: expertWorkerSpecSnapshot(snapshotID, 7),
		},
	})

	_, err := service.Run(context.Background(), &RunExpertRequest{
		OrganizationID: 7,
		UserID:         5,
		ExpertSlug:     "review",
		PromptOverride: &override,
	})

	require.NoError(t, err)
	require.NotNil(t, dispatcher.lastReq.WorkerSpecPromptOverride)
	assert.Equal(t, override, *dispatcher.lastReq.WorkerSpecPromptOverride)
}

func TestRunRejectsIncompleteResourceManagedExpertBinding(t *testing.T) {
	snapshotID := int64(42)
	resourceID := int64(90)
	store := newFakeStore()
	require.NoError(t, store.Create(context.Background(), &expertdom.Expert{
		OrganizationID:          7,
		Slug:                    "review",
		Name:                    "Review",
		AgentSlug:               "resource-native",
		WorkerSpecSnapshotID:    &snapshotID,
		OrchestrationResourceID: &resourceID,
	}))
	dispatcher := &fakeDispatcher{}
	service := NewService(Deps{Store: store, Dispatch: dispatcher})

	_, err := service.Run(context.Background(), &RunExpertRequest{
		OrganizationID: 7,
		UserID:         5,
		ExpertSlug:     "review",
	})

	require.ErrorIs(t, err, ErrExpertResourceBindingCorrupt)
	assert.Nil(t, dispatcher.lastReq)
}
