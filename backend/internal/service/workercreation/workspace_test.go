package workercreation

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	"github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
	"github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	repositoryservice "github.com/anthropics/agentsmesh/backend/internal/service/repository"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceResolverValidatesScopedReferencesAndReturnsCompilerNames(t *testing.T) {
	fixture := newWorkspaceFixture()
	resolver := newWorkspaceResolver(fixture.deps())
	workspace := validWorkspaceDraft()
	workerType := slugkit.MustNewForTest("codex-cli")
	scope := specservice.Scope{OrgID: 77, UserID: 7}
	resolved, err := resolver.ResolveWorkspace(
		context.Background(),
		scope,
		workerType,
		workspace,
	)
	require.NoError(t, err)
	assert.Equal(t, workspace, resolved)
	assert.Equal(t, repositoryAccess{
		ID:     22,
		OrgID:  77,
		UserID: 7,
	}, fixture.repositories.last)
	assert.Equal(t, []int64{3}, fixture.skills.ids)
	assert.Equal(t, []int64{4}, fixture.knowledge.ids)
	assert.Equal(t, []int64{5}, fixture.envBundles.ids)

	err = resolver.ResolveSecretReference(
		context.Background(),
		scope,
		workerType,
		"SIGNING_KEY",
		specdomain.SecretReference{
			Kind: slugkit.MustNewForTest("env-bundle"),
			ID:   6,
		},
	)
	require.NoError(t, err)

	references, err := resolver.ResolveCompilationReferences(
		context.Background(),
		scope,
		workerType,
		workspace,
		map[string]specdomain.SecretReference{
			"SIGNING_KEY": {
				Kind: slugkit.MustNewForTest("env-bundle"),
				ID:   6,
			},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "org/repo", references.RepositorySlug)
	assert.Equal(t, []string{"code-review"}, references.SkillSlugs)
	assert.Equal(t, []knowledgeReference{{Slug: "docs", Mode: specdomain.KnowledgeMountReadWrite}}, references.Knowledge)
	assert.Equal(t, []string{"runtime-preferences", "signing-secrets"}, references.EnvBundleNames)
}

func TestWorkspaceResolverRejectsCrossScopeAndIncompatibleReferences(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*workspaceFixture)
		match  string
	}{
		{
			name: "repository from another organization",
			mutate: func(fixture *workspaceFixture) {
				fixture.repositories.err = repositoryservice.ErrNoPermission
			},
			match: "repository",
		},
		{
			name: "skill from another organization",
			mutate: func(fixture *workspaceFixture) {
				otherOrg := int64(78)
				fixture.skills.rows[3].OrganizationID = &otherOrg
			},
			match: "skill",
		},
		{
			name: "skill incompatible with worker type",
			mutate: func(fixture *workspaceFixture) {
				fixture.skills.rows[3].AgentFilter = json.RawMessage(`["claude-code"]`)
			},
			match: "worker type",
		},
		{
			name: "skill package is unavailable",
			mutate: func(fixture *workspaceFixture) {
				fixture.skills.rows[3].ContentSha = ""
			},
			match: "package",
		},
		{
			name: "skill slug is invalid",
			mutate: func(fixture *workspaceFixture) {
				fixture.skills.rows[3].Slug = "../escape"
			},
			match: "slug",
		},
		{
			name: "knowledge base from another organization",
			mutate: func(fixture *workspaceFixture) {
				fixture.knowledge.rows[4].OrganizationID = 78
			},
			match: "knowledge",
		},
		{
			name: "credential bundle cannot be a runtime bundle",
			mutate: func(fixture *workspaceFixture) {
				fixture.envBundles.rows[5].Kind = envbundle.KindCredential
			},
			match: "runtime environment bundle",
		},
		{
			name: "bundle scoped to another user",
			mutate: func(fixture *workspaceFixture) {
				fixture.envBundles.rows[5].OwnerID = 999
			},
			match: "environment bundle",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newWorkspaceFixture()
			test.mutate(fixture)

			_, err := newWorkspaceResolver(fixture.deps()).ResolveWorkspace(
				context.Background(),
				specservice.Scope{OrgID: 77, UserID: 7},
				slugkit.MustNewForTest("codex-cli"),
				validWorkspaceDraft(),
			)

			require.Error(t, err)
			assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
			assert.ErrorContains(t, err, test.match)
		})
	}
}

func TestWorkspaceResolverRejectsDuplicateEnvBundleNames(t *testing.T) {
	fixture := newWorkspaceFixture()
	fixture.envBundles.rows[6].Name = fixture.envBundles.rows[5].Name
	resolver := newWorkspaceResolver(fixture.deps())

	_, err := resolver.ResolveCompilationReferences(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
		validWorkspaceDraft(),
		map[string]specdomain.SecretReference{
			"SIGNING_KEY": {
				Kind: slugkit.MustNewForTest("env-bundle"),
				ID:   6,
			},
		},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
	assert.ErrorContains(t, err, "name")
}

func TestWorkspaceResolverRejectsDuplicateSkillSlugs(t *testing.T) {
	fixture := newWorkspaceFixture()
	fixture.skills.rows[9] = &skill.Skill{
		ID:         9,
		Slug:       fixture.skills.rows[3].Slug,
		IsActive:   true,
		ContentSha: "sha-duplicate",
		StorageKey: "skills/duplicate.tar.gz",
	}
	workspace := validWorkspaceDraft()
	workspace.SkillIDs = []int64{3, 9}

	_, err := newWorkspaceResolver(fixture.deps()).ResolveWorkspace(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		slugkit.MustNewForTest("codex-cli"),
		workspace,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
	assert.ErrorContains(t, err, "slug")
}

func TestWorkspaceResolverRejectsInvalidSecretReferences(t *testing.T) {
	scope := specservice.Scope{OrgID: 77, UserID: 7}
	workerType := slugkit.MustNewForTest("codex-cli")

	tests := []struct {
		name      string
		field     string
		reference specdomain.SecretReference
		mutate    func(*workspaceFixture)
		match     string
	}{
		{
			name: "unsupported reference kind",
			reference: specdomain.SecretReference{
				Kind: slugkit.MustNewForTest("vault-secret"),
				ID:   6,
			},
			match: "kind",
		},
		{
			name: "runtime bundle cannot satisfy a secret field",
			reference: specdomain.SecretReference{
				Kind: slugkit.MustNewForTest("env-bundle"),
				ID:   6,
			},
			mutate: func(fixture *workspaceFixture) {
				fixture.envBundles.rows[6].Kind = envbundle.KindRuntime
			},
			match: "credential",
		},
		{
			name: "credential bundle omits the declared field",
			reference: specdomain.SecretReference{
				Kind: slugkit.MustNewForTest("env-bundle"),
				ID:   6,
			},
			mutate: func(fixture *workspaceFixture) {
				fixture.envBundles.rows[6].Data = envbundle.BundleData{}
			},
			match: "does not configure",
		},
		{
			name:  "credential field is not declared",
			field: "UNDECLARED_KEY",
			reference: specdomain.SecretReference{
				Kind: slugkit.MustNewForTest("env-bundle"),
				ID:   6,
			},
			match: "not declared",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newWorkspaceFixture()
			resolver := newWorkspaceResolver(fixture.deps())
			if test.mutate != nil {
				test.mutate(fixture)
			}
			err := resolver.ResolveSecretReference(
				context.Background(),
				scope,
				workerType,
				secretField(test.field),
				test.reference,
			)
			require.Error(t, err)
			assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
			assert.ErrorContains(t, err, test.match)
		})
	}
}

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

type workspaceFixture struct {
	repositories *workspaceRepositoryLookup
	skills       *workspaceSkillLookup
	knowledge    *workspaceKnowledgeLookup
	envBundles   *workspaceEnvBundleLookup
}

func newWorkspaceFixture() *workspaceFixture {
	userID := int64(7)
	return &workspaceFixture{
		repositories: &workspaceRepositoryLookup{
			row: &gitprovider.Repository{
				ID:             22,
				OrganizationID: 77,
				Slug:           "org/repo",
				Visibility:     "organization",
				IsActive:       true,
			},
		},
		skills: &workspaceSkillLookup{
			rows: map[int64]*skill.Skill{
				3: {
					ID:         3,
					Slug:       "code-review",
					IsActive:   true,
					ContentSha: "sha-code-review",
					StorageKey: "skills/code-review.tar.gz",
				},
			},
		},
		knowledge: &workspaceKnowledgeLookup{
			rows: map[int64]*knowledgebase.KnowledgeBase{
				4: {ID: 4, OrganizationID: 77, Slug: "docs"},
			},
		},
		envBundles: &workspaceEnvBundleLookup{
			rows: map[int64]*envbundle.EnvBundle{
				5: {
					ID:         5,
					OwnerScope: envbundle.OwnerScopeUser,
					OwnerID:    userID,
					Name:       "runtime-preferences",
					Kind:       envbundle.KindRuntime,
					IsActive:   true,
				},
				6: {
					ID:         6,
					OwnerScope: envbundle.OwnerScopeUser,
					OwnerID:    userID,
					Name:       "signing-secrets",
					Kind:       envbundle.KindCredential,
					Data:       envbundle.BundleData{"SIGNING_KEY": "encrypted-value"},
					IsActive:   true,
				},
			},
		},
	}
}

func secretField(value string) string {
	if value == "" {
		return "SIGNING_KEY"
	}
	return value
}

func (fixture *workspaceFixture) deps() workspaceResolverDeps {
	return workspaceResolverDeps{
		Repositories: fixture.repositories,
		Skills:       fixture.skills,
		Knowledge:    fixture.knowledge,
		EnvBundles:   fixture.envBundles,
		Definitions: staticWorkerDefinitions{
			"codex-cli": workerDefinition(
				"codex-cli",
				"codex",
				"AGENT codex\nEXECUTABLE codex\n",
				"pty",
			),
		},
	}
}

func validWorkspaceDraft() specdomain.Workspace {
	repositoryID := int64(22)
	return specdomain.Workspace{
		RepositoryID: &repositoryID,
		Branch:       "main",
		SkillIDs:     []int64{3},
		KnowledgeMounts: []specdomain.KnowledgeMount{
			{KnowledgeBaseID: 4, Mode: specdomain.KnowledgeMountReadWrite},
		},
		EnvBundleIDs: []specdomain.RuntimeEnvBundleID{5},
		Instructions: "Review before editing.",
		InitialTask:  "Fix the failing test.",
	}
}
