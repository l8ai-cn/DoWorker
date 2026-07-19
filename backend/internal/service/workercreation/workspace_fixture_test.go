package workercreation

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	"github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
	"github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

type repositoryAccess struct {
	ID     int64
	OrgID  int64
	UserID int64
}

type workspaceRepositoryLookup struct {
	row   *gitprovider.Repository
	err   error
	calls int
	last  repositoryAccess
}

func (lookup *workspaceRepositoryLookup) GetAccessibleByID(
	_ context.Context,
	id, orgID, userID int64,
) (*gitprovider.Repository, error) {
	lookup.calls++
	lookup.last = repositoryAccess{ID: id, OrgID: orgID, UserID: userID}
	return lookup.row, lookup.err
}

type workspaceSkillLookup struct {
	rows map[int64]*skill.Skill
	ids  []int64
}

func (lookup *workspaceSkillLookup) GetAnyByID(
	_ context.Context,
	id int64,
) (*skill.Skill, error) {
	lookup.ids = append(lookup.ids, id)
	return lookup.rows[id], nil
}

type workspaceKnowledgeLookup struct {
	rows map[int64]*knowledgebase.KnowledgeBase
	ids  []int64
}

func (lookup *workspaceKnowledgeLookup) Get(
	_ context.Context,
	_ int64,
	id int64,
) (*knowledgebase.KnowledgeBase, error) {
	lookup.ids = append(lookup.ids, id)
	return lookup.rows[id], nil
}

type workspaceEnvBundleLookup struct {
	rows map[int64]*envbundle.EnvBundle
	ids  []int64
}

func (lookup *workspaceEnvBundleLookup) GetByID(
	_ context.Context,
	id int64,
) (*envbundle.EnvBundle, error) {
	lookup.ids = append(lookup.ids, id)
	return lookup.rows[id], nil
}

type workspaceCommitLookup struct{}

func (*workspaceCommitLookup) ResolveRepositoryCommit(
	context.Context,
	specservice.Scope,
	*gitprovider.Repository,
	string,
) (string, error) {
	return "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil
}

func (*workspaceCommitLookup) ResolveKnowledgeBaseCommit(
	context.Context,
	specservice.Scope,
	*knowledgebase.KnowledgeBase,
	string,
) (string, error) {
	return "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
}

type workspaceFixture struct {
	repositories *workspaceRepositoryLookup
	skills       *workspaceSkillLookup
	knowledge    *workspaceKnowledgeLookup
	envBundles   *workspaceEnvBundleLookup
	commits      *workspaceCommitLookup
}

func newWorkspaceFixture() *workspaceFixture {
	userID := int64(7)
	return &workspaceFixture{
		repositories: &workspaceRepositoryLookup{
			row: &gitprovider.Repository{
				ID:             22,
				OrganizationID: 77,
				Slug:           "org/repo",
				HttpCloneURL:   "https://example.com/org/repo.git",
				DefaultBranch:  "main",
				Visibility:     "organization",
				IsActive:       true,
			},
		},
		skills: &workspaceSkillLookup{
			rows: map[int64]*skill.Skill{
				3: {
					ID: 3, Slug: "code-review", IsActive: true, Version: 2,
					ContentSha:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					StorageKey:  "skills/code-review.tar.gz",
					PackageSize: 128,
				},
			},
		},
		knowledge: &workspaceKnowledgeLookup{
			rows: map[int64]*knowledgebase.KnowledgeBase{
				4: {
					ID: 4, OrganizationID: 77, Slug: "engineering-docs",
					HTTPCloneURL:  "https://example.com/kb/engineering-docs.git",
					DefaultBranch: "main",
				},
			},
		},
		envBundles: &workspaceEnvBundleLookup{
			rows: map[int64]*envbundle.EnvBundle{
				5: {
					ID: 5, OwnerScope: envbundle.OwnerScopeUser,
					OwnerID: userID, Name: "runtime-preferences",
					Kind: envbundle.KindRuntime, IsActive: true,
				},
				6: {
					ID: 6, OwnerScope: envbundle.OwnerScopeUser,
					OwnerID: userID, Name: "signing-secrets",
					Kind:     envbundle.KindCredential,
					Data:     envbundle.BundleData{"SIGNING_KEY": "encrypted-value"},
					IsActive: true,
				},
			},
		},
		commits: &workspaceCommitLookup{},
	}
}

func secretField(value string) string {
	if value == "" {
		return "SIGNING_KEY"
	}
	return value
}
