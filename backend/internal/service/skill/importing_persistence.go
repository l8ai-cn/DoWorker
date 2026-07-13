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

	prepared, err := s.prepareImportedFiles(ctx, files)
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
		ContentSha:        prepared.ContentSha,
		StorageKey:        prepared.StorageKey,
		PackageSize:       prepared.PackageSize,
		Version:           1,
		CreatedByID:       &userID,
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

func (s *Service) prepareImportedFiles(
	ctx context.Context,
	files []gitops.FileChange,
) (*extensionsvc.PreparedSkill, error) {
	dir, cleanup, err := materializeFileChanges(files)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	prepared, err := s.packager.PrepareFromDir(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("skill: package: %w", err)
	}
	return prepared, nil
}
