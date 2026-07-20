package workspace

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func (m *Manager) prepareAuthURL(repoURL string, _ *WorktreeOptions) string {
	return repoURL
}

// TempWorkspace creates a temporary workspace directory
func (m *Manager) TempWorkspace(podKey string) string {
	path := filepath.Join(m.root, "temp", podKey)
	os.MkdirAll(path, 0755)
	return path
}

// GetWorkspaceRoot returns the workspace root directory
func (m *Manager) GetWorkspaceRoot() string {
	return m.root
}

// ListWorktrees lists all active worktrees.
// Worktrees are located at sandboxes/{podKey}/workspace
func (m *Manager) ListWorktrees() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sandboxesDir := filepath.Join(m.root, "sandboxes")
	entries, err := os.ReadDir(sandboxesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var worktrees []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		worktreePath := filepath.Join(sandboxesDir, entry.Name(), "workspace")
		// Only include if worktree actually exists
		if _, err := os.Stat(worktreePath); err == nil {
			worktrees = append(worktrees, worktreePath)
		}
	}

	return worktrees, nil
}

// extractRepoName extracts repository name from URL.
// Supports: SCP-style (git@host:user/repo), ssh:// (with optional port), and HTTPS URLs.
func extractRepoName(repoURL string) string {
	// Handle ssh:// URLs: ssh://git@host:port/user/repo.git
	if strings.HasPrefix(repoURL, "ssh://") {
		if u, err := url.Parse(repoURL); err == nil {
			p := strings.TrimPrefix(u.Path, "/")
			p = strings.TrimSuffix(p, ".git")
			if p != "" {
				return strings.ReplaceAll(p, "/", "-")
			}
		}
	}

	// Handle SCP-style SSH URLs: git@github.com:user/repo.git
	if strings.Contains(repoURL, "@") && !strings.Contains(repoURL, "://") {
		if idx := strings.LastIndex(repoURL, ":"); idx != -1 {
			path := repoURL[idx+1:]
			path = strings.TrimSuffix(path, ".git")
			return strings.ReplaceAll(path, "/", "-")
		}
	}

	normalizedPath := strings.ReplaceAll(repoURL, `\`, "/")
	parts := strings.Split(strings.TrimSuffix(normalizedPath, "/"), "/")
	if len(parts) >= 2 {
		name := parts[len(parts)-1]
		name = strings.TrimSuffix(name, ".git")
		owner := parts[len(parts)-2]
		return fmt.Sprintf("%s-%s", owner, name)
	}

	return ""
}
