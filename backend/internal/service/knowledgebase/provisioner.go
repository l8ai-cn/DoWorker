package knowledgebase

import (
	"context"
	"embed"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/infra/gitea"
)

//go:embed scaffold/*.tmpl
var scaffoldFS embed.FS

var scaffoldFiles = map[string]string{
	"scaffold/llms.txt.tmpl":      "llms.txt",
	"scaffold/AGENTS.md.tmpl":     "AGENTS.md",
	"scaffold/wiki-index.md.tmpl": "wiki/index.md",
	"scaffold/wiki-log.md.tmpl":   "wiki/log.md",
	"scaffold/raw-readme.md.tmpl": "raw/README.md",
}

var scaffoldTemplates = template.Must(template.ParseFS(scaffoldFS, "scaffold/*.tmpl"))

type scaffoldData struct {
	Name        string
	Description string
	Date        string
}

func renderScaffold(name, description string) ([]gitea.FileChange, error) {
	if description == "" {
		description = "Knowledge base maintained by Do Worker agents."
	}
	data := scaffoldData{Name: name, Description: description, Date: time.Now().UTC().Format("2006-01-02")}
	changes := make([]gitea.FileChange, 0, len(scaffoldFiles))
	for tmplPath, repoPath := range scaffoldFiles {
		var sb strings.Builder
		tmplName := tmplPath[len("scaffold/"):]
		if err := scaffoldTemplates.ExecuteTemplate(&sb, tmplName, data); err != nil {
			return nil, fmt.Errorf("knowledgebase: render %s: %w", tmplName, err)
		}
		changes = append(changes, gitea.FileChange{Path: repoPath, Content: sb.String()})
	}
	return changes, nil
}

// provisionRepo creates the Gitea repo and seeds the llm-wiki scaffold in a
// single initial commit. Returns the repo name inside the KB namespace.
func (s *Service) provisionRepo(ctx context.Context, orgID int64, slug, name, description, branch string) (*gitea.Repo, string, error) {
	if err := s.git.EnsureNamespace(ctx); err != nil {
		return nil, "", fmt.Errorf("knowledgebase: ensure namespace: %w", err)
	}
	// Prefix with org ID: KB slugs are unique per Do Worker org, but all
	// repos share one Gitea namespace.
	repoName := fmt.Sprintf("org%d-%s", orgID, slug)
	repo, err := s.git.CreateRepo(ctx, repoName, branch)
	if err != nil {
		return nil, "", fmt.Errorf("knowledgebase: create repo: %w", err)
	}
	changes, err := renderScaffold(name, description)
	if err == nil {
		err = s.git.CommitFiles(ctx, repoName, branch,
			"init: knowledge base scaffold (llms.txt, AGENTS.md, raw/, wiki/)",
			gitea.CommitAuthor{Name: "Do Worker", Email: "kb@agentsmesh.local"},
			changes, nil)
	}
	if err != nil {
		_ = s.git.DeleteRepo(ctx, repoName)
		return nil, "", fmt.Errorf("knowledgebase: seed scaffold: %w", err)
	}
	return repo, repoName, nil
}

func (s *Service) provisionMountDeployKeys(
	ctx context.Context,
	repoName string,
) (*mountDeployKeys, error) {
	keys, err := newMountDeployKeys()
	if err != nil {
		return nil, err
	}
	if _, err := s.git.CreateDeployKey(
		ctx,
		repoName,
		"agentsmesh-read-only",
		keys.readOnlyPublic,
		true,
	); err != nil {
		return nil, fmt.Errorf("knowledgebase: create read-only deploy key: %w", err)
	}
	if _, err := s.git.CreateDeployKey(
		ctx,
		repoName,
		"agentsmesh-read-write",
		keys.readWritePublic,
		false,
	); err != nil {
		return nil, fmt.Errorf("knowledgebase: create read-write deploy key: %w", err)
	}
	return keys, nil
}
