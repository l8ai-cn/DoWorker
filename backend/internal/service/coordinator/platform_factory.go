package coordinator

import (
	"context"
	"errors"
	"fmt"

	coordinatordom "github.com/l8ai-cn/agentcloud/backend/internal/domain/coordinator"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/git"
)

type defaultPlatformFactory struct {
	repos  RepoResolver
	tokens TokenProvider
}

func NewPlatformFactory(repos RepoResolver, tokens TokenProvider) PlatformFactory {
	return &defaultPlatformFactory{repos: repos, tokens: tokens}
}

func (f *defaultPlatformFactory) For(ctx context.Context, project *coordinatordom.Project) (TaskPlatform, string, error) {
	repo, err := f.repos.GetByID(ctx, project.RepositoryID)
	if err != nil {
		return nil, "", fmt.Errorf("load repository: %w", err)
	}
	if repo.ImportedByUserID == nil {
		return nil, "", errors.New("repository has no owning user for provider token lookup")
	}
	token, err := f.tokens.GetDecryptedProviderTokenByTypeAndURL(ctx, *repo.ImportedByUserID, repo.ProviderType, repo.ProviderBaseURL)
	if err != nil {
		return nil, "", fmt.Errorf("resolve provider token: %w", err)
	}

	switch project.PlatformType {
	case coordinatordom.PlatformTypeCNB:
		client, err := git.NewIssueClient(git.ProviderTypeCNB, repo.ProviderBaseURL, token)
		if err != nil {
			return nil, "", err
		}
		return NewCNBPlatform(client), repo.Slug, nil
	case coordinatordom.PlatformTypeLinear:
		return NewLinearPlatform(token, repo.ProviderBaseURL), repo.Slug, nil
	default:
		return nil, "", fmt.Errorf("unsupported coordinator platform %q", project.PlatformType)
	}
}
