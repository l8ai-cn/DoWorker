package expert

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstallMarketplaceExpertClonesTemplateIdempotently(t *testing.T) {
	store := newFakeStore()
	store.nextID = 200
	svc := NewService(Deps{Store: store})
	request := MarketplaceInstallationRequest{
		InstallationID:       "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		TargetOrganizationID: 9, ActorUserID: 14,
		RuntimeSnapshot: []byte(`{"market_application_slug":"software-delivery-expert"}`),
	}

	first, existing, err := svc.InstallMarketplaceExpert(context.Background(), request)
	require.NoError(t, err)
	require.False(t, existing)
	require.Equal(t, int64(9), first.OrganizationID)
	require.Equal(t, "market-aaaaaaaaaaaa4aaa8aaaaaaaaaaaaaaa", first.Slug)
	require.Equal(t, "软件交付专家", first.Name)
	require.Equal(t, "codex-cli", first.AgentSlug)
	require.Nil(t, first.RunnerID)
	require.Nil(t, first.RepositoryID)
	require.Nil(t, first.BranchName)
	require.Equal(t, []string{"worktree", "e2e", "gh-merge", "merge"}, []string(first.SkillSlugs))

	second, existing, err := svc.InstallMarketplaceExpert(context.Background(), request)
	require.NoError(t, err)
	require.True(t, existing)
	require.Equal(t, first.ID, second.ID)
	require.Len(t, store.rows, 1)
}

func TestInstallMarketplaceExpertRejectsUnknownTemplate(t *testing.T) {
	svc := NewService(Deps{Store: newFakeStore()})

	_, _, err := svc.InstallMarketplaceExpert(context.Background(), MarketplaceInstallationRequest{
		InstallationID:       "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		TargetOrganizationID: 9,
		ActorUserID:          14,
		RuntimeSnapshot:      []byte(`{"market_application_slug":"missing"}`),
	})

	require.ErrorIs(t, err, ErrMarketApplicationNotFound)
}
