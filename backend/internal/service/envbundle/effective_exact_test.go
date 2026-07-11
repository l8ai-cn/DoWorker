package envbundle

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEffectiveByIDsReturnsOnlyExactVisibleBundles(t *testing.T) {
	service, _ := newTestService(t)
	ctx := context.Background()
	selected, err := service.Create(ctx, &CreateParams{
		OwnerScope: envbundle.OwnerScopeUser,
		OwnerID:    7,
		AgentSlug:  strPtr("codex-cli"),
		Name:       "selected",
		Kind:       envbundle.KindCredential,
		Data:       map[string]string{"TOKEN": "secret"},
	})
	require.NoError(t, err)
	_, err = service.Create(ctx, &CreateParams{
		OwnerScope: envbundle.OwnerScopeUser,
		OwnerID:    7,
		AgentSlug:  strPtr("codex-cli"),
		Name:       "unselected",
		Kind:       envbundle.KindCredential,
		Data:       map[string]string{"TOKEN": "other"},
	})
	require.NoError(t, err)

	bundles, err := service.GetEffectiveByIDs(
		ctx,
		7,
		77,
		"codex-cli",
		[]int64{selected.ID},
	)

	require.NoError(t, err)
	require.Len(t, bundles, 1)
	assert.Equal(t, selected.ID, bundles[0].ID)
	assert.Equal(t, "secret", bundles[0].Data["TOKEN"])
}

func TestGetEffectiveByIDsRejectsUnavailableBundle(t *testing.T) {
	service, db := newTestService(t)
	ctx := context.Background()
	require.NoError(t, db.Exec(
		`INSERT INTO env_bundles
			(owner_scope, owner_id, agent_slug, name, kind, kind_primary, data, is_active)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		envbundle.OwnerScopeUser,
		7,
		"codex-cli",
		"corrupt",
		envbundle.KindCredential,
		false,
		`{"TOKEN":"not-ciphertext"}`,
		true,
	).Error)
	var row envbundle.EnvBundle
	require.NoError(t, db.Where("name = ?", "corrupt").First(&row).Error)

	_, err := service.GetEffectiveByIDs(
		ctx,
		7,
		77,
		"codex-cli",
		[]int64{row.ID},
	)

	require.Error(t, err)
	assert.ErrorContains(t, err, "decrypt")
}
