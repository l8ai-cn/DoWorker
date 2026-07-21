package gitops

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/infra/gitea"
)

type TreeSnapshot struct {
	Files map[string][]byte
}

func CaptureTree(
	ctx context.Context,
	gitService Service,
	repoName, ref string,
) (*TreeSnapshot, error) {
	entries, err := gitService.ListTree(ctx, repoName, ref)
	if err != nil {
		return nil, fmt.Errorf("gitops: capture tree: %w", err)
	}
	files := make(map[string][]byte, len(entries))
	for _, entry := range entries {
		if entry.Type != "file" {
			continue
		}
		content, _, err := gitService.ReadFile(ctx, repoName, ref, entry.Path)
		if err != nil {
			return nil, fmt.Errorf("gitops: capture %s: %w", entry.Path, err)
		}
		files[entry.Path] = content
	}
	return &TreeSnapshot{Files: files}, nil
}

func RestoreTree(
	ctx context.Context,
	gitService Service,
	repoName, branch string,
	snapshot *TreeSnapshot,
) error {
	switch typed := gitService.(type) {
	case *service:
		return typed.restoreTree(ctx, repoName, branch, snapshot)
	case *Fake:
		return typed.restoreTree(repoName, snapshot)
	default:
		return errors.New("gitops: service does not support tree restoration")
	}
}

func (s *service) restoreTree(
	ctx context.Context,
	repoName, branch string,
	snapshot *TreeSnapshot,
) error {
	current, err := captureTreeEntries(ctx, s, repoName, branch)
	if err != nil {
		return err
	}
	changes := make([]gitea.TreeChange, 0, len(current)+len(snapshot.Files))
	for path, content := range snapshot.Files {
		entry, exists := current[path]
		if exists && bytes.Equal(entry.content, content) {
			delete(current, path)
			continue
		}
		change := gitea.TreeChange{Path: path, Content: content}
		if exists {
			change.SHA = entry.sha
			delete(current, path)
		}
		changes = append(changes, change)
	}
	for path, entry := range current {
		changes = append(changes, gitea.TreeChange{
			Path: path, SHA: entry.sha, Delete: true,
		})
	}
	if len(changes) == 0 {
		return nil
	}
	return s.git.CommitTreeChanges(
		ctx, repoName, branch, "restore: failed skill mutation",
		giteaAuthor(defaultAuthor), changes,
	)
}

type capturedTreeEntry struct {
	content []byte
	sha     string
}

func captureTreeEntries(
	ctx context.Context,
	gitService Service,
	repoName, ref string,
) (map[string]capturedTreeEntry, error) {
	entries, err := gitService.ListTree(ctx, repoName, ref)
	if err != nil {
		return nil, fmt.Errorf("gitops: inspect tree for restore: %w", err)
	}
	files := make(map[string]capturedTreeEntry, len(entries))
	for _, entry := range entries {
		if entry.Type != "file" {
			continue
		}
		content, file, err := gitService.ReadFile(ctx, repoName, ref, entry.Path)
		if err != nil {
			return nil, fmt.Errorf("gitops: inspect %s for restore: %w", entry.Path, err)
		}
		files[entry.Path] = capturedTreeEntry{content: content, sha: file.SHA}
	}
	return files, nil
}

func (f *Fake) restoreTree(repoName string, snapshot *TreeSnapshot) error {
	repo, ok := f.Repos[repoName]
	if !ok {
		return ErrNotFound
	}
	repo.Files = make(map[string][]byte, len(snapshot.Files))
	repo.SHAs = make(map[string]string, len(snapshot.Files))
	for path, content := range snapshot.Files {
		repo.put(path, content)
	}
	return nil
}
