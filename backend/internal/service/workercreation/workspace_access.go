package workercreation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/gitprovider"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	repositoryservice "github.com/l8ai-cn/agentcloud/backend/internal/service/repository"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func (resolver *workspaceResolver) resolveRepository(
	ctx context.Context,
	scope specservice.Scope,
	id int64,
) (*gitprovider.Repository, error) {
	if resolver == nil {
		return nil, specservice.ErrResolverUnavailable
	}
	row := resolver.repositories[id]
	if row == nil {
		if resolver.deps.Repositories == nil {
			return nil, specservice.ErrResolverUnavailable
		}
		var err error
		row, err = resolver.deps.Repositories.GetAccessibleByID(ctx, id, scope.OrgID, scope.UserID)
		if err != nil {
			if errors.Is(err, repositoryservice.ErrNoPermission) ||
				errors.Is(err, repositoryservice.ErrRepositoryNotFound) {
				return nil, invalidWorkspaceReference("repository", id, "not accessible", err)
			}
			return nil, err
		}
	}
	if row == nil || row.ID != id || row.OrganizationID != scope.OrgID || !row.IsActive {
		return nil, invalidWorkspaceReference("repository", id, "not accessible", nil)
	}
	resolver.repositories[id] = row
	return row, nil
}

func (resolver *workspaceResolver) resolveSkill(
	ctx context.Context,
	scope specservice.Scope,
	workerType slugkit.Slug,
	id int64,
) (*skill.Skill, error) {
	if resolver == nil {
		return nil, specservice.ErrResolverUnavailable
	}
	row := resolver.skills[id]
	if row == nil {
		if resolver.deps.Skills == nil {
			return nil, specservice.ErrResolverUnavailable
		}
		var err error
		row, err = resolver.deps.Skills.GetAnyByID(ctx, id)
		if err != nil {
			if errors.Is(err, skill.ErrNotFound) {
				return nil, invalidWorkspaceReference("skill", id, "not found", err)
			}
			return nil, err
		}
	}
	if row == nil || row.ID != id || !row.VisibleTo(scope.OrgID) || !row.IsActive {
		return nil, invalidWorkspaceReference("skill", id, "not accessible", nil)
	}
	if err := slugkit.Validate(row.Slug); err != nil {
		return nil, invalidWorkspaceReference("skill", id, "slug is invalid", err)
	}
	if row.ContentSha == "" || row.StorageKey == "" {
		return nil, invalidWorkspaceReference("skill", id, "package is unavailable", nil)
	}
	allowed, err := skillAllowsWorker(row.AgentFilter, workerType.String())
	if err != nil {
		return nil, invalidWorkspaceReference("skill", id, err.Error(), nil)
	}
	if !allowed {
		return nil, invalidWorkspaceReference("skill", id, "not compatible with worker type", nil)
	}
	resolver.skills[id] = row
	return row, nil
}

func skillAllowsWorker(filter json.RawMessage, workerType string) (bool, error) {
	if len(filter) == 0 {
		return true, nil
	}
	var allowed []string
	if err := json.Unmarshal(filter, &allowed); err != nil {
		return false, fmt.Errorf("agent filter is invalid")
	}
	if len(allowed) == 0 {
		return true, nil
	}
	for _, slug := range allowed {
		if slug == workerType {
			return true, nil
		}
	}
	return false, nil
}

func invalidWorkspaceReference(field string, id int64, reason string, cause error) error {
	if cause != nil {
		return fmt.Errorf("%w: %s %d: %s: %w", specservice.ErrInvalidDraft, field, id, reason, cause)
	}
	return fmt.Errorf("%w: %s %d: %s", specservice.ErrInvalidDraft, field, id, reason)
}
