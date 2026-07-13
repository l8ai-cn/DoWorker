package skill

import (
	"context"
	"fmt"
	"strings"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
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

	pkg, err := s.packageFromGit(ctx, repoName, branchOf(repo))
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
		ContentSha:     pkg.ContentSha,
		StorageKey:     pkg.StorageKey,
		PackageSize:    pkg.PackageSize,
		Version:        1,
		CreatedByID:    &userID,
	}
	if repo.HTTPCloneURL != "" {
		u := repo.HTTPCloneURL
		row.HTTPCloneURL = &u
	}

	if err := s.store.Create(ctx, row); err != nil {
		s.cleanupRepo(ctx, repoName)
		return nil, fmt.Errorf("skill: persist row: %w", err)
	}
	return row, nil
}

// Update re-renders SKILL.md/skill.json from the patched fields, commits them to
// Git (source of truth), re-packages, bumps the version, and refreshes the DB
// cache row.
func (s *Service) Update(ctx context.Context, req *UpdateSkillRequest) (*skilldom.Skill, error) {
	row, err := s.store.GetByID(ctx, req.OrganizationID, req.SkillID)
	if err != nil {
		return nil, err
	}
	repoName := s.gitops.RepoNameFromPath(row.GitRepoPath)
	branch := branchOrDefault(row.DefaultBranch)

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrNameRequired
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

	body := ""
	if req.Instructions != nil {
		body = *req.Instructions
	} else {
		data, _, rerr := s.gitops.ReadFile(ctx, repoName, branch, "SKILL.md")
		if rerr != nil {
			return nil, fmt.Errorf("skill: read current SKILL.md: %w", rerr)
		}
		body = extractSkillBody(string(data))
	}
	if strings.TrimSpace(body) == "" {
		return nil, ErrInstructionsRequired
	}

	files, err := renderSkillFiles(row.Slug, row.DisplayName, row.Description, row.License, body, row.Tags)
	if err != nil {
		return nil, err
	}
	if err := s.gitops.Commit(ctx, repoName, branch, "update: skill configuration", gitops.Author{}, files); err != nil {
		return nil, fmt.Errorf("skill: commit: %w", err)
	}

	pkg, err := s.packageFromGit(ctx, repoName, branch)
	if err != nil {
		return nil, err
	}
	row.ContentSha = pkg.ContentSha
	row.StorageKey = pkg.StorageKey
	row.PackageSize = pkg.PackageSize
	row.Version++

	if err := s.store.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("skill: update row: %w", err)
	}
	return row, nil
}

// packagedSkill decouples authoring from the extension package struct.
type packagedSkill struct {
	ContentSha  string
	StorageKey  string
	PackageSize int64
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
