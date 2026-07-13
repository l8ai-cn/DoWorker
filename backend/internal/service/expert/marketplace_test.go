package expert

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListMarketApplicationsReturnsInstallableExpertTemplates(t *testing.T) {
	svc := newTestService(newFakeStore(), nil, &fakeDispatcher{})

	items := svc.ListMarketApplications()

	require.Len(t, items, 3)
	assert.Equal(t, "software-delivery-expert", items[0].Slug)
	assert.Equal(t, "codex-cli", items[0].AgentSlug)
	assert.Contains(t, items[0].SkillSlugs, "e2e")
	assert.NotEmpty(t, items[0].Outcomes)
	assert.JSONEq(t,
		`{"market_application_slug":"software-delivery-expert"}`,
		string(items[0].RuntimeSnapshot()),
	)
}

func TestInstallMarketApplicationCreatesExpertAndIsIdempotent(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(store, nil, &fakeDispatcher{})

	first, alreadyInstalled, err := svc.InstallMarketApplication(
		context.Background(), 7, 9, "software-delivery-expert",
	)
	require.NoError(t, err)
	assert.False(t, alreadyInstalled)
	assert.Equal(t, "软件交付专家", first.Name)
	assert.Equal(t, "codex-cli", first.AgentSlug)
	assert.Equal(t, []string{"worktree", "e2e", "gh-merge", "merge"}, []string(first.SkillSlugs))

	second, alreadyInstalled, err := svc.InstallMarketApplication(
		context.Background(), 7, 9, "software-delivery-expert",
	)
	require.NoError(t, err)
	assert.True(t, alreadyInstalled)
	assert.Equal(t, first.ID, second.ID)
}

func TestInstallMarketApplicationRejectsUnknownSlug(t *testing.T) {
	svc := newTestService(newFakeStore(), nil, &fakeDispatcher{})

	_, _, err := svc.InstallMarketApplication(context.Background(), 7, 9, "missing")

	assert.ErrorIs(t, err, ErrMarketApplicationNotFound)
	assert.False(t, errors.Is(err, context.Canceled))
}
