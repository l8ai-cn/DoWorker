package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

const skillConfigSchema = 1

// CreateSkillRequest authors a new git-backed skill.
type CreateSkillRequest struct {
	OrganizationID int64
	UserID         int64
	Slug           string // optional; derived from Name when empty
	Name           string // display name; also the SKILL.md frontmatter name
	Description    string
	License        string
	Instructions   string // SKILL.md body (markdown)
}

// UpdateSkillRequest patches an authored skill. Nil fields are left unchanged;
// when Instructions is nil the current SKILL.md body is preserved from Git.
type UpdateSkillRequest struct {
	OrganizationID int64
	SkillID        int64
	Name           *string
	Description    *string
	License        *string
	Instructions   *string
}

// skillConfig is the structured skill.json committed alongside SKILL.md.
type skillConfig struct {
	Schema      int    `json:"schema"`
	Slug        string `json:"slug"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	License     string `json:"license,omitempty"`
}

// renderSkillFiles builds SKILL.md (frontmatter shape parseFrontmatter
// understands: name/description/license) + skill.json. The frontmatter name is
// the slug so the packaged artifact's slug matches the authored slug.
func renderSkillFiles(slug, displayName, description, license, body string) ([]gitops.FileChange, error) {
	var md strings.Builder
	md.WriteString("---\n")
	fmt.Fprintf(&md, "name: %s\n", slug)
	if d := strings.TrimSpace(description); d != "" {
		fmt.Fprintf(&md, "description: %s\n", sanitizeFrontmatterValue(d))
	}
	if l := strings.TrimSpace(license); l != "" {
		fmt.Fprintf(&md, "license: %s\n", sanitizeFrontmatterValue(l))
	}
	md.WriteString("---\n\n")
	md.WriteString(strings.TrimRight(body, "\n"))
	md.WriteString("\n")

	cfg := skillConfig{
		Schema:      skillConfigSchema,
		Slug:        slug,
		Name:        strings.TrimSpace(displayName),
		Description: strings.TrimSpace(description),
		License:     strings.TrimSpace(license),
	}
	cfgJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("skill: render skill.json: %w", err)
	}

	return []gitops.FileChange{
		{Path: "SKILL.md", Content: []byte(md.String())},
		{Path: "skill.json", Content: append(cfgJSON, '\n')},
	}, nil
}

// sanitizeFrontmatterValue keeps a value single-line so it stays parseable by
// the flat key:value frontmatter reader.
func sanitizeFrontmatterValue(v string) string {
	v = strings.ReplaceAll(v, "\r", " ")
	v = strings.ReplaceAll(v, "\n", " ")
	return strings.TrimSpace(v)
}

// extractSkillBody returns the markdown body after the SKILL.md frontmatter
// block. When no frontmatter is present the whole content is treated as body.
func extractSkillBody(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return content
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.TrimLeft(strings.Join(lines[i+1:], "\n"), "\n")
		}
	}
	return content
}

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

	files, err := renderSkillFiles(slug, req.Name, req.Description, req.License, req.Instructions)
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

	files, err := renderSkillFiles(row.Slug, row.DisplayName, row.Description, row.License, body)
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

// Delete removes the DB row first (authoritative for existence), then
// best-effort deletes the backing repo (mirrors the expert/KB Delete ordering).
func (s *Service) Delete(ctx context.Context, orgID, id int64) error {
	row, err := s.store.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}
	if err := s.store.Delete(ctx, orgID, id); err != nil {
		return err
	}
	if row.GitRepoPath != "" {
		repoName := s.gitops.RepoNameFromPath(row.GitRepoPath)
		if delErr := s.gitops.DeleteRepo(ctx, repoName); delErr != nil {
			s.logger.Warn("skill: repo delete failed", "repo", repoName, "error", delErr)
		}
	}
	return nil
}

// Get returns one authored skill from the DB cache.
func (s *Service) Get(ctx context.Context, orgID int64, slug string) (*skilldom.Skill, error) {
	return s.store.GetBySlug(ctx, orgID, slug)
}

// List returns authored skills for an org from the DB cache.
func (s *Service) List(ctx context.Context, orgID int64, limit, offset int) ([]skilldom.Skill, int64, error) {
	return s.store.List(ctx, orgID, limit, offset)
}

// packageFromGit materializes the repo tree into a temp dir and runs it through
// the extension packager, cleaning up the temp dir afterward.
func (s *Service) packageFromGit(ctx context.Context, repoName, branch string) (*packagedSkill, error) {
	dir, cleanup, err := materializeRepo(ctx, s.gitops, repoName, branch)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	pkg, err := s.packager.PackageFromDir(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("skill: package: %w", err)
	}
	return &packagedSkill{
		ContentSha:  pkg.ContentSha,
		StorageKey:  pkg.StorageKey,
		PackageSize: pkg.PackageSize,
	}, nil
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
