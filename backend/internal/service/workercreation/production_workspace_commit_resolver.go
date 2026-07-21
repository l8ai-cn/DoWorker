package workercreation

import (
	"context"
	"fmt"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/gitprovider"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/knowledgebase"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/git"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
)

type ProviderTokenLookup interface {
	GetDecryptedProviderTokenByTypeAndURL(
		context.Context,
		int64,
		string,
		string,
	) (string, error)
}

type GiteaCommitLookup interface {
	ResolvePublicBranchCommit(context.Context, string, string) (string, error)
	ResolveBranchCommit(context.Context, string, string) (string, error)
}

type providerBranchLookup interface {
	GetBranch(context.Context, string, string) (*git.Branch, error)
}

type providerBranchFactory func(string, string, string) (providerBranchLookup, error)

type productionWorkspaceCommitResolver struct {
	providerTokens        ProviderTokenLookup
	internalGitea         GiteaCommitLookup
	internalGiteaBaseURLs map[string]struct{}
	newProvider           providerBranchFactory
}

func NewProductionWorkspaceCommitResolver(
	providerTokens ProviderTokenLookup,
	internalGitea GiteaCommitLookup,
	giteaBaseURLs ...string,
) WorkspaceCommitResolver {
	return &productionWorkspaceCommitResolver{
		providerTokens:        providerTokens,
		internalGitea:         internalGitea,
		internalGiteaBaseURLs: normalizedProviderBaseURLs(giteaBaseURLs),
		newProvider: func(providerType, baseURL, token string) (providerBranchLookup, error) {
			return git.NewProvider(providerType, baseURL, token)
		},
	}
}

func (resolver *productionWorkspaceCommitResolver) ResolveRepositoryCommit(
	ctx context.Context,
	scope specservice.Scope,
	repository *gitprovider.Repository,
	branch string,
) (string, error) {
	if resolver == nil || repository == nil || strings.TrimSpace(branch) == "" {
		return "", fmt.Errorf("repository commit resolution requires repository and branch")
	}
	if resolver.isInternalGitea(repository.ProviderBaseURL) {
		if resolver.internalGitea == nil {
			return "", fmt.Errorf("internal Gitea commit resolver is unavailable")
		}
		return resolver.internalGitea.ResolvePublicBranchCommit(ctx, repository.Slug, branch)
	}
	if resolver.providerTokens == nil {
		return "", fmt.Errorf("repository provider token resolver is unavailable")
	}
	token, err := resolver.providerTokens.GetDecryptedProviderTokenByTypeAndURL(
		ctx,
		scope.UserID,
		repository.ProviderType,
		repository.ProviderBaseURL,
	)
	if err != nil {
		return "", fmt.Errorf("resolve repository provider token: %w", err)
	}
	if strings.TrimSpace(token) == "" {
		return "", fmt.Errorf("repository provider token is required")
	}
	provider, err := resolver.newProvider(repository.ProviderType, repository.ProviderBaseURL, token)
	if err != nil {
		return "", fmt.Errorf("create repository provider client: %w", err)
	}
	resolved, err := provider.GetBranch(ctx, repository.ExternalID, branch)
	if err != nil {
		return "", fmt.Errorf("resolve repository branch %q: %w", branch, err)
	}
	if resolved == nil {
		return "", fmt.Errorf("repository branch %q was not found", branch)
	}
	return resolved.CommitSHA, nil
}

func (resolver *productionWorkspaceCommitResolver) ResolveKnowledgeBaseCommit(
	ctx context.Context,
	_ specservice.Scope,
	knowledgeBase *knowledgebase.KnowledgeBase,
	branch string,
) (string, error) {
	if resolver == nil || resolver.internalGitea == nil {
		return "", fmt.Errorf("knowledge base commit resolver is unavailable")
	}
	if knowledgeBase == nil {
		return "", fmt.Errorf("knowledge base commit resolution requires a knowledge base")
	}
	return resolver.internalGitea.ResolveBranchCommit(ctx, knowledgeBase.GitRepoPath, branch)
}

func (resolver *productionWorkspaceCommitResolver) isInternalGitea(baseURL string) bool {
	_, found := resolver.internalGiteaBaseURLs[normalizeProviderBaseURL(baseURL)]
	return resolver.internalGitea != nil && found
}

func normalizeProviderBaseURL(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}

func normalizedProviderBaseURLs(values []string) map[string]struct{} {
	normalized := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value = normalizeProviderBaseURL(value); value != "" {
			normalized[value] = struct{}{}
		}
	}
	return normalized
}
