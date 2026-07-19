package workercreation

import (
	"context"
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	"github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
	"github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type RepositoryLookup interface {
	GetAccessibleByID(context.Context, int64, int64, int64) (*gitprovider.Repository, error)
}

type SkillLookup interface {
	GetAnyByID(context.Context, int64) (*skill.Skill, error)
}

type KnowledgeLookup interface {
	Get(context.Context, int64, int64) (*knowledgebase.KnowledgeBase, error)
}

type EnvBundleLookup interface {
	GetByID(context.Context, int64) (*envbundle.EnvBundle, error)
}

type workspaceResolverDeps struct {
	Repositories RepositoryLookup
	Skills       SkillLookup
	Knowledge    KnowledgeLookup
	EnvBundles   EnvBundleLookup
	Definitions  WorkerDefinitionProvider
	Commits      WorkspaceCommitResolver
}

type workspaceResolver struct {
	deps         workspaceResolverDeps
	repositories map[int64]*gitprovider.Repository
	skills       map[int64]*skill.Skill
	knowledge    map[int64]*knowledgebase.KnowledgeBase
	envBundles   map[int64]*envbundle.EnvBundle
}

type knowledgeReference struct {
	Slug string
	Mode specdomain.KnowledgeMountMode
}

type compilationReferences struct {
	RepositorySlug    string
	SkillSlugs        []string
	Knowledge         []knowledgeReference
	EnvBundleNames    []string
	ConfigDocumentIDs []string
}

func newWorkspaceResolver(deps workspaceResolverDeps) *workspaceResolver {
	return &workspaceResolver{
		deps:         deps,
		repositories: make(map[int64]*gitprovider.Repository),
		skills:       make(map[int64]*skill.Skill),
		knowledge:    make(map[int64]*knowledgebase.KnowledgeBase),
		envBundles:   make(map[int64]*envbundle.EnvBundle),
	}
}

func (resolver *workspaceResolver) ResolveWorkspace(
	ctx context.Context,
	scope specservice.Scope,
	workerType slugkit.Slug,
	workspace specdomain.Workspace,
) (specdomain.Workspace, error) {
	if _, err := resolver.resolveWorkspaceReferences(ctx, scope, workerType, workspace); err != nil {
		return specdomain.Workspace{}, err
	}
	if len(workspace.SkillPackages) == 0 && len(workspace.SkillIDs) > 0 {
		workspace.SkillPackages = resolver.resolvedSkillPackages(workspace.SkillIDs)
	}
	return cloneResolvedWorkspace(workspace), nil
}

func (resolver *workspaceResolver) resolveWorkspaceReferences(
	ctx context.Context,
	scope specservice.Scope,
	workerType slugkit.Slug,
	workspace specdomain.Workspace,
) (compilationReferences, error) {
	var references compilationReferences
	skillSlugs := make(map[string]int64, len(workspace.SkillIDs))
	if workspace.RepositoryID != nil {
		repository, err := resolver.resolveRepository(ctx, scope, *workspace.RepositoryID)
		if err != nil {
			return compilationReferences{}, err
		}
		references.RepositorySlug = repository.Slug
	}
	if len(workspace.SkillPackages) > 0 {
		for _, pkg := range workspace.SkillPackages {
			references.SkillSlugs = append(references.SkillSlugs, pkg.Slug)
		}
	} else {
		for _, id := range workspace.SkillIDs {
			row, err := resolver.resolveSkill(ctx, scope, workerType, id)
			if err != nil {
				return compilationReferences{}, err
			}
			if existingID, exists := skillSlugs[row.Slug]; exists {
				return compilationReferences{}, fmt.Errorf(
					"%w: skills %d and %d share slug %q",
					specservice.ErrInvalidDraft,
					existingID,
					id,
					row.Slug,
				)
			}
			skillSlugs[row.Slug] = id
			references.SkillSlugs = append(references.SkillSlugs, row.Slug)
		}
	}
	for _, mount := range workspace.KnowledgeMounts {
		row, err := resolver.resolveKnowledge(ctx, scope, mount.KnowledgeBaseID)
		if err != nil {
			return compilationReferences{}, err
		}
		references.Knowledge = append(references.Knowledge, knowledgeReference{
			Slug: row.Slug,
			Mode: mount.Mode,
		})
	}
	for _, id := range workspace.EnvBundleIDs {
		bundle, err := resolver.resolveEnvBundle(ctx, scope, workerType, int64(id))
		if err != nil {
			return compilationReferences{}, err
		}
		if bundle.Kind != envbundle.KindRuntime && bundle.Kind != envbundle.KindShared {
			return compilationReferences{}, invalidWorkspaceReference(
				"runtime environment bundle",
				bundle.ID,
				"bundle kind is not runtime-safe",
				nil,
			)
		}
		definition, err := resolver.workerDefinition(workerType)
		if err != nil {
			return compilationReferences{}, err
		}
		if field := modelResourceManagedRuntimeField(definition, bundle.Data); field != "" {
			return compilationReferences{}, invalidWorkspaceReference(
				"runtime environment bundle",
				bundle.ID,
				fmt.Sprintf("field %q is managed by the model resource", field),
				nil,
			)
		}
		if err := appendEnvBundleName(&references, bundle.Name); err != nil {
			return compilationReferences{}, err
		}
	}
	configDocumentIDs, err := resolver.resolveConfigDocumentIDs(
		ctx,
		scope,
		workerType,
		workspace.ConfigDocumentBindings,
	)
	if err != nil {
		return compilationReferences{}, err
	}
	references.ConfigDocumentIDs = configDocumentIDs
	return references, nil
}

func (resolver *workspaceResolver) workerDefinition(
	workerType slugkit.Slug,
) (workerdefinition.Definition, error) {
	if resolver == nil || resolver.deps.Definitions == nil {
		return workerdefinition.Definition{}, specservice.ErrResolverUnavailable
	}
	definition, exists := resolver.deps.Definitions.Get(workerType.String())
	if !exists {
		return workerdefinition.Definition{}, invalidWorkerType(
			fmt.Sprintf("canonical definition for %q does not exist", workerType),
		)
	}
	return definition, nil
}

func (resolver *workspaceResolver) resolvedRepository(id *int64) *gitprovider.Repository {
	if resolver == nil || id == nil {
		return nil
	}
	return resolver.repositories[*id]
}
