package gitea

import (
	"context"
	"fmt"
	"net/http"
)

type TreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"` // blob | tree
	Size int64  `json:"size"`
	SHA  string `json:"sha"`
}

// ListTree enumerates the whole repo tree in one call (recursive git tree),
// which is how kb_search discovers wiki pages without N directory listings.
func (c *Client) ListTree(ctx context.Context, repo, ref string) ([]TreeEntry, error) {
	var out struct {
		Tree      []TreeEntry `json:"tree"`
		Truncated bool        `json:"truncated"`
	}
	path := fmt.Sprintf("/repos/%s/%s/git/trees/%s?recursive=true&per_page=1000",
		c.cfg.Namespace, escapePath(repo), escapePath(ref))
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Tree, nil
}
