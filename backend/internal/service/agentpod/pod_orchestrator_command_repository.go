package agentpod

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
)

type podCommandRepositoryConfig struct {
	httpCloneURL       string
	sshCloneURL        string
	sourceBranch       string
	sourceCommitSha    string
	preparationScript  string
	preparationTimeout int
}

func podCommandRepository(
	req *OrchestrateCreatePodRequest,
	resolved *agentfileResolved,
	effectiveBranch *string,
) (podCommandRepositoryConfig, error) {
	config := podCommandRepositoryConfig{preparationTimeout: 300}
	if repo := resolved.Repository; repo != nil {
		config.httpCloneURL = repo.HttpCloneURL
		config.sshCloneURL = repo.SshCloneURL
		config.sourceBranch = repo.DefaultBranch
		if repo.PreparationScript != nil {
			config.preparationScript = *repo.PreparationScript
		}
		if repo.PreparationTimeout != nil {
			config.preparationTimeout = *repo.PreparationTimeout
		}
	}
	if effectiveBranch != nil && *effectiveBranch != "" {
		config.sourceBranch = *effectiveBranch
	}
	if req.preResolvedDependencies == nil ||
		req.preResolvedDependencies.Repository == nil {
		return config, nil
	}
	repository := req.preResolvedDependencies.Repository
	commit, err := artifactRepositoryCommit(repository)
	if err != nil {
		return podCommandRepositoryConfig{}, err
	}
	config.httpCloneURL = repository.HTTPCloneURL
	config.sshCloneURL = repository.SSHCloneURL
	config.sourceBranch = repository.Branch
	config.sourceCommitSha = commit
	config.preparationScript = repository.PreparationScript
	if repository.PreparationTimeoutSeconds > 0 {
		config.preparationTimeout = int(repository.PreparationTimeoutSeconds)
	}
	return config, nil
}

func (o *PodOrchestrator) podCommandGitCredential(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
) (artifactGitCredential, error) {
	if req.preResolvedDependencies != nil {
		if req.preResolvedDependencies.Repository == nil {
			return artifactGitCredential{
				credentialType: workerdependency.RepositoryCredentialTypeNone,
			}, nil
		}
		return o.artifactRepositoryCredential(ctx, req, req.preResolvedDependencies.Repository)
	}
	if o.userService == nil {
		return runtimeGitCredential(nil)
	}
	gitCred := o.getUserGitCredential(ctx, req.UserID)
	if gitCred == nil {
		return runtimeGitCredential(nil)
	}
	return runtimeGitCredential(gitCred)
}
