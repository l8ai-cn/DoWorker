package extension

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func (s *Service) GetWorkerSkillsByPackages(
	ctx context.Context,
	packages []specdomain.SkillPackageBinding,
	_ string,
) ([]*ResolvedSkill, error) {
	if s == nil || s.storage == nil {
		return nil, fmt.Errorf("%w: worker skill resolver is unavailable", ErrInvalidInput)
	}
	resolved := make([]*ResolvedSkill, 0, len(packages))
	for _, pkg := range packages {
		if pkg.SkillID <= 0 || pkg.Version <= 0 || pkg.PackageSize < 0 ||
			pkg.ContentSHA == "" || pkg.StorageKey == "" {
			return nil, fmt.Errorf("%w: pinned skill package is incomplete", ErrInvalidInput)
		}
		if err := slugkit.Validate(pkg.Slug); err != nil {
			return nil, fmt.Errorf("%w: pinned skill slug: %v", ErrInvalidInput, err)
		}
		downloadURL, err := s.storage.GetInternalURL(
			ctx,
			pkg.StorageKey,
			presignedURLExpiry,
		)
		if err != nil {
			return nil, fmt.Errorf("sign pinned worker skill %d: %w", pkg.SkillID, err)
		}
		resolved = append(resolved, &ResolvedSkill{
			CatalogSkillID: pkg.SkillID,
			Slug:           pkg.Slug,
			ContentSha:     pkg.ContentSHA,
			DownloadURL:    downloadURL,
			PackageSize:    pkg.PackageSize,
			TargetDir:      fmt.Sprintf("skills/%s", pkg.Slug),
		})
	}
	return resolved, nil
}

func (s *Service) GetWorkerSkillsByIDs(
	ctx context.Context,
	orgID int64,
	ids []int64,
	agentSlug string,
) ([]*ResolvedSkill, error) {
	if s == nil || s.catalog == nil || s.storage == nil {
		return nil, fmt.Errorf("%w: worker skill resolver is unavailable", ErrInvalidInput)
	}
	resolved := make([]*ResolvedSkill, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, fmt.Errorf("%w: skill id must be positive", ErrInvalidInput)
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("%w: duplicate skill id %d", ErrInvalidInput, id)
		}
		seen[id] = struct{}{}
		row, err := s.catalog.GetAnyByID(ctx, id)
		if err != nil || row == nil || row.ID != id || !row.IsActive {
			return nil, fmt.Errorf("%w: skill %d", ErrNotFound, id)
		}
		if !row.VisibleTo(orgID) {
			return nil, fmt.Errorf("%w: skill %d", ErrForbidden, id)
		}
		if err := slugkit.Validate(row.Slug); err != nil {
			return nil, fmt.Errorf("%w: skill %d slug: %v", ErrInvalidInput, id, err)
		}
		supportsAgent, err := workerSkillSupportsAgent(row, agentSlug)
		if err != nil {
			return nil, fmt.Errorf("%w: skill %d agent filter: %v", ErrInvalidInput, id, err)
		}
		if !supportsAgent {
			return nil, fmt.Errorf(
				"%w: skill %d does not support agent %q",
				ErrInvalidInput,
				id,
				agentSlug,
			)
		}
		if row.ContentSha == "" || row.StorageKey == "" {
			return nil, fmt.Errorf("%w: skill %d package is unavailable", ErrInvalidInput, id)
		}
		downloadURL, err := s.storage.GetInternalURL(
			ctx,
			row.StorageKey,
			presignedURLExpiry,
		)
		if err != nil {
			return nil, fmt.Errorf("sign worker skill %d: %w", id, err)
		}
		resolved = append(resolved, &ResolvedSkill{
			CatalogSkillID: row.ID,
			Slug:           row.Slug,
			ContentSha:     row.ContentSha,
			DownloadURL:    downloadURL,
			PackageSize:    row.PackageSize,
			TargetDir:      fmt.Sprintf("skills/%s", row.Slug),
		})
	}
	return resolved, nil
}

func workerSkillSupportsAgent(row *skill.Skill, agentSlug string) (bool, error) {
	if len(row.AgentFilter) == 0 {
		return true, nil
	}
	var filter []string
	if err := json.Unmarshal(row.AgentFilter, &filter); err != nil {
		return false, err
	}
	if len(filter) == 0 {
		return true, nil
	}
	for _, allowed := range filter {
		if agentSlugMatches(allowed, agentSlug) {
			return true, nil
		}
	}
	return false, nil
}
