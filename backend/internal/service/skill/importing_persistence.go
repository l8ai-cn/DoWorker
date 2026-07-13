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
	files, err = prepareImportedSkillFiles(slug, tags, files)
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

	pkg, err := s.packageImportedFiles(ctx, files)
	if err != nil {
		s.cleanupRepo(ctx, repoName)
		return nil, err
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

func (s *Service) packageImportedFiles(ctx context.Context, files []gitops.FileChange) (*packagedSkill, error) {
	dir, cleanup, err := materializeFileChanges(files)
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
		Created:     pkg.Created,
	}, nil
}
