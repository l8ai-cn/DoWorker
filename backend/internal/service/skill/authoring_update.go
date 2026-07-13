package skill

import (
	"context"
	"errors"
	"fmt"
	"strings"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

const maxSkillMutationAttempts = 4

var ErrMutationConflict = errors.New("skill: concurrent mutation conflict")

func (s *Service) Update(ctx context.Context, req *UpdateSkillRequest) (*skilldom.Skill, error) {
	var result *skilldom.Skill
	err := s.store.WithMutationLock(ctx, req.SkillID, func(store skilldom.Repository) error {
		for attempt := 0; attempt < maxSkillMutationAttempts; attempt++ {
			row, err := store.GetByID(ctx, req.OrganizationID, req.SkillID)
			if err != nil {
				return err
			}
			updated, conflict, err := s.updateOnce(ctx, store, row, req)
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

func (s *Service) updateOnce(
	ctx context.Context,
	store skilldom.Repository,
	row *skilldom.Skill,
	req *UpdateSkillRequest,
) (*skilldom.Skill, bool, error) {
	repoName := s.gitops.RepoNameFromPath(row.GitRepoPath)
	branch := branchOrDefault(row.DefaultBranch)
	snapshot, err := gitops.CaptureTree(ctx, s.gitops, repoName, branch)
	if err != nil {
		return nil, false, err
	}
	if err := applySkillUpdate(row, req); err != nil {
		return nil, false, err
	}
	body, err := s.updatedSkillBody(ctx, repoName, branch, req.Instructions)
	if err != nil {
		return nil, false, err
	}
	files, err := s.renderUpdatedSkillFiles(ctx, repoName, branch, row, body)
	if err != nil {
		return nil, false, err
	}
	if err := s.gitops.Commit(ctx, repoName, branch,
		"update: skill configuration", gitops.Author{}, files); err != nil {
		return nil, false, fmt.Errorf("skill: commit: %w", err)
	}
	prepared, err := s.prepareFromGit(ctx, repoName, branch)
	if err != nil {
		return nil, false, s.restoreMutation(ctx, repoName, branch, snapshot, err)
	}
	expectedVersion := row.Version
	conflict, err := s.publishPreparedPackage(
		ctx,
		store,
		prepared,
		func(locked skilldom.Repository, pkg *extensionsvc.PackagedSkill) (bool, error) {
			applyStoredPackage(row, pkg)
			row.Version = expectedVersion + 1
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

func applySkillUpdate(row *skilldom.Skill, req *UpdateSkillRequest) error {
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return ErrNameRequired
		}
		row.DisplayName = name
	}
	if req.Description != nil {
		row.Description = strings.TrimSpace(*req.Description)
	}
	if req.License != nil {
		row.License = strings.TrimSpace(*req.License)
	}
	if req.Tags != nil {
		row.Tags = skilldom.NormalizeTags(*req.Tags)
	}
	return nil
}

func (s *Service) updatedSkillBody(
	ctx context.Context,
	repoName, branch string,
	instructions *string,
) (string, error) {
	if instructions != nil {
		if strings.TrimSpace(*instructions) == "" {
			return "", ErrInstructionsRequired
		}
		return *instructions, nil
	}
	data, _, err := s.gitops.ReadFile(ctx, repoName, branch, "SKILL.md")
	if err != nil {
		return "", fmt.Errorf("skill: read current SKILL.md: %w", err)
	}
	body := extractSkillBody(string(data))
	if strings.TrimSpace(body) == "" {
		return "", ErrInstructionsRequired
	}
	return body, nil
}

func (s *Service) renderUpdatedSkillFiles(
	ctx context.Context,
	repoName, branch string,
	row *skilldom.Skill,
	body string,
) ([]gitops.FileChange, error) {
	files, err := renderSkillFiles(
		row.Slug, row.DisplayName, row.Description, row.License, body, row.Tags,
	)
	if err != nil {
		return nil, err
	}
	current, _, err := s.gitops.ReadFile(ctx, repoName, branch, "skill.json")
	if err != nil {
		return nil, fmt.Errorf("skill: read current skill.json: %w", err)
	}
	files[1].Content, err = mergeAuthoredSkillConfig(current, files[1].Content)
	return files, err
}
