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
	var result *skilldom.Skill
	err := s.store.WithMutationLock(ctx, initial.ID, func(store skilldom.Repository) error {
		row := initial
		for attempt := 0; attempt < maxSkillMutationAttempts; attempt++ {
			if attempt > 0 {
				var err error
				row, err = store.GetByID(ctx, *initial.OrganizationID, initial.ID)
				if err != nil {
					return err
				}
			}
			updated, conflict, err := s.refreshImportedSkillOnce(
				ctx, store, row, src, info, upstreamFiles,
			)
			if err != nil {
				return err
			}
			if !conflict {
				result = updated
				return nil
			}
		}
		return ErrMutationConflict
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) refreshImportedSkillOnce(
	ctx context.Context,
	store skilldom.Repository,
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
	snapshot, err := gitops.CaptureTree(ctx, s.gitops, repoName, branch)
	if err != nil {
		return nil, false, err
	}
	if err := s.gitops.Commit(ctx, repoName, branch,
		fmt.Sprintf("sync: upstream %s", shortSha(src.CommitSha)),
		gitops.Author{}, files); err != nil {
		return nil, false, fmt.Errorf("skill: commit upstream sync: %w", err)
	}
	prepared, err := s.prepareImportedFiles(ctx, files)
	if err != nil {
		return nil, false, s.restoreMutation(ctx, repoName, branch, snapshot, err)
	}
	expectedVersion := row.Version
	previousContentSha := row.ContentSha
	conflict, err := s.publishPreparedPackage(
		ctx,
		store,
		prepared,
		func(locked skilldom.Repository, pkg *extensionsvc.PackagedSkill) (bool, error) {
			applyImportedSkillRefresh(row, src, info, pkg)
			if pkg.ContentSha != previousContentSha {
				row.Version = expectedVersion + 1
			}
			updated, updateErr := locked.UpdateIfVersion(ctx, row, expectedVersion)
			if updateErr != nil {
				return false, fmt.Errorf("skill: update row: %w", updateErr)
			}
			return !updated, nil
		},
		func(
			locked skilldom.Repository,
			pkg *extensionsvc.PackagedSkill,
			cause error,
		) error {
			return s.compensatePackagedMutation(
				ctx,
				locked,
				repoName,
				branch,
				snapshot,
				pkg,
				cause,
			)
		},
	)
	if err != nil {
		return nil, false, err
	}
	return row, conflict, nil
}

func applyImportedSkillRefresh(
	row *skilldom.Skill,
	src *extensionsvc.ClonedSkillSource,
	info extensionsvc.SkillInfo,
	pkg *extensionsvc.PackagedSkill,
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
