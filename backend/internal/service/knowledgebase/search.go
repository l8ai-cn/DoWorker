package knowledgebase

import (
	"context"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
)

type SearchMatch struct {
	KBSlug string `json:"kb_slug"`
	Path   string `json:"path"`
	Line   int    `json:"line"`
	Text   string `json:"text"`
}

const (
	searchMaxFilesPerKB = 200
	searchMaxFileBytes  = 512 * 1024
)

// Search greps wiki pages (plus llms.txt / AGENTS.md) across the given KBs —
// or every KB in the org when slugs is empty. Case-insensitive substring
// match; no vector index by design (llm-wiki philosophy: structured wiki +
// grep beats RAG at this scale).
func (s *Service) Search(ctx context.Context, orgID int64, slugs []string, query string, limit int) ([]*SearchMatch, error) {
	if limit <= 0 {
		limit = 20
	}
	kbs, err := s.searchTargets(ctx, orgID, slugs)
	if err != nil {
		return nil, err
	}

	needle := strings.ToLower(query)
	matches := []*SearchMatch{}
	for _, kb := range kbs {
		if len(matches) >= limit {
			break
		}
		kbMatches, err := s.searchOneKB(ctx, kb, needle, limit-len(matches))
		if err != nil {
			s.log.Warn("kb search failed for one kb", "kb_slug", kb.Slug, "error", err)
			continue
		}
		matches = append(matches, kbMatches...)
	}
	return matches, nil
}

func (s *Service) searchTargets(ctx context.Context, orgID int64, slugs []string) ([]*knowledgebase.KnowledgeBase, error) {
	if len(slugs) == 0 {
		return s.repo.List(ctx, &knowledgebase.ListFilter{OrganizationID: orgID})
	}
	return s.repo.ListBySlugs(ctx, orgID, slugs)
}

func (s *Service) searchOneKB(ctx context.Context, kb *knowledgebase.KnowledgeBase, needle string, limit int) ([]*SearchMatch, error) {
	repoName := repoNameFromPath(kb.GitRepoPath)
	tree, err := s.git.ListTree(ctx, repoName, kb.DefaultBranch)
	if err != nil {
		return nil, err
	}

	matches := []*SearchMatch{}
	scanned := 0
	for _, entry := range tree {
		if len(matches) >= limit || scanned >= searchMaxFilesPerKB {
			break
		}
		if entry.Type != "blob" || !searchablePath(entry.Path) || entry.Size > searchMaxFileBytes {
			continue
		}
		scanned++
		file, err := s.git.GetFile(ctx, repoName, kb.DefaultBranch, entry.Path)
		if err != nil {
			continue
		}
		content, err := file.DecodedContent()
		if err != nil {
			continue
		}
		for i, line := range strings.Split(content, "\n") {
			if strings.Contains(strings.ToLower(line), needle) {
				matches = append(matches, &SearchMatch{
					KBSlug: kb.Slug, Path: entry.Path, Line: i + 1, Text: strings.TrimSpace(line),
				})
				if len(matches) >= limit {
					break
				}
			}
		}
	}
	return matches, nil
}

// searchablePath limits grep to the compiled wiki layer + index/schema files.
// raw/ is immutable source material and can be arbitrarily large.
func searchablePath(path string) bool {
	if path == "llms.txt" || path == "AGENTS.md" {
		return true
	}
	return strings.HasPrefix(path, "wiki/") && strings.HasSuffix(path, ".md")
}
