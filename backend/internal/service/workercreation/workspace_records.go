package workercreation

import (
	"context"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/knowledgebase"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func (resolver *workspaceResolver) resolveKnowledge(
	ctx context.Context,
	scope specservice.Scope,
	id int64,
) (*knowledgebase.KnowledgeBase, error) {
	if resolver == nil {
		return nil, specservice.ErrResolverUnavailable
	}
	row := resolver.knowledge[id]
	if row == nil {
		if resolver.deps.Knowledge == nil {
			return nil, specservice.ErrResolverUnavailable
		}
		var err error
		row, err = resolver.deps.Knowledge.Get(ctx, scope.OrgID, id)
		if err != nil {
			if err == knowledgebase.ErrNotFound {
				return nil, invalidWorkspaceReference("knowledge base", id, "not found", err)
			}
			return nil, err
		}
	}
	if row == nil || row.ID != id || row.OrganizationID != scope.OrgID {
		return nil, invalidWorkspaceReference("knowledge base", id, "not accessible", nil)
	}
	resolver.knowledge[id] = row
	return row, nil
}

func (resolver *workspaceResolver) resolveEnvBundle(
	ctx context.Context,
	scope specservice.Scope,
	workerType slugkit.Slug,
	id int64,
) (*envbundle.EnvBundle, error) {
	if resolver == nil {
		return nil, specservice.ErrResolverUnavailable
	}
	bundle := resolver.envBundles[id]
	if bundle == nil {
		if resolver.deps.EnvBundles == nil {
			return nil, specservice.ErrResolverUnavailable
		}
		var err error
		bundle, err = resolver.deps.EnvBundles.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
	}
	if bundle == nil || bundle.ID != id || !bundle.IsActive {
		return nil, invalidWorkspaceReference("environment bundle", id, "not accessible", nil)
	}
	if !envBundleVisibleTo(bundle, scope) {
		return nil, invalidWorkspaceReference("environment bundle", id, "not accessible", nil)
	}
	if bundle.AgentSlug != nil && *bundle.AgentSlug != workerType.String() {
		return nil, invalidWorkspaceReference(
			"environment bundle",
			id,
			fmt.Sprintf("not compatible with worker type %q", workerType),
			nil,
		)
	}
	resolver.envBundles[id] = bundle
	return bundle, nil
}

func envBundleVisibleTo(bundle *envbundle.EnvBundle, scope specservice.Scope) bool {
	switch bundle.OwnerScope {
	case envbundle.OwnerScopeUser:
		return bundle.OwnerID == scope.UserID
	case envbundle.OwnerScopeOrg:
		return bundle.OwnerID == scope.OrgID
	default:
		return false
	}
}

func cloneResolvedWorkspace(workspace specdomain.Workspace) specdomain.Workspace {
	cloned := workspace
	if workspace.RepositoryID != nil {
		id := *workspace.RepositoryID
		cloned.RepositoryID = &id
	}
	cloned.SkillIDs = append([]int64{}, workspace.SkillIDs...)
	cloned.SkillPackages = append(
		[]specdomain.SkillPackageBinding{},
		workspace.SkillPackages...,
	)
	cloned.KnowledgeMounts = append([]specdomain.KnowledgeMount{}, workspace.KnowledgeMounts...)
	cloned.EnvBundleIDs = append([]specdomain.RuntimeEnvBundleID{}, workspace.EnvBundleIDs...)
	return cloned
}

func (resolver *workspaceResolver) resolvedSkillPackages(
	ids []int64,
) []specdomain.SkillPackageBinding {
	packages := make([]specdomain.SkillPackageBinding, 0, len(ids))
	for _, id := range ids {
		row := resolver.skills[id]
		packages = append(packages, specdomain.SkillPackageBinding{
			SkillID:     row.ID,
			Slug:        row.Slug,
			Version:     row.Version,
			ContentSHA:  row.ContentSha,
			StorageKey:  row.StorageKey,
			PackageSize: row.PackageSize,
		})
	}
	return packages
}
