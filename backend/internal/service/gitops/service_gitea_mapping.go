package gitops

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/anthropics/agentsmesh/backend/internal/infra/gitea"
)

func giteaAuthor(author Author) gitea.CommitAuthor {
	return gitea.CommitAuthor{Name: author.Name, Email: author.Email}
}

func toGiteaChanges(changes []FileChange) []gitea.FileChange {
	out := make([]gitea.FileChange, 0, len(changes))
	for _, change := range changes {
		out = append(out, gitea.FileChange{Path: change.Path, Content: string(change.Content)})
	}
	return out
}

func treeType(value string) string {
	switch value {
	case "blob":
		return "file"
	case "tree":
		return "dir"
	default:
		return value
	}
}

func baseName(path string) string {
	return repoNameFromPath(path)
}

func isNotFound(err error) bool {
	var httpError *gitea.HTTPError
	return errors.As(err, &httpError) && httpError.StatusCode == http.StatusNotFound
}

func mapNotFound(err error) error {
	if isNotFound(err) {
		return ErrNotFound
	}
	return fmt.Errorf("gitops: %w", err)
}
