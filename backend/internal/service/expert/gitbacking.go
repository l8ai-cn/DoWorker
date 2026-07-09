package expert

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

// ErrGitBackingDisabled is returned by the repo-content read methods when the
// expert has no backing repo (gitops disabled or a legacy row not yet
// provisioned).
var ErrGitBackingDisabled = errors.New("expert: git backing not available")

// ErrFileNotFound mirrors gitops.ErrNotFound so REST callers can map content
// misses to 404 without importing gitops.
var ErrFileNotFound = gitops.ErrNotFound

// expertConfig is the structured expert.json committed alongside agent.md. Git
// (agent.md + expert.json) is the source of truth; the DB columns are a cache.
type expertConfig struct {
	Schema          int                        `json:"schema"`
	Name            string                     `json:"name"`
	Description     string                     `json:"description,omitempty"`
	Avatar          string                     `json:"avatar,omitempty"`     // 形象 (repo-relative path)
	ExpertType      string                     `json:"expertType,omitempty"` // 类型
	AgentSlug       string                     `json:"agentSlug"`
	InteractionMode string                     `json:"interactionMode"`
	AutomationLevel string                     `json:"automationLevel,omitempty"`
	Perpetual       bool                       `json:"perpetual"`
	SkillSlugs      []string                   `json:"skillSlugs,omitempty"`
	KnowledgeMounts []expertdom.KnowledgeMount `json:"knowledgeMounts,omitempty"`
	UsedEnvBundles  []string                   `json:"usedEnvBundles,omitempty"`
	ConfigOverrides map[string]any             `json:"configOverrides,omitempty"`
	Repository      *expertConfigRepository    `json:"repository,omitempty"`
}

type expertConfigRepository struct {
	RepositoryID *int64 `json:"repositoryId,omitempty"`
	Branch       string `json:"branch,omitempty"`
}

// AvatarInput is a validated avatar upload forwarded from the REST handler. The
// platform controls the extension (derived from a magic-byte sniff), so the
// repo-relative path is always assets/avatar.<ext>.
type AvatarInput struct {
	Data []byte
	Ext  string // "png" | "jpg" | "webp" | "gif"
}

func (a *AvatarInput) repoPath() string {
	ext := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(a.Ext)), ".")
	if ext == "" {
		ext = "png"
	}
	return "assets/avatar." + ext
}

const expertConfigSchema = 1

// buildExpertConfig renders the expert.json view of a row. avatarPath overrides
// any cached path (used when a fresh avatar is committed in the same operation).
func (s *Service) buildExpertConfig(e *expertdom.Expert, avatarPath string) expertConfig {
	cfg := expertConfig{
		Schema:          expertConfigSchema,
		Name:            e.Name,
		AgentSlug:       e.AgentSlug,
		InteractionMode: e.InteractionMode,
		AutomationLevel: e.AutomationLevel,
		Perpetual:       e.Perpetual,
		SkillSlugs:      []string(e.SkillSlugs),
		UsedEnvBundles:  []string(e.UsedEnvBundles),
		KnowledgeMounts: expertdom.ParseKnowledgeMounts(e.KnowledgeMounts),
	}
	if e.Description != nil {
		cfg.Description = *e.Description
	}
	meta := parseExpertMetadata(e.Metadata)
	cfg.ExpertType = meta.ExpertType
	cfg.Avatar = meta.Avatar
	if avatarPath != "" {
		cfg.Avatar = avatarPath
	}
	if len(e.ConfigOverrides) > 0 {
		var overrides map[string]any
		if err := json.Unmarshal(e.ConfigOverrides, &overrides); err == nil && len(overrides) > 0 {
			cfg.ConfigOverrides = overrides
		}
	}
	if e.RepositoryID != nil {
		branch := ""
		if e.BranchName != nil {
			branch = *e.BranchName
		}
		cfg.Repository = &expertConfigRepository{RepositoryID: e.RepositoryID, Branch: branch}
	}
	return cfg
}

// renderExpertFiles builds the repo file set. agent.md is the AgentFile layer
// source, expert.json the structured config, plus README.md (seed only) and
// assets/avatar.<ext> when a fresh avatar is supplied.
func (s *Service) renderExpertFiles(
	e *expertdom.Expert, layer string, avatar *AvatarInput, includeReadme bool,
) ([]gitops.FileChange, error) {
	avatarPath := ""
	if avatar != nil && len(avatar.Data) > 0 {
		avatarPath = avatar.repoPath()
	}
	cfg := s.buildExpertConfig(e, avatarPath)
	cfgJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("expert: render expert.json: %w", err)
	}
	changes := []gitops.FileChange{
		{Path: "agent.md", Content: []byte(layer)},
		{Path: "expert.json", Content: append(cfgJSON, '\n')},
	}
	if includeReadme {
		changes = append(changes, gitops.FileChange{
			Path:    "README.md",
			Content: []byte(renderExpertReadme(e)),
		})
	}
	if avatarPath != "" {
		changes = append(changes, gitops.FileChange{Path: avatarPath, Content: avatar.Data})
	}
	return changes, nil
}

func renderExpertReadme(e *expertdom.Expert) string {
	desc := ""
	if e.Description != nil {
		desc = strings.TrimSpace(*e.Description)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n\n", e.Name)
	if desc != "" {
		fmt.Fprintf(&sb, "%s\n\n", desc)
	}
	sb.WriteString("This repository is managed by Do Worker. `agent.md` is the AgentFile ")
	sb.WriteString("layer source and `expert.json` is the structured expert configuration ")
	sb.WriteString("(both are the source of truth for this expert).\n")
	return sb.String()
}

// provisionExpertRepo creates the expert repo and seeds it. On success the
// caller stores the returned repo descriptor on the DB row (cache).
func (s *Service) provisionExpertRepo(
	ctx context.Context, e *expertdom.Expert, layer string, avatar *AvatarInput,
) (*gitops.Repo, error) {
	seed, err := s.renderExpertFiles(e, layer, avatar, true)
	if err != nil {
		return nil, err
	}
	return s.gitops.Provision(ctx, gitops.ProvisionParams{
		OrgID:         e.OrganizationID,
		Slug:          e.Slug,
		CommitMessage: "init: expert scaffold (agent.md, expert.json, README.md)",
		Seed:          seed,
	})
}

// applyRepo copies a provisioned repo descriptor onto the DB row cache columns.
func applyRepo(e *expertdom.Expert, repo *gitops.Repo) {
	path := repo.Path
	e.GitRepoPath = &path
	e.DefaultBranch = repo.DefaultBranch
	if e.DefaultBranch == "" {
		e.DefaultBranch = "main"
	}
	url := repo.HTTPCloneURL
	e.HTTPCloneURL = &url
}

// ensureExpertRepo lazily provisions a repo for a legacy row (GitRepoPath nil)
// on its next update/run. No-op when gitops is disabled or the row is already
// backed. Returns whether a repo was newly provisioned.
func (s *Service) ensureExpertRepo(
	ctx context.Context, e *expertdom.Expert, layer string, avatar *AvatarInput,
) (bool, error) {
	if s.gitops == nil || e.GitRepoPath != nil {
		return false, nil
	}
	repo, err := s.provisionExpertRepo(ctx, e, layer, avatar)
	if err != nil {
		return false, err
	}
	applyRepo(e, repo)
	return true, nil
}

// commitExpertChanges re-renders agent.md/expert.json (+ avatar) and commits
// them to Git. Git is the source of truth, so this runs before the DB cache
// update.
func (s *Service) commitExpertChanges(
	ctx context.Context, e *expertdom.Expert, layer string, avatar *AvatarInput,
) error {
	if s.gitops == nil || e.GitRepoPath == nil {
		return nil
	}
	changes, err := s.renderExpertFiles(e, layer, avatar, false)
	if err != nil {
		return err
	}
	repoName := s.gitops.RepoNameFromPath(*e.GitRepoPath)
	return s.gitops.Commit(ctx, repoName, s.branchOf(e), "update: expert configuration", gitops.Author{}, changes)
}

// readAgentFileFromGit reads agent.md from the expert's repo. Returns
// (content, true) on success; (\"\", false) on miss/disabled/error so callers
// fall back to the DB cache.
func (s *Service) readAgentFileFromGit(ctx context.Context, e *expertdom.Expert) (string, bool) {
	if s.gitops == nil || e.GitRepoPath == nil {
		return "", false
	}
	repoName := s.gitops.RepoNameFromPath(*e.GitRepoPath)
	data, _, err := s.gitops.ReadFile(ctx, repoName, s.branchOf(e), "agent.md")
	if err != nil {
		return "", false
	}
	return string(data), true
}

// refreshAgentfileCache reconciles the DB agentfile_layer cache with the
// Git-sourced layer, best-effort (a failure must not fail the run).
func (s *Service) refreshAgentfileCache(ctx context.Context, e *expertdom.Expert, layer string) {
	current := ""
	if e.AgentfileLayer != nil {
		current = *e.AgentfileLayer
	}
	if strings.TrimSpace(current) == strings.TrimSpace(layer) {
		return
	}
	cache := layer
	e.AgentfileLayer = &cache
	if err := s.store.Update(ctx, e); err != nil {
		s.logger.Warn("expert: agentfile cache refresh failed", "expert_id", e.ID, "error", err)
	}
}

// GitEnabled reports whether git-backing is configured for this service.
func (s *Service) GitEnabled() bool { return s.gitops != nil }

// ReadExpertFile returns the decoded content + entry of a file in the expert's
// repo. relPath must already be sanitized by the caller. Returns
// ErrGitBackingDisabled when the expert has no repo, ErrFileNotFound on miss.
func (s *Service) ReadExpertFile(
	ctx context.Context, orgID int64, slug, relPath string,
) ([]byte, *gitops.Entry, error) {
	e, err := s.store.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, nil, err
	}
	if s.gitops == nil || e.GitRepoPath == nil {
		return nil, nil, ErrGitBackingDisabled
	}
	repoName := s.gitops.RepoNameFromPath(*e.GitRepoPath)
	return s.gitops.ReadFile(ctx, repoName, s.branchOf(e), relPath)
}

// ListExpertTree enumerates the expert repo tree. Returns
// ErrGitBackingDisabled when the expert has no repo.
func (s *Service) ListExpertTree(
	ctx context.Context, orgID int64, slug string,
) ([]gitops.Entry, error) {
	e, err := s.store.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if s.gitops == nil || e.GitRepoPath == nil {
		return nil, ErrGitBackingDisabled
	}
	repoName := s.gitops.RepoNameFromPath(*e.GitRepoPath)
	return s.gitops.ListTree(ctx, repoName, s.branchOf(e))
}

func (s *Service) branchOf(e *expertdom.Expert) string {
	if strings.TrimSpace(e.DefaultBranch) != "" {
		return e.DefaultBranch
	}
	return "main"
}

// --- metadata helpers ---

type expertMetadata struct {
	Avatar     string `json:"avatar,omitempty"`
	ExpertType string `json:"expertType,omitempty"`
}

func parseExpertMetadata(raw json.RawMessage) expertMetadata {
	var m expertMetadata
	if len(raw) == 0 || string(raw) == "null" {
		return m
	}
	_ = json.Unmarshal(raw, &m)
	return m
}

// mergeMetadata sets avatar/expertType on the row's metadata jsonb while
// preserving any other keys. avatarPath/expertType are applied only when
// non-nil. Always returns valid JSON (never nil), so the NOT NULL column holds
// at least "{}".
func mergeMetadata(raw json.RawMessage, avatarPath, expertType *string) json.RawMessage {
	obj := map[string]any{}
	if len(raw) > 0 && string(raw) != "null" {
		_ = json.Unmarshal(raw, &obj)
	}
	if avatarPath != nil {
		if *avatarPath == "" {
			delete(obj, "avatar")
		} else {
			obj["avatar"] = *avatarPath
		}
	}
	if expertType != nil {
		if strings.TrimSpace(*expertType) == "" {
			delete(obj, "expertType")
		} else {
			obj["expertType"] = strings.TrimSpace(*expertType)
		}
	}
	b, err := json.Marshal(obj)
	if err != nil || len(b) == 0 {
		return json.RawMessage("{}")
	}
	return b
}
