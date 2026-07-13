package skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

// materializeRepo writes an am-skills repo tree into a fresh temp dir so the
// filesystem-based extension packager can consume it. It returns the temp dir
// path and a cleanup func (deferred os.RemoveAll); callers MUST call cleanup.
//
// Each tree path is sanitized to stay within the temp dir (defense against a
// hostile/garbled repo tree escaping via "..").
func materializeRepo(
	ctx context.Context, g gitops.Service, repoName, ref string,
) (string, func(), error) {
	entries, err := g.ListTree(ctx, repoName, ref)
	if err != nil {
		return "", nil, fmt.Errorf("skill: list repo tree: %w", err)
	}

	dir, err := os.MkdirTemp("", "skill-gitops-*")
	if err != nil {
		return "", nil, fmt.Errorf("skill: create temp dir: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	root := filepath.Clean(dir)

	for _, e := range entries {
		if e.Type != "file" {
			continue
		}
		rel := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(e.Path, "/")))
		if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
			cleanup()
			return "", nil, fmt.Errorf("skill: unsafe repo path %q", e.Path)
		}
		target := filepath.Join(root, rel)
		if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
			cleanup()
			return "", nil, fmt.Errorf("skill: repo path escapes temp dir %q", e.Path)
		}

		data, _, err := g.ReadFile(ctx, repoName, ref, e.Path)
		if err != nil {
			cleanup()
			return "", nil, fmt.Errorf("skill: read %q: %w", e.Path, err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("skill: mkdir for %q: %w", e.Path, err)
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("skill: write %q: %w", e.Path, err)
		}
	}

	return dir, cleanup, nil
}

func materializeFileChanges(files []gitops.FileChange) (string, func(), error) {
	dir, err := os.MkdirTemp("", "skill-import-*")
	if err != nil {
		return "", nil, fmt.Errorf("skill: create import temp dir: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	root := filepath.Clean(dir)

	for _, file := range files {
		rel := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(file.Path, "/")))
		if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
			cleanup()
			return "", nil, fmt.Errorf("skill: unsafe import path %q", file.Path)
		}
		target := filepath.Join(root, rel)
		if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
			cleanup()
			return "", nil, fmt.Errorf("skill: import path escapes temp dir %q", file.Path)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("skill: mkdir for %q: %w", file.Path, err)
		}
		if err := os.WriteFile(target, file.Content, 0o644); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("skill: write %q: %w", file.Path, err)
		}
	}
	return dir, cleanup, nil
}
