package workercreation

import (
	"context"
	"errors"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	"github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
	"github.com/anthropics/agentsmesh/backend/internal/infra/git"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type providerTokenFixture struct {
	token string
	err   error
}

func (fixture providerTokenFixture) GetDecryptedProviderTokenByTypeAndURL(
	context.Context,
	int64,
	string,
	string,
) (string, error) {
	return fixture.token, fixture.err
}

type giteaCommitFixture struct {
	publicCommit string
	commit       string
}

func (fixture giteaCommitFixture) ResolvePublicBranchCommit(
	context.Context,
	string,
	string,
) (string, error) {
	return fixture.publicCommit, nil
}

func (fixture giteaCommitFixture) ResolveBranchCommit(
	context.Context,
	string,
	string,
) (string, error) {
	return fixture.commit, nil
}

type branchFixture struct {
	commit string
}

func (fixture branchFixture) GetBranch(context.Context, string, string) (*git.Branch, error) {
	return &git.Branch{CommitSHA: fixture.commit}, nil
}

func TestProductionWorkspaceCommitResolverUsesConfiguredInternalGitea(t *testing.T) {
	resolver := NewProductionWorkspaceCommitResolver(
		nil,
		giteaCommitFixture{publicCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		"http://gitea:3000/",
	)

	commit, err := resolver.ResolveRepositoryCommit(
		context.Background(),
		specservice.Scope{OrgID: 7, UserID: 9},
		&gitprovider.Repository{ProviderBaseURL: "http://gitea:3000", Slug: "dev-org/demo"},
		"main",
	)

	require.NoError(t, err)
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", commit)
}

func TestProductionWorkspaceCommitResolverUsesTrustedInternalGiteaAlias(t *testing.T) {
	resolver := NewProductionWorkspaceCommitResolver(
		nil,
		giteaCommitFixture{publicCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		"http://localhost:12409",
		"http://gitea:3000",
	)

	commit, err := resolver.ResolveRepositoryCommit(
		context.Background(),
		specservice.Scope{OrgID: 7, UserID: 9},
		&gitprovider.Repository{ProviderBaseURL: "http://gitea:3000", Slug: "dev-org/demo"},
		"main",
	)

	require.NoError(t, err)
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", commit)
}

func TestProductionWorkspaceCommitResolverRequiresExternalProviderToken(t *testing.T) {
	resolver := NewProductionWorkspaceCommitResolver(
		providerTokenFixture{err: errors.New("missing token")},
		nil,
		"",
	)

	_, err := resolver.ResolveRepositoryCommit(
		context.Background(),
		specservice.Scope{OrgID: 7, UserID: 9},
		&gitprovider.Repository{ProviderType: "github", ProviderBaseURL: "https://github.com"},
		"main",
	)

	require.ErrorContains(t, err, "resolve repository provider token")
}

func TestProductionWorkspaceCommitResolverPinsExternalBranch(t *testing.T) {
	resolver := NewProductionWorkspaceCommitResolver(
		providerTokenFixture{token: "actor-token"},
		nil,
		"",
	).(*productionWorkspaceCommitResolver)
	resolver.newProvider = func(string, string, string) (providerBranchLookup, error) {
		return branchFixture{commit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}, nil
	}

	commit, err := resolver.ResolveRepositoryCommit(
		context.Background(),
		specservice.Scope{OrgID: 7, UserID: 9},
		&gitprovider.Repository{
			ProviderType: "github", ProviderBaseURL: "https://github.com", ExternalID: "org/repo",
		},
		"main",
	)

	require.NoError(t, err)
	assert.Equal(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", commit)
}

func TestProductionWorkspaceCommitResolverPinsKnowledgeBase(t *testing.T) {
	resolver := NewProductionWorkspaceCommitResolver(
		nil,
		giteaCommitFixture{commit: "cccccccccccccccccccccccccccccccccccccccc"},
		"http://gitea:3000",
	)

	commit, err := resolver.ResolveKnowledgeBaseCommit(
		context.Background(),
		specservice.Scope{OrgID: 7, UserID: 9},
		&knowledgebase.KnowledgeBase{GitRepoPath: "am-kb/handbook"},
		"main",
	)

	require.NoError(t, err)
	assert.Equal(t, "cccccccccccccccccccccccccccccccccccccccc", commit)
}
