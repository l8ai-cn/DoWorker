package gitea

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
)

type TreeChange struct {
	Path    string
	Content []byte
	SHA     string
	Delete  bool
}

func (c *Client) CommitTreeChanges(
	ctx context.Context,
	repo, branch, message string,
	author CommitAuthor,
	changes []TreeChange,
) error {
	files := make([]map[string]any, 0, len(changes))
	for _, change := range changes {
		file := map[string]any{
			"path": change.Path,
		}
		switch {
		case change.Delete:
			file["operation"] = "delete"
			file["sha"] = change.SHA
		case change.SHA != "":
			file["operation"] = "update"
			file["sha"] = change.SHA
			file["content"] = base64.StdEncoding.EncodeToString(change.Content)
		default:
			file["operation"] = "create"
			file["content"] = base64.StdEncoding.EncodeToString(change.Content)
		}
		files = append(files, file)
	}
	body := map[string]any{
		"branch":  branch,
		"message": message,
		"files":   files,
	}
	if author.Name != "" {
		identity := map[string]string{"name": author.Name, "email": author.Email}
		body["author"] = identity
		body["committer"] = identity
	}
	return c.do(
		ctx,
		http.MethodPost,
		fmt.Sprintf("/repos/%s/%s/contents", c.cfg.Namespace, repo),
		body,
		nil,
	)
}
