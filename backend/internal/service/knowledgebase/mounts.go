package knowledgebase

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
)

type AgentMountInput struct {
	AgentSlug string
	Mode      string
}

func (s *Service) SetAgentMounts(ctx context.Context, orgID, kbID int64, inputs []AgentMountInput) error {
	if _, err := s.repo.Get(ctx, orgID, kbID); err != nil {
		return err
	}
	mounts := make([]*knowledgebase.AgentMount, 0, len(inputs))
	for _, in := range inputs {
		mode := in.Mode
		if mode == "" {
			mode = knowledgebase.MountModeReadOnly
		}
		if !knowledgebase.ValidMountMode(mode) {
			return fmt.Errorf("%w: mount mode must be ro or rw, got %q", ErrInvalidInput, in.Mode)
		}
		mounts = append(mounts, &knowledgebase.AgentMount{AgentSlug: in.AgentSlug, Mode: mode})
	}
	return s.repo.ReplaceAgentMounts(ctx, orgID, kbID, mounts)
}

func (s *Service) ListAgentMounts(ctx context.Context, orgID, kbID int64) ([]*knowledgebase.AgentMount, error) {
	return s.repo.ListAgentMounts(ctx, orgID, kbID)
}

// ResolvedMount pairs a KB with the effective mount mode for one pod.
type ResolvedMount struct {
	KB            *knowledgebase.KnowledgeBase
	Mode          string
	SSHCloneURL   string
	GitKnownHosts string
	GitPrivateKey string
}

// MountRequest is a per-pod KB selection (KB slug, not agent slug).
type MountRequest struct {
	KBSlug string
	Mode   string
}

// ResolveMountsForPod merges the agent's default mounts with per-request
// slug selections. Request selections win on mode conflicts; unknown slugs
// are rejected so a typo doesn't silently drop a mount.
func (s *Service) ResolveMountsForPod(
	ctx context.Context, orgID int64, agentSlug string, requested []MountRequest,
) ([]*ResolvedMount, error) {
	modeBySlug := map[string]string{}

	if agentSlug != "" {
		defaults, err := s.repo.ListMountsForAgent(ctx, orgID, agentSlug)
		if err != nil {
			return nil, err
		}
		if len(defaults) > 0 {
			kbByID := map[int64]*knowledgebase.KnowledgeBase{}
			kbs, err := s.repo.List(ctx, &knowledgebase.ListFilter{OrganizationID: orgID})
			if err != nil {
				return nil, err
			}
			for _, kb := range kbs {
				kbByID[kb.ID] = kb
			}
			for _, m := range defaults {
				if kb, ok := kbByID[m.KnowledgeBaseID]; ok {
					modeBySlug[kb.Slug] = m.Mode
				}
			}
		}
	}

	for _, req := range requested {
		mode := req.Mode
		if mode == "" {
			mode = knowledgebase.MountModeReadOnly
		}
		if !knowledgebase.ValidMountMode(mode) {
			return nil, fmt.Errorf("%w: mount mode must be ro or rw, got %q", ErrInvalidInput, req.Mode)
		}
		modeBySlug[req.KBSlug] = mode
	}

	if len(modeBySlug) == 0 {
		return nil, nil
	}
	slugs := make([]string, 0, len(modeBySlug))
	for slug := range modeBySlug {
		slugs = append(slugs, slug)
	}
	kbs, err := s.repo.ListBySlugs(ctx, orgID, slugs)
	if err != nil {
		return nil, err
	}
	found := map[string]*knowledgebase.KnowledgeBase{}
	for _, kb := range kbs {
		found[kb.Slug] = kb
	}
	resolved := make([]*ResolvedMount, 0, len(slugs))
	for _, slug := range slugs {
		kb, ok := found[slug]
		if !ok {
			return nil, fmt.Errorf("%w: knowledge base %q not found", ErrNotFound, slug)
		}
		if strings.TrimSpace(kb.GitRepoPath) == "" || strings.TrimSpace(kb.HTTPCloneURL) == "" {
			return nil, fmt.Errorf("%w: knowledge base %q repository is unavailable", ErrNotConfigured, slug)
		}
		mode := modeBySlug[slug]
		privateKey, err := s.mountPrivateKey(kb.SourceConfig, mode)
		if err != nil {
			return nil, fmt.Errorf("knowledgebase: resolve mount %q: %w", slug, err)
		}
		sshCloneURL := s.git.SSHCloneURL(repoNameFromPath(kb.GitRepoPath))
		if sshCloneURL == "" {
			return nil, fmt.Errorf("%w: Gitea SSH clone URL is not configured", ErrNotConfigured)
		}
		knownHosts := s.git.SSHKnownHosts()
		if knownHosts == "" {
			return nil, fmt.Errorf("%w: Gitea SSH host key is not configured", ErrNotConfigured)
		}
		resolved = append(resolved, &ResolvedMount{
			KB:            kb,
			Mode:          mode,
			SSHCloneURL:   sshCloneURL,
			GitKnownHosts: knownHosts,
			GitPrivateKey: privateKey,
		})
	}
	return resolved, nil
}
