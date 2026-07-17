package gitea

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type DeployKey struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Key      string `json:"key"`
	ReadOnly bool   `json:"read_only"`
}

func (c *Client) CreateDeployKey(
	ctx context.Context,
	repo, title, publicKey string,
	readOnly bool,
) (*DeployKey, error) {
	var key DeployKey
	err := c.do(
		ctx,
		http.MethodPost,
		fmt.Sprintf("/repos/%s/%s/keys", c.cfg.Namespace, repo),
		map[string]any{
			"title":     title,
			"key":       strings.TrimSpace(publicKey),
			"read_only": readOnly,
		},
		&key,
	)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (c *Client) ListDeployKeys(ctx context.Context, repo string) ([]DeployKey, error) {
	var keys []DeployKey
	err := c.do(
		ctx,
		http.MethodGet,
		fmt.Sprintf("/repos/%s/%s/keys", c.cfg.Namespace, repo),
		nil,
		&keys,
	)
	if IsHTTPStatus(err, http.StatusNotFound) {
		return nil, nil
	}
	return keys, err
}

func (c *Client) DeleteDeployKey(ctx context.Context, repo string, id int64) error {
	err := c.do(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("/repos/%s/%s/keys/%d", c.cfg.Namespace, repo, id),
		nil,
		nil,
	)
	if IsHTTPStatus(err, http.StatusNotFound) {
		return nil
	}
	return err
}
