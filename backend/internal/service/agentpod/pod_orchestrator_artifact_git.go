package agentpod

import (
	"context"
	"errors"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/user"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	userService "github.com/l8ai-cn/agentcloud/backend/internal/service/user"
)

type artifactGitCredential struct {
	credentialType string
	token          string
	sshPrivateKey  string
}

func (o *PodOrchestrator) artifactRepositoryCredential(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	repository *workerdependency.Repository,
) (artifactGitCredential, error) {
	if repository == nil || repository.Credential.Type == workerdependency.RepositoryCredentialTypeNone {
		return artifactGitCredential{credentialType: workerdependency.RepositoryCredentialTypeNone}, nil
	}
	if repository.Credential.Type == user.CredentialTypeRunnerLocal {
		return runtimeGitCredential(nil)
	}
	if o.userService == nil ||
		repository.Credential.CredentialID == nil ||
		repository.Credential.OwnerUserID != req.UserID {
		return artifactGitCredential{}, ErrWorkerSpecDependencyUnavailable
	}
	credential, err := o.userService.GetDecryptedCredentialToken(
		ctx,
		req.UserID,
		*repository.Credential.CredentialID,
	)
	if err != nil {
		return artifactGitCredential{}, errors.Join(
			ErrWorkerSpecDependencyUnavailable,
			err,
		)
	}
	if credential.Type != repository.Credential.Type {
		return artifactGitCredential{}, ErrWorkerSpecDependencyUnavailable
	}
	return runtimeGitCredential(credential)
}

func runtimeGitCredential(
	credential *userService.DecryptedCredential,
) (artifactGitCredential, error) {
	if credential == nil {
		return artifactGitCredential{
			credentialType: user.CredentialTypeRunnerLocal,
		}, nil
	}
	switch credential.Type {
	case user.CredentialTypeRunnerLocal:
		return artifactGitCredential{
			credentialType: user.CredentialTypeRunnerLocal,
		}, nil
	case user.CredentialTypeOAuth, user.CredentialTypePAT:
		if strings.TrimSpace(credential.Token) == "" {
			return artifactGitCredential{}, ErrWorkerSpecDependencyUnavailable
		}
		return artifactGitCredential{
			credentialType: credential.Type,
			token:          credential.Token,
		}, nil
	case user.CredentialTypeSSHKey:
		if strings.TrimSpace(credential.SSHPrivateKey) == "" {
			return artifactGitCredential{}, ErrWorkerSpecDependencyUnavailable
		}
		return artifactGitCredential{
			credentialType: credential.Type,
			sshPrivateKey:  credential.SSHPrivateKey,
		}, nil
	default:
		return artifactGitCredential{}, ErrWorkerSpecDependencyUnavailable
	}
}

func artifactRepositoryCommit(
	repository *workerdependency.Repository,
) (string, error) {
	if repository == nil {
		return "", nil
	}
	commit := strings.TrimSpace(repository.CommitSHA)
	if commit == "" {
		return "", ErrWorkerSpecDependencyUnavailable
	}
	return commit, nil
}
