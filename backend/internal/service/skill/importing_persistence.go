package skill

import (
	"context"
	"fmt"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

func (s *Service) createImportedSkill(
	ctx context.Context, req *ImportFromGitRequest,
	src *extensionsvc.ClonedSkillSource, info extensionsvc.SkillInfo,
	subdir string, files []gitops.FileChange,
) (*skilldom.Skill, error) {
	slug, err := s.resolveSlug(ctx, req.OrganizationID, "", info.Slug, 0)
	if err != nil {
		return nil, err
	}
	tags := skilldom.NormalizeTags(info.Tags)
	files, err = prepareImportedSkillFiles(info.DirPath, slug, tags, files)
	if err != nil {
		return nil, err
	}

	repo, err := s.gitops.Provision(ctx, gitops.ProvisionParams{
		OrgID:         req.OrganizationID,
		Slug:          slug,
		CommitMessage: fmt.Sprintf("import: %s (%s)", req.URL, shortSha(src.CommitSha)),
		Seed:          files,
	})
	if err != nil {
		return nil, fmt.Errorf("skill: provision repo: %w", err)
	}
	repoName := s.gitops.RepoNameFromPath(repo.Path)

	pkg, err := s.packager.PackageFromDir(ctx, info.DirPath)
	if err != nil {
		s.cleanupRepo(ctx, repoName)
		return nil, fmt.Errorf("skill: package: %w", err)
	}

	orgID := req.OrganizationID
	userID := req.UserID
	row := &skilldom.Skill{
		OrganizationID:    &orgID,
		Slug:              slug,
		DisplayName:       displayNameOr(info.DisplayName, slug),
		Description:       info.Description,
		License:           info.License,
		Category:          info.Category,
		Compatibility:     info.Compatibility,
		AllowedTools:      info.AllowedTools,
		Tags:              tags,
		AgentFilter:       marshalAgentFilter(req.AgentFilter),
		IsActive:          true,
		GitRepoPath:       repo.Path,
		DefaultBranch:     branchOf(repo),
		UpstreamURL:       req.URL,
		UpstreamSubdir:    subdir,
		UpstreamCommitSha: src.CommitSha,
		InstallSource:     skilldom.SourceImport,
		ContentSha:        pkg.ContentSha,
		StorageKey:        pkg.StorageKey,
		PackageSize:       pkg.PackageSize,
		Version:           1,
		CreatedByID:       &userID,
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

func (s *Service) refreshImportedSkill(
	ctx context.Context, row *skilldom.Skill,
	src *extensionsvc.ClonedSkillSource, info extensionsvc.SkillInfo,
	files []gitops.FileChange,
) (*skilldom.Skill, error) {
	files, err := prepareImportedSkillFiles(info.DirPath, row.Slug, row.Tags, files)
	if err != nil {
		return nil, err
	}
	repoName := s.gitops.RepoNameFromPath(row.GitRepoPath)
	branch := branchOrDefault(row.DefaultBranch)

	if err := s.gitops.Commit(ctx, repoName, branch,
		fmt.Sprintf("sync: upstream %s", shortSha(src.CommitSha)), gitops.Author{}, files); err != nil {
		return nil, fmt.Errorf("skill: commit upstream sync: %w", err)
	}

	pkg, err := s.packager.PackageFromDir(ctx, info.DirPath)
	if err != nil {
		return nil, fmt.Errorf("skill: package: %w", err)
	}
	if pkg.ContentSha != row.ContentSha {
		row.Version++
	}
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

	if err := s.store.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("skill: update row: %w", err)
	}
	return row, nil
}
