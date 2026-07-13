package skill

import (
	"context"
	"fmt"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

func (s *Service) refreshImportedSkill(
	ctx context.Context,
	initial *skilldom.Skill,
	src *extensionsvc.ClonedSkillSource,
	info extensionsvc.SkillInfo,
	upstreamFiles []gitops.FileChange,
) (*skilldom.Skill, error) {
	row := initial
	for attempt := 0; attempt < maxSkillMutationAttempts; attempt++ {
		if attempt > 0 {
			var err error
			row, err = s.store.GetByID(ctx, *initial.OrganizationID, initial.ID)
			if err != nil {
				return nil, err
			}
		}
		updated, conflict, err := s.refreshImportedSkillOnce(
			ctx, row, src, info, upstreamFiles,
		)
		if err != nil {
			return nil, err
		}
		if !conflict {
			return updated, nil
		}
	}
	return nil, ErrMutationConflict
}

func (s *Service) refreshImportedSkillOnce(
	ctx context.Context,
	row *skilldom.Skill,
	src *extensionsvc.ClonedSkillSource,
	info extensionsvc.SkillInfo,
	upstreamFiles []gitops.FileChange,
) (*skilldom.Skill, bool, error) {
	files, err := prepareImportedSkillFiles(row.Slug, row.Tags, upstreamFiles)
	if err != nil {
		return nil, false, err
	}
	repoName := s.gitops.RepoNameFromPath(row.GitRepoPath)
	branch := branchOrDefault(row.DefaultBranch)
	if err := s.gitops.Commit(ctx, repoName, branch,
		fmt.Sprintf("sync: upstream %s", shortSha(src.CommitSha)),
		gitops.Author{}, files); err != nil {
		return nil, false, fmt.Errorf("skill: commit upstream sync: %w", err)
	}
	pkg, err := s.packageImportedFiles(ctx, files)
	if err != nil {
		return nil, false, err
	}
	expectedVersion := row.Version
	applyImportedSkillRefresh(row, src, info, pkg)
	row.Version = expectedVersion + 1
	updated, err := s.store.UpdateIfVersion(ctx, row, expectedVersion)
	if err != nil {
		return nil, false, fmt.Errorf("skill: update row: %w", err)
	}
	return row, !updated, nil
}

func applyImportedSkillRefresh(
	row *skilldom.Skill,
	src *extensionsvc.ClonedSkillSource,
	info extensionsvc.SkillInfo,
	pkg *packagedSkill,
) {
	row.DisplayName = displayNameOr(info.DisplayName, row.Slug)
	row.Description = info.Description
	row.License = info.License
	row.Category = info.Category
	row.Compatibility = info.Compatibility
	row.AllowedTools = info.AllowedTools
	row.UpstreamCommitSha = src.CommitSha
	row.ContentSha = pkg.ContentSha
	row.StorageKey = pkg.StorageKey
	row.PackageSize = pkg.PackageSize
}
