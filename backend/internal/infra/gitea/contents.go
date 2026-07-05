package gitea

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
)

type FileChange struct {
	Path    string
	Content string
}

type CommitAuthor struct {
	Name  string
	Email string
}

// CommitFiles creates or updates multiple files in a single commit via the
// Gitea batch contents API. Works on empty repositories (creates the initial
// commit on the default branch).
func (c *Client) CommitFiles(
	ctx context.Context, repo, branch, message string,
	author CommitAuthor, changes []FileChange, isUpdate map[string]string,
) error {
	files := make([]map[string]any, 0, len(changes))
	for _, ch := range changes {
		op := map[string]any{
			"operation": "create",
			"path":      ch.Path,
			"content":   base64.StdEncoding.EncodeToString([]byte(ch.Content)),
		}
		if sha, ok := isUpdate[ch.Path]; ok {
			op["operation"] = "update"
			op["sha"] = sha
		}
		files = append(files, op)
	}
	body := map[string]any{
		"branch":  branch,
		"message": message,
		"files":   files,
	}
	if author.Name != "" {
		body["author"] = map[string]string{"name": author.Name, "email": author.Email}
		body["committer"] = map[string]string{"name": author.Name, "email": author.Email}
	}
	return c.do(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/contents", c.cfg.Namespace, repo), body, nil)
}

type ContentEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	SHA     string `json:"sha"`
	Type    string `json:"type"` // file | dir
	Size    int64  `json:"size"`
	Content string `json:"content"` // base64, only for file GETs
}

func (e *ContentEntry) DecodedContent() (string, error) {
	data, err := base64.StdEncoding.DecodeString(e.Content)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetFile fetches a single file (content base64-populated).
func (c *Client) GetFile(ctx context.Context, repo, branch, path string) (*ContentEntry, error) {
	var entry ContentEntry
	url := fmt.Sprintf("/repos/%s/%s/contents/%s?ref=%s", c.cfg.Namespace, repo, escapePath(path), branch)
	if err := c.do(ctx, http.MethodGet, url, nil, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// ListDir lists directory entries (content field empty).
func (c *Client) ListDir(ctx context.Context, repo, branch, path string) ([]*ContentEntry, error) {
	var entries []*ContentEntry
	url := fmt.Sprintf("/repos/%s/%s/contents/%s?ref=%s", c.cfg.Namespace, repo, escapePath(path), branch)
	if err := c.do(ctx, http.MethodGet, url, nil, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
