package runner

import (
	"io/fs"
	"path/filepath"
	"strings"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

const maxSearchResults = 500

func (h *RunnerMessageHandler) sandboxFsSearch(workspaceRoot, query, include, exclude string) (*runnerv1.SandboxFsResultEvent, error) {
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.sandboxFsSearchWorkspace(workspace, query, include, exclude)
}

func (h *RunnerMessageHandler) sandboxFsSearchWorkspace(
	workspace *sandboxWorkspace,
	query, include, exclude string,
) (*runnerv1.SandboxFsResultEvent, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	includes := splitGlobs(include)
	excludes := splitGlobs(exclude)
	var entries []*runnerv1.SandboxFsEntry
	_ = fs.WalkDir(workspace.root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if len(entries) >= maxSearchResults {
			return filepath.SkipAll
		}
		rel := filepath.ToSlash(path)
		if q != "" && !strings.Contains(strings.ToLower(rel), q) && !strings.Contains(strings.ToLower(d.Name()), q) {
			return nil
		}
		if !matchGlobs(rel, includes, excludes) {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return nil
		}
		entries = append(entries, fsEntryFromInfo(filepath.Dir(rel), fi))
		return nil
	})
	return &runnerv1.SandboxFsResultEvent{
		Entries:       entries,
		WorkspaceRoot: workspace.displayPath(),
	}, nil
}

func splitGlobs(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func matchGlobs(path string, includes, excludes []string) bool {
	for _, g := range excludes {
		if globMatch(g, path) {
			return false
		}
	}
	if len(includes) == 0 {
		return true
	}
	for _, g := range includes {
		if globMatch(g, path) {
			return true
		}
	}
	return false
}

func globMatch(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)
	if ok, _ := filepath.Match(pattern, path); ok {
		return true
	}
	if strings.Contains(pattern, "**") {
		prefix := strings.TrimSuffix(pattern, "**")
		prefix = strings.TrimSuffix(prefix, "/")
		return strings.HasPrefix(path, prefix)
	}
	base := filepath.Base(path)
	ok, _ := filepath.Match(pattern, base)
	return ok
}
