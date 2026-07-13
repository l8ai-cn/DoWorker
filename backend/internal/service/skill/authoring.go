package skill

import (
	"context"
	"fmt"
	"strings"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

// Create provisions an am-skills repo (seeding SKILL.md + skill.json), packages
// it through the extension bridge, and records the DB cache row. On any
// post-provision failure the fresh repo is compensating-deleted.
func (s *Service) Create(ctx context.Context, req *CreateSkillRequest) (*skilldom.Skill, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, ErrNameRequired
	}
	if strings.TrimSpace(req.Instructions) == "" {
		return nil, ErrInstructionsRequired
	}
	slug, err := s.resolveSlug(ctx, req.OrganizationID, req.Slug, req.Name, 0)
	if err != nil {
		return nil, err
	}

	tags := skilldom.NormalizeTags(req.Tags)
	files, err := renderSkillFiles(slug, req.Name, req.Description, req.License, req.Instructions, tags)
	if err != nil {
		return nil, err
	}

	repo, err := s.gitops.Provision(ctx, gitops.ProvisionParams{
		OrgID:         req.OrganizationID,
		Slug:          slug,
		CommitMessage: "init: skill scaffold (SKILL.md, skill.json)",
		Seed:          files,
	})
	if err != nil {
		return nil, fmt.Errorf("skill: provision repo: %w", err)
	}
	repoName := s.gitops.RepoNameFromPath(repo.Path)

	prepared, err := s.prepareFromGit(ctx, repoName, branchOf(repo))
	if err != nil {
		s.cleanupRepo(ctx, repoName)
		return nil, err
	}

	orgID := req.OrganizationID
	userID := req.UserID
	row := &skilldom.Skill{
		OrganizationID: &orgID,
		Slug:           slug,
		DisplayName:    strings.TrimSpace(req.Name),
		Description:    strings.TrimSpace(req.Description),
		License:        strings.TrimSpace(req.License),
		Tags:           tags,
		IsActive:       true,
		GitRepoPath:    repo.Path,
		DefaultBranch:  branchOf(repo),
		InstallSource:  skilldom.SourceGitops,
		ContentSha:     prepared.ContentSha,
		StorageKey:     prepared.StorageKey,
		PackageSize:    prepared.PackageSize,
		Version:        1,
		CreatedByID:    &userID,
	}
	if repo.HTTPCloneURL != "" {
		u := repo.HTTPCloneURL
		row.HTTPCloneURL = &u
	}

	_, err = s.publishPreparedPackage(
		ctx,
		s.store,
		prepared,
		func(store skilldom.Repository, pkg *extensionsvc.PackagedSkill) (bool, error) {
			applyStoredPackage(row, pkg)
			return false, store.Create(ctx, row)
		},
		func(
			store skilldom.Repository,
			pkg *extensionsvc.PackagedSkill,
			cause error,
		) error {
			return s.cleanupCreatedPackage(
				ctx,
				store,
				pkg,
				fmt.Errorf("skill: persist row: %w", cause),
			)
		},
	)
	if err != nil {
		s.cleanupRepo(ctx, repoName)
		return nil, err
	}
	return row, nil
}

func applyStoredPackage(
	row *skilldom.Skill,
	pkg *extensionsvc.PackagedSkill,
) {
	row.ContentSha = pkg.ContentSha
	row.StorageKey = pkg.StorageKey
	row.PackageSize = pkg.PackageSize
}

func (s *Service) resolveSlug(ctx context.Context, orgID int64, explicit, nameSeed string, excludeID int64) (string, error) {
	seed := strings.TrimSpace(explicit)
	if seed == "" {
		seed = nameSeed
	}
	return slugkit.GenerateUnique(ctx, seed, slugkit.FromExistsCheck(func(ctx context.Context, candidate string) (bool, error) {
		return s.store.SlugExists(ctx, orgID, candidate, excludeID)
	}))
}
