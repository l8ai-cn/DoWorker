package extension

import (
	"context"
	"encoding/json"
	"testing"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWorkerSkillsByIDsResolvesExactCatalogPackages(t *testing.T) {
	orgID := int64(77)
	catalog := &svcMockCatalog{
		getAnyByIDFn: func(_ context.Context, id int64) (*skilldom.Skill, error) {
			return &skilldom.Skill{
				ID:             id,
				OrganizationID: &orgID,
				Slug:           "reviewer",
				IsActive:       true,
				ContentSha:     "sha-reviewer",
				StorageKey:     "skills/reviewer.tar.gz",
				PackageSize:    123,
			}, nil
		},
	}
	service := newTestService(&svcMockRepo{}, &svcMockStorage{}, nil)
	service.SetSkillCatalog(catalog)

	skills, err := service.GetWorkerSkillsByIDs(
		context.Background(),
		orgID,
		[]int64{41},
		"codex-cli",
	)

	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, int64(41), skills[0].CatalogSkillID)
	assert.Equal(t, "reviewer", skills[0].Slug)
	assert.Equal(t, "sha-reviewer", skills[0].ContentSha)
	assert.Contains(t, skills[0].DownloadURL, "skills/reviewer.tar.gz")
}

func TestGetWorkerSkillsByIDsRejectsCrossOrganizationSkill(t *testing.T) {
	otherOrgID := int64(88)
	catalog := &svcMockCatalog{
		getAnyByIDFn: func(_ context.Context, id int64) (*skilldom.Skill, error) {
			return &skilldom.Skill{
				ID:             id,
				OrganizationID: &otherOrgID,
				Slug:           "private",
				IsActive:       true,
				ContentSha:     "sha-private",
				StorageKey:     "skills/private.tar.gz",
			}, nil
		},
	}
	service := newTestService(&svcMockRepo{}, &svcMockStorage{}, nil)
	service.SetSkillCatalog(catalog)

	_, err := service.GetWorkerSkillsByIDs(
		context.Background(),
		77,
		[]int64{41},
		"codex-cli",
	)

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestGetWorkerSkillsByIDsRejectsInvalidAgentFilter(t *testing.T) {
	catalog := &svcMockCatalog{
		getAnyByIDFn: func(_ context.Context, id int64) (*skilldom.Skill, error) {
			return &skilldom.Skill{
				ID:          id,
				Slug:        "reviewer",
				IsActive:    true,
				AgentFilter: json.RawMessage(`{"invalid":true}`),
				ContentSha:  "sha-reviewer",
				StorageKey:  "skills/reviewer.tar.gz",
			}, nil
		},
	}
	service := newTestService(&svcMockRepo{}, &svcMockStorage{}, nil)
	service.SetSkillCatalog(catalog)

	_, err := service.GetWorkerSkillsByIDs(
		context.Background(),
		77,
		[]int64{41},
		"codex-cli",
	)

	assert.ErrorIs(t, err, ErrInvalidInput)
}
