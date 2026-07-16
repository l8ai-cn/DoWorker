package skill

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func (s *Service) EnsurePlatformSkill(
	ctx context.Context,
	req *EnsurePlatformSkillRequest,
) (*skilldom.Skill, bool, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, false, ErrNameRequired
	}
	if strings.TrimSpace(req.Instructions) == "" {
		return nil, false, ErrInstructionsRequired
	}
	tags, err := ValidateTags(req.Tags)
	if err != nil {
		return nil, false, err
	}
	createRequest := &CreateSkillRequest{
		OrganizationID: req.RepositoryOwnerOrganizationID,
		UserID:         req.UserID,
		Slug:           strings.TrimSpace(req.Slug),
		Name:           req.Name,
		Description:    req.Description,
		License:        req.License,
		Instructions:   req.Instructions,
		Tags:           tags,
	}
	if err := slugkit.ValidateIdentifier("skills.slug", createRequest.Slug); err != nil {
		return nil, false, err
	}
	files, err := renderSkillFiles(
		createRequest.Slug,
		createRequest.Name,
		createRequest.Description,
		createRequest.License,
		createRequest.Instructions,
		tags,
	)
	if err != nil {
		return nil, false, err
	}
	existing, err := s.store.GetPlatformBySlug(ctx, createRequest.Slug)
	if err == nil {
		matches, matchErr := s.platformSkillMatches(ctx, existing, createRequest, files)
		if matchErr != nil {
			return nil, false, matchErr
		}
		if !matches {
			return nil, false, ErrPlatformSkillConflict
		}
		return existing, false, nil
	}
	if !errors.Is(err, skilldom.ErrNotFound) {
		return nil, false, err
	}
	created, err := s.create(ctx, createRequest, true)
	return created, err == nil, err
}

func (s *Service) platformSkillMatches(
	ctx context.Context,
	existing *skilldom.Skill,
	request *CreateSkillRequest,
	files []gitops.FileChange,
) (bool, error) {
	if existing == nil ||
		existing.OrganizationID != nil ||
		!existing.IsActive ||
		existing.ContentSha == "" ||
		existing.StorageKey == "" ||
		existing.PackageSize <= 0 ||
		existing.DisplayName != strings.TrimSpace(request.Name) ||
		existing.Description != strings.TrimSpace(request.Description) ||
		existing.License != strings.TrimSpace(request.License) ||
		!slices.Equal(
			[]string(existing.Tags),
			[]string(skilldom.NormalizeTags(request.Tags)),
		) {
		return false, nil
	}
	repoName := s.gitops.RepoNameFromPath(existing.GitRepoPath)
	for _, file := range files {
		content, _, err := s.gitops.ReadFile(
			ctx,
			repoName,
			branchOrDefault(existing.DefaultBranch),
			file.Path,
		)
		if err != nil {
			return false, fmt.Errorf("skill: read platform skill %s: %w", file.Path, err)
		}
		if !slices.Equal(content, file.Content) {
			return false, nil
		}
	}
	return true, nil
}
