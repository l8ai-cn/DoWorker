package workercreation

import (
	"context"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/gitprovider"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdependencyartifact"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func buildRepositoryResolution(
	ctx context.Context,
	scope control.Scope,
	refs ArtifactReferences,
	spec workerspec.Spec,
	workspace *workspaceResolver,
) (*workerdependencyartifact.RepositoryResolution, error) {
	if spec.Workspace.RepositoryID == nil {
		return nil, nil
	}
	id := *spec.Workspace.RepositoryID
	reference := refs.Repository
	if reference == nil {
		return nil, fmt.Errorf("WorkerTemplate artifact repository reference is missing")
	}
	pin, err := referencePin(scope, *reference, id)
	if err != nil {
		return nil, err
	}
	repository := workspace.resolvedRepository(spec.Workspace.RepositoryID)
	if repository == nil {
		return nil, fmt.Errorf("WorkerTemplate artifact repository %d was not resolved", id)
	}
	if repository.DefaultBranch == "" || spec.Workspace.Branch == "" {
		return nil, fmt.Errorf("WorkerTemplate artifact repository branch is missing")
	}
	if workspace.deps.Commits == nil {
		return nil, fmt.Errorf("WorkerTemplate artifact repository commit resolver is unavailable")
	}
	commit, err := workspace.deps.Commits.ResolveRepositoryCommit(
		ctx,
		specScope(scope),
		repository,
		spec.Workspace.Branch,
	)
	if err != nil {
		return nil, err
	}
	commit, err = validateResolvedCommit("repository", commit)
	if err != nil {
		return nil, err
	}
	credentialType := workerdependency.RepositoryCredentialTypeNone
	script, scriptDigest, timeout, err := repositoryPreparation(repository)
	if err != nil {
		return nil, err
	}
	return &workerdependencyartifact.RepositoryResolution{
		ResourceResolution: pin, HTTPCloneURL: repository.HttpCloneURL,
		SSHCloneURL: repository.SshCloneURL, Branch: spec.Workspace.Branch,
		CommitSHA: commit, CredentialType: credentialType,
		PreparationScript:         script,
		PreparationScriptDigest:   scriptDigest,
		PreparationTimeoutSeconds: timeout,
	}, nil
}

func buildSkillResolutions(
	scope control.Scope,
	refs ArtifactReferences,
	spec workerspec.Spec,
	workspace *workspaceResolver,
) ([]workerdependencyartifact.SkillResolution, error) {
	result := make([]workerdependencyartifact.SkillResolution, 0, len(spec.Workspace.SkillIDs))
	packages := skillPackageIndex(spec.Workspace.SkillPackages)
	for _, id := range spec.Workspace.SkillIDs {
		packageBinding, err := requiredSkillPackage(packages, id)
		if err != nil {
			return nil, err
		}
		reference := refs.Skills[id]
		pin, err := referencePin(scope, reference, id)
		if err != nil {
			return nil, err
		}
		slug, err := slugkit.NewFromTrusted(packageBinding.Slug)
		if err != nil {
			return nil, err
		}
		result = append(result, workerdependencyartifact.SkillResolution{
			ResourceResolution: pin, Slug: slug,
			Version:       packageBinding.Version,
			ContentDigest: digestFromSHA(packageBinding.ContentSHA),
			StorageKey:    packageBinding.StorageKey,
			PackageSize:   packageBinding.PackageSize,
		})
	}
	return result, nil
}

func buildKnowledgeResolutions(
	ctx context.Context,
	scope control.Scope,
	refs ArtifactReferences,
	spec workerspec.Spec,
	workspace *workspaceResolver,
) ([]workerdependencyartifact.KnowledgeBaseResolution, error) {
	result := make([]workerdependencyartifact.KnowledgeBaseResolution, 0, len(spec.Workspace.KnowledgeMounts))
	for _, mount := range spec.Workspace.KnowledgeMounts {
		id := mount.KnowledgeBaseID
		row := workspace.knowledge[id]
		reference := refs.KnowledgeBases[id]
		if row == nil {
			return nil, fmt.Errorf("WorkerTemplate artifact knowledge base %d was not resolved", id)
		}
		if row.DefaultBranch == "" {
			return nil, fmt.Errorf("WorkerTemplate artifact knowledge base branch is missing")
		}
		if _, err := referencePin(scope, reference, id); err != nil {
			return nil, err
		}
		if workspace.deps.Commits == nil {
			return nil, fmt.Errorf(
				"WorkerTemplate artifact knowledge base commit resolver is unavailable",
			)
		}
		commit, err := workspace.deps.Commits.ResolveKnowledgeBaseCommit(
			ctx,
			specScope(scope),
			row,
			row.DefaultBranch,
		)
		if err != nil {
			return nil, err
		}
		commit, err = validateResolvedCommit("knowledge base", commit)
		if err != nil {
			return nil, err
		}
		slug, err := slugkit.NewFromTrusted(row.Slug)
		if err != nil {
			return nil, err
		}
		pin, err := referencePin(scope, reference, id)
		if err != nil {
			return nil, err
		}
		result = append(result, workerdependencyartifact.KnowledgeBaseResolution{
			ResourceResolution: pin, Slug: slug, HTTPCloneURL: row.HTTPCloneURL,
			Branch: row.DefaultBranch, CommitSHA: commit, Mode: mount.Mode,
		})
	}
	return result, nil
}

func repositoryPreparation(repository *gitprovider.Repository) (string, string, uint32, error) {
	if repository.PreparationScript == nil || *repository.PreparationScript == "" {
		return "", "", 0, nil
	}
	if repository.PreparationTimeout == nil || *repository.PreparationTimeout <= 0 {
		return "", "", 0, fmt.Errorf(
			"WorkerTemplate artifact repository preparation timeout is missing",
		)
	}
	script := *repository.PreparationScript
	return script, workerdependency.TextDigest(script), uint32(*repository.PreparationTimeout), nil
}

func specScope(scope control.Scope) specservice.Scope {
	return specservice.Scope{OrgID: scope.OrganizationID, UserID: scope.ActorID}
}
