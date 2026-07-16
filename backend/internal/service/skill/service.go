// Package skill is an independent, gitops-backed authoring source for skills
// (namespace am-skills). It provisions one repo per skill (SKILL.md +
// skill.json), commits edits, and bridges into the existing filesystem-based
// extension packager. It prepares bytes before taking the package publication
// lock, then stores them while the package and catalog mutation are serialized.
//
// This is ADDITIVE: it coexists with the external-import / marketplace install
// flow (service/extension), which is untouched. The service consumes
// gitops.Service (never the raw gitea client) and returns nil when git-backing
// or the packager is not configured (feature-disabled convention).
package skill

import (
	"context"
	"errors"
	"log/slog"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

// Domain errors surfaced to the REST layer.
var (
	ErrNameRequired          = errors.New("skill: name is required")
	ErrInstructionsRequired  = errors.New("skill: instructions (SKILL.md body) are required")
	ErrInvalidTags           = errors.New("skill: invalid tags")
	ErrPlatformSkillConflict = errors.New(
		"skill: platform skill exists with different content",
	)
)

// SkillPackagerBridge separates local package preparation from publication.
type SkillPackagerBridge interface {
	PrepareCatalogFromDir(
		ctx context.Context,
		dir, repoIdentity string,
	) (*extensionsvc.PreparedSkill, error)
	StorePrepared(
		ctx context.Context,
		prepared *extensionsvc.PreparedSkill,
	) (*extensionsvc.PackagedSkill, error)
	DeletePackage(ctx context.Context, storageKey string) error
}

type Service struct {
	store    skilldom.Repository
	gitops   gitops.Service
	packager SkillPackagerBridge
	logger   *slog.Logger
}

type Deps struct {
	// Store is the authored_skills DB cache (List/Get/Delete row).
	Store skilldom.Repository
	// Gitops is the git-backing choke point (namespace am-skills).
	Gitops gitops.Service
	// Packager bridges to the existing extension packager pipeline.
	Packager SkillPackagerBridge
	Logger   *slog.Logger
}

// NewService returns a git-backed skill authoring service, or nil when any of
// its hard dependencies (gitops, packager, store) is missing — in which case
// the feature is disabled and the REST routes no-op.
func NewService(deps Deps) *Service {
	if deps.Gitops == nil || deps.Packager == nil || deps.Store == nil {
		return nil
	}
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		store:    deps.Store,
		gitops:   deps.Gitops,
		packager: deps.Packager,
		logger:   logger.With("component", "skill"),
	}
}

// cleanupRepo best-effort deletes a freshly-provisioned repo when a later step
// (packaging or the DB insert) fails, mirroring the expert/KB compensating
// cleanup pattern.
func (s *Service) cleanupRepo(ctx context.Context, repoName string) {
	if delErr := s.gitops.DeleteRepo(ctx, repoName); delErr != nil {
		s.logger.Warn("skill: compensating repo delete failed", "repo", repoName, "error", delErr)
	}
}

func branchOf(repo *gitops.Repo) string {
	if repo != nil && repo.DefaultBranch != "" {
		return repo.DefaultBranch
	}
	return "main"
}

func branchOrDefault(branch string) string {
	if branch != "" {
		return branch
	}
	return "main"
}
