package gitea

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type repositoryVisibility struct {
	Private bool `json:"private"`
}

type branchCommit struct {
	Commit struct {
		ID string `json:"id"`
	} `json:"commit"`
}

func (c *Client) ResolvePublicBranchCommit(
	ctx context.Context,
	repositoryPath, branch string,
) (string, error) {
	if !validRepositoryPath(repositoryPath) || strings.TrimSpace(branch) == "" {
		return "", fmt.Errorf("gitea repository path and branch are required")
	}
	var visibility repositoryVisibility
	if err := c.do(ctx, http.MethodGet, "/repos/"+escapePath(repositoryPath), nil, &visibility); err != nil {
		return "", err
	}
	if visibility.Private {
		return "", fmt.Errorf("gitea repository %q is not public", repositoryPath)
	}
	return c.ResolveBranchCommit(ctx, repositoryPath, branch)
}

func (c *Client) ResolveBranchCommit(
	ctx context.Context,
	repositoryPath, branch string,
) (string, error) {
	if !validRepositoryPath(repositoryPath) || strings.TrimSpace(branch) == "" {
		return "", fmt.Errorf("gitea repository path and branch are required")
	}
	var response branchCommit
	path := "/repos/" + escapePath(repositoryPath) + "/branches/" + url.PathEscape(branch)
	if err := c.do(ctx, http.MethodGet, path, nil, &response); err != nil {
		return "", err
	}
	return strings.TrimSpace(response.Commit.ID), nil
}

func validRepositoryPath(value string) bool {
	parts := strings.Split(value, "/")
	return len(parts) == 2 && strings.TrimSpace(parts[0]) != "" &&
		strings.TrimSpace(parts[1]) != ""
}
