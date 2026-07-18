package workercreation

import (
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdependencyartifact"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func buildRepositoryResolution(
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
	if _, err := referencePin(scope, *reference, id); err != nil {
		return nil, err
	}
	repository := workspace.resolvedRepository(spec.Workspace.RepositoryID)
	if repository == nil {
		return nil, fmt.Errorf("WorkerTemplate artifact repository %d was not resolved", id)
	}
	if repository.DefaultBranch == "" || spec.Workspace.Branch == "" {
		return nil, fmt.Errorf("WorkerTemplate artifact repository branch is missing")
	}
	return nil, fmt.Errorf(
		"WorkerTemplate artifact repository %d has no immutable commit pin",
		id,
	)
}

func buildSkillResolutions(
	scope control.Scope,
	refs ArtifactReferences,
	spec workerspec.Spec,
	workspace *workspaceResolver,
) ([]workerdependencyartifact.SkillResolution, error) {
	result := make([]workerdependencyartifact.SkillResolution, 0, len(spec.Workspace.SkillIDs))
	for _, id := range spec.Workspace.SkillIDs {
		row := workspace.skills[id]
		reference := refs.Skills[id]
		pin, err := referencePin(scope, reference, id)
		if err != nil {
			return nil, err
		}
		slug, err := slugkit.NewFromTrusted(row.Slug)
		if err != nil {
			return nil, err
		}
		result = append(result, workerdependencyartifact.SkillResolution{
			ResourceResolution: pin, Slug: slug, Version: row.Version,
			ContentDigest: digestFromSHA(row.ContentSha),
			StorageKey:    row.StorageKey, PackageSize: row.PackageSize,
		})
	}
	return result, nil
}

func buildKnowledgeResolutions(
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
		return nil, fmt.Errorf(
			"WorkerTemplate artifact knowledge base %d has no immutable commit pin",
			id,
		)
	}
	return result, nil
}
