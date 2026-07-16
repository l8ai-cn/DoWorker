package skill

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

// ErrImportURLRequired is returned when ImportFromGit gets an empty URL.
var ErrImportURLRequired = errors.New("skill: import url is required")

// ImportFromGitRequest imports skills from an external git repo. Single-skill
// repos and collections (anthropics/skills-style) are auto-detected; every
// discovered skill fans out into its own internal am-skills repo + catalog row.
type ImportFromGitRequest struct {
	OrganizationID int64
	UserID         int64
	URL            string
	Branch         string
	Subdir         string   // optional: import only this directory
	AgentFilter    []string // optional agent whitelist; empty = all agents
	AuthType       string   // optional: extension.AuthType* for private repos
	AuthCredential string
}

// ImportFromGit clones the source, discovers skills, and ingests each one via
// importSkillDir. Per-skill failures don't abort the batch; they are joined
// into the returned error alongside the successfully imported rows.
func (s *Service) ImportFromGit(ctx context.Context, req *ImportFromGitRequest) ([]*skilldom.Skill, error) {
	url := strings.TrimSpace(req.URL)
	if url == "" {
		return nil, ErrImportURLRequired
	}

	var auth *extensionsvc.SkillSourceAuth
	if req.AuthType != "" && req.AuthCredential != "" {
		auth = &extensionsvc.SkillSourceAuth{Type: req.AuthType, Credential: req.AuthCredential}
	}
	src, err := extensionsvc.CloneSkillSource(ctx, url, req.Branch, auth)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	infos, err := extensionsvc.ScanSkillSource(src.Dir, strings.TrimSpace(req.Subdir))
	if err != nil {
		return nil, err
	}
	if len(infos) == 0 {
		return nil, fmt.Errorf("skill: no SKILL.md found in %s", url)
	}

	var (
		rows       []*skilldom.Skill
		importErrs []error
	)
	for _, info := range infos {
		row, err := s.importSkillDir(ctx, req, src, info)
		if err != nil {
			importErrs = append(importErrs, fmt.Errorf("%s: %w", info.Slug, err))
			continue
		}
		rows = append(rows, row)
	}
	return rows, errors.Join(importErrs...)
}

// importSkillDir mirrors one discovered skill directory into its internal
// am-skills repo, packages it, and upserts the catalog row. Re-importing the
// same upstream(url+subdir) updates the existing row in place.
func (s *Service) importSkillDir(
	ctx context.Context, req *ImportFromGitRequest,
	src *extensionsvc.ClonedSkillSource, info extensionsvc.SkillInfo,
) (*skilldom.Skill, error) {
	subdir := extensionsvc.SkillSourceSubdir(src.Dir, info)
	files, err := readSkillDirFiles(info.DirPath)
	if err != nil {
		return nil, err
	}

	existing, err := s.store.FindByUpstream(ctx, req.OrganizationID, req.URL, subdir)
	switch {
	case err == nil:
		return s.refreshImportedSkill(ctx, existing, src, info, files)
	case errors.Is(err, skilldom.ErrNotFound):
		return s.createImportedSkill(ctx, req, src, info, subdir, files)
	default:
		return nil, err
	}
}

const (
	maxImportFiles      = 500
	maxImportTotalBytes = 50 * 1024 * 1024
)

// readSkillDirFiles loads a skill directory as gitops file changes (paths
// relative to the skill root), skipping VCS metadata and enforcing the same
// size envelope as the upload path.
func readSkillDirFiles(dir string) ([]gitops.FileChange, error) {
	var (
		files []gitops.FileChange
		total int64
	)
	root := filepath.Clean(dir)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		total += int64(len(data))
		if len(files) >= maxImportFiles || total > maxImportTotalBytes {
			return fmt.Errorf("skill: source exceeds import limits (%d files / %d bytes)", maxImportFiles, maxImportTotalBytes)
		}
		files = append(files, gitops.FileChange{Path: filepath.ToSlash(rel), Content: data})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("skill: directory %q has no files", dir)
	}
	return files, nil
}

func marshalAgentFilter(filter []string) []byte {
	if len(filter) == 0 {
		return []byte("[]")
	}
	out := make([]string, 0, len(filter))
	for _, f := range filter {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	b, err := jsonMarshal(out)
	if err != nil {
		return []byte("[]")
	}
	return b
}

func displayNameOr(name, fallback string) string {
	if n := strings.TrimSpace(name); n != "" {
		return n
	}
	return fallback
}

func shortSha(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	if sha == "" {
		return "unknown"
	}
	return sha
}
