package expert

import (
	"context"
	"testing"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateRejectsResourceManagedExpertBeforeGitOrStoreMutation(t *testing.T) {
	for _, managed := range managedExpertBindings() {
		t.Run(managed.name, func(t *testing.T) {
			store := newFakeStore()
			git := gitops.NewFake("am-experts")
			row := managed.expert()
			seedExpertRepo(t, git, row)
			require.NoError(t, store.Create(context.Background(), row))
			svc := newTestService(store, git, &fakeDispatcher{})

			_, err := svc.Update(context.Background(), &UpdateExpertRequest{
				OrganizationID: row.OrganizationID,
				ExpertID:       row.ID,
				Name:           strptr("Changed"),
			})

			require.EqualError(t, err, "expert definition changes must go through resource validate-plan-apply")
			stored, getErr := store.GetByID(context.Background(), row.OrganizationID, row.ID)
			require.NoError(t, getErr)
			assert.Equal(t, "Review", stored.Name)
			assert.NotContains(t, git.Repos["org7-review"].Files, "expert.json")
		})
	}
}

func TestDeleteRejectsResourceManagedExpertBeforeStoreOrGitMutation(t *testing.T) {
	for _, managed := range managedExpertBindings() {
		t.Run(managed.name, func(t *testing.T) {
			store := newFakeStore()
			git := gitops.NewFake("am-experts")
			row := managed.expert()
			seedExpertRepo(t, git, row)
			require.NoError(t, store.Create(context.Background(), row))
			svc := newTestService(store, git, &fakeDispatcher{})

			err := svc.Delete(context.Background(), row.OrganizationID, row.ID)

			require.EqualError(t, err, "expert definition changes must go through resource validate-plan-apply")
			_, getErr := store.GetByID(context.Background(), row.OrganizationID, row.ID)
			require.NoError(t, getErr)
			assert.Contains(t, git.Repos, "org7-review")
		})
	}
}

type managedExpertBinding struct {
	name   string
	expert func() *expertdom.Expert
}

func managedExpertBindings() []managedExpertBinding {
	snapshotID := int64(42)
	resourceID := int64(90)
	resourceRevision := int64(3)
	return []managedExpertBinding{
		{
			name: "worker spec snapshot",
			expert: func() *expertdom.Expert {
				row := managedExpert()
				row.WorkerSpecSnapshotID = &snapshotID
				return row
			},
		},
		{
			name: "orchestration resource",
			expert: func() *expertdom.Expert {
				row := managedExpert()
				row.OrchestrationResourceID = &resourceID
				return row
			},
		},
		{
			name: "orchestration resource revision",
			expert: func() *expertdom.Expert {
				row := managedExpert()
				row.OrchestrationResourceRevision = &resourceRevision
				return row
			},
		},
	}
}

func managedExpert() *expertdom.Expert {
	return &expertdom.Expert{
		OrganizationID: 7,
		Slug:           "review",
		Name:           "Review",
		AgentSlug:      "resource-native",
	}
}

func seedExpertRepo(t *testing.T, git *gitops.Fake, row *expertdom.Expert) {
	t.Helper()
	repo, err := git.Provision(context.Background(), gitops.ProvisionParams{
		OrgID: row.OrganizationID,
		Slug:  row.Slug,
	})
	require.NoError(t, err)
	row.GitRepoPath = &repo.Path
}
