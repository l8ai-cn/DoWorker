package skill

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
)

func (s *PlatformCatalogService) publish(
	ctx context.Context,
	req *EnsurePlatformSkillRequest,
	tags []string,
	prepared *extensionsvc.PreparedSkill,
) (*skilldom.Skill, bool, error) {
	var row *skilldom.Skill
	created := false
	conflict := false
	err := s.store.WithPackageLock(
		ctx,
		operatorCatalogIdentityPrefix+req.Slug,
		func(locked skilldom.Repository) error {
			var publishErr error
			conflict, publishErr = s.publishLocked(
				ctx, locked, req, tags, prepared, &row, &created,
			)
			return publishErr
		},
	)
	if err != nil {
		return nil, false, fmt.Errorf("skill: publish operator catalog: %w", err)
	}
	if conflict {
		return nil, false, ErrPlatformSkillConflict
	}
	return row, created, nil
}

func (s *PlatformCatalogService) publishLocked(
	ctx context.Context,
	store skilldom.Repository,
	req *EnsurePlatformSkillRequest,
	tags []string,
	prepared *extensionsvc.PreparedSkill,
	row **skilldom.Skill,
	created *bool,
) (bool, error) {
	return publishPreparedPackage(
		ctx,
		store,
		s.packager,
		prepared,
		func(locked skilldom.Repository, pkg *extensionsvc.PackagedSkill) (bool, error) {
			existing, getErr := locked.GetPlatformBySlug(ctx, req.Slug)
			if getErr == nil {
				*row = existing
				return !platformSkillMatches(existing, req, tags, pkg), nil
			}
			if !errors.Is(getErr, skilldom.ErrNotFound) {
				return false, getErr
			}
			userID := req.UserID
			*row = newOperatorSkill(req, tags, userID)
			applyStoredPackage(*row, pkg)
			if createErr := locked.Create(ctx, *row); createErr != nil {
				return false, createErr
			}
			*created = true
			return false, nil
		},
		func(
			locked skilldom.Repository,
			pkg *extensionsvc.PackagedSkill,
			cause error,
		) error {
			return cleanupCreatedPackage(ctx, locked, s.packager, pkg, cause)
		},
	)
}

func newOperatorSkill(
	req *EnsurePlatformSkillRequest,
	tags []string,
	userID int64,
) *skilldom.Skill {
	return &skilldom.Skill{
		Slug:          req.Slug,
		DisplayName:   strings.TrimSpace(req.Name),
		Description:   strings.TrimSpace(req.Description),
		License:       strings.TrimSpace(req.License),
		Tags:          tags,
		IsActive:      true,
		DefaultBranch: "main",
		InstallSource: skilldom.SourceOperator,
		Version:       1,
		CreatedByID:   &userID,
	}
}

func platformSkillMatches(
	existing *skilldom.Skill,
	req *EnsurePlatformSkillRequest,
	tags []string,
	pkg *extensionsvc.PackagedSkill,
) bool {
	return existing != nil &&
		existing.OrganizationID == nil &&
		existing.IsActive &&
		existing.GitRepoPath == "" &&
		existing.InstallSource == skilldom.SourceOperator &&
		existing.DisplayName == strings.TrimSpace(req.Name) &&
		existing.Description == strings.TrimSpace(req.Description) &&
		existing.License == strings.TrimSpace(req.License) &&
		slices.Equal([]string(existing.Tags), []string(skilldom.NormalizeTags(tags))) &&
		existing.ContentSha == pkg.ContentSha &&
		existing.StorageKey == pkg.StorageKey &&
		existing.PackageSize == pkg.PackageSize
}
