package workercreation

import (
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdependencyartifact"
)

type workspaceDependencies struct {
	repository *workerdependencyartifact.RepositoryResolution
	skills     []workerdependencyartifact.SkillResolution
	knowledge  []workerdependencyartifact.KnowledgeBaseResolution
	bundles    []workerdependencyartifact.RuntimeBundleResolution
	secrets    []workerdependencyartifact.SecretReferenceResolution
}

func buildWorkspaceDependencies(
	scope control.Scope,
	refs ArtifactReferences,
	spec workerspec.Spec,
	definition workerdefinition.Definition,
	workspace *workspaceResolver,
) (workspaceDependencies, error) {
	repository, err := buildRepositoryResolution(scope, refs, spec, workspace)
	if err != nil {
		return workspaceDependencies{}, err
	}
	skills, err := buildSkillResolutions(scope, refs, spec, workspace)
	if err != nil {
		return workspaceDependencies{}, err
	}
	knowledge, err := buildKnowledgeResolutions(scope, refs, spec, workspace)
	if err != nil {
		return workspaceDependencies{}, err
	}
	bundles, err := buildBundleResolutions(scope, refs, spec, definition, workspace)
	if err != nil {
		return workspaceDependencies{}, err
	}
	secrets, err := buildSecretResolutions(scope, refs, spec, definition, workspace)
	if err != nil {
		return workspaceDependencies{}, err
	}
	return workspaceDependencies{
		repository: repository, skills: skills, knowledge: knowledge,
		bundles: bundles, secrets: secrets,
	}, nil
}

func referencePin(
	scope control.Scope,
	reference control.ResolvedReference,
	domainID int64,
) (workerdependencyartifact.ResourceResolution, error) {
	if reference.Kind == "" {
		return workerdependencyartifact.ResourceResolution{}, fmt.Errorf(
			"WorkerTemplate artifact resource reference is missing",
		)
	}
	return workerdependencyartifact.BindResourceProjection(scope, reference, domainID)
}
