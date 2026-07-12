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
		RuntimeSnapshot: []byte(`{
		  "name":"商品优化应用",
		  "agent_slug":"codex-cli",
		  "interaction_mode":"pty",
		  "automation_level":"autonomous",
		  "skill_slugs":["e2e"]
		}`),
	}

	first, existing, err := svc.InstallMarketplaceExpert(context.Background(), request)
	require.NoError(t, err)
	require.False(t, existing)
	require.Equal(t, int64(9), first.OrganizationID)
	require.Equal(t, "market-aaaaaaaaaaaa4aaa8aaaaaaaaaaaaaaa", first.Slug)
	require.Nil(t, first.RunnerID)
	require.Nil(t, first.RepositoryID)
	require.Nil(t, first.BranchName)
	require.Equal(t, []string{"e2e"}, []string(first.SkillSlugs))

	second, existing, err := svc.InstallMarketplaceExpert(context.Background(), request)
	require.NoError(t, err)
	require.True(t, existing)
	require.Equal(t, first.ID, second.ID)
	require.Len(t, store.rows, 1)
}
