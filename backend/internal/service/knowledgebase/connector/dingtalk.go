package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// DingTalkConnector syncs a DingTalk knowledge workspace (钉钉知识库).
// Node listing walks the workspace tree; doc content is fetched via the
// docs export API in markdown when available.
//
// source_config: {"app_key": "...", "app_secret": "...", "workspace_id": "...", "operator_id": "..."}
// operator_id is the union id of the user the app acts on behalf of —
// DingTalk wiki APIs are operator-scoped.
type DingTalkConnector struct {
	HTTP    *http.Client
	BaseURL string // test override; empty = https://api.dingtalk.com
}

type dingtalkConfig struct {
	AppKey      string `json:"app_key"`
	AppSecret   string `json:"app_secret"`
	WorkspaceID string `json:"workspace_id"`
	OperatorID  string `json:"operator_id"`
}

func (c *DingTalkConnector) SourceType() string { return "dingtalk" }

func (c *DingTalkConnector) base() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return "https://api.dingtalk.com"
}

func (c *DingTalkConnector) ListDocs(ctx context.Context, config json.RawMessage) ([]DocRef, error) {
	cfg, token, err := c.auth(ctx, config)
	if err != nil {
		return nil, err
	}
	return c.listNodes(ctx, token, cfg, cfg.WorkspaceID)
}

func (c *DingTalkConnector) FetchDoc(ctx context.Context, config json.RawMessage, ref DocRef) (*Doc, error) {
	cfg, token, err := c.auth(ctx, config)
	if err != nil {
		return nil, err
	}
	q := url.Values{"operatorId": {cfg.OperatorID}, "targetFormat": {"markdown"}}
	var out struct {
		Content string `json:"content"`
	}
	path := fmt.Sprintf("/v1.0/doc/documents/%s/export?%s", url.PathEscape(ref.ID), q.Encode())
	if err := c.get(ctx, token, path, &out); err != nil {
		return nil, err
	}
	md := out.Content
	if md == "" {
		md = "# " + ref.Title + "\n\n(内容导出为空)"
	}
	return &Doc{Ref: ref, Markdown: md}, nil
}

func (c *DingTalkConnector) auth(ctx context.Context, config json.RawMessage) (*dingtalkConfig, string, error) {
	var cfg dingtalkConfig
	if err := decodeConfig(config, &cfg); err != nil {
		return nil, "", err
	}
	if cfg.AppKey == "" || cfg.AppSecret == "" || cfg.WorkspaceID == "" || cfg.OperatorID == "" {
		return nil, "", fmt.Errorf("%w: dingtalk requires app_key/app_secret/workspace_id/operator_id", ErrBadConfig)
	}
	body, _ := json.Marshal(map[string]string{"appKey": cfg.AppKey, "appSecret": cfg.AppSecret})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.base()+"/v1.0/oauth2/accessToken", bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	var out struct {
		AccessToken string `json:"accessToken"`
	}
	if err := doJSON(c.HTTP, req, &out); err != nil {
		return nil, "", fmt.Errorf("dingtalk auth: %w", err)
	}
	if out.AccessToken == "" {
		return nil, "", fmt.Errorf("dingtalk auth: empty accessToken")
	}
	return &cfg, out.AccessToken, nil
}

// listNodes walks the workspace node tree; parentNodeID equal to the
// workspace id means "root children" per the DingTalk wiki API contract.
func (c *DingTalkConnector) listNodes(ctx context.Context, token string, cfg *dingtalkConfig, parentNodeID string) ([]DocRef, error) {
	var refs []DocRef
	nextToken := ""
	for {
		q := url.Values{
			"operatorId":   {cfg.OperatorID},
			"parentNodeId": {parentNodeID},
			"maxResults":   {"50"},
		}
		if nextToken != "" {
			q.Set("nextToken", nextToken)
		}
		var out struct {
			Nodes []struct {
				NodeID      string `json:"nodeId"`
				Name        string `json:"name"`
				Type        string `json:"type"` // FILE | FOLDER
				DocumentID  string `json:"documentId"`
				HasChildren bool   `json:"hasChildren"`
			} `json:"nodes"`
			HasMore   bool   `json:"hasMore"`
			NextToken string `json:"nextToken"`
		}
		path := fmt.Sprintf("/v2.0/wiki/workspaces/%s/nodes?%s", url.PathEscape(cfg.WorkspaceID), q.Encode())
		if err := c.get(ctx, token, path, &out); err != nil {
			return nil, err
		}
		for _, n := range out.Nodes {
			if n.Type == "FILE" && n.DocumentID != "" {
				refs = append(refs, DocRef{
					ID:    n.DocumentID,
					Title: n.Name,
					Path:  DocPath(n.Name, n.DocumentID),
				})
			}
			if n.HasChildren {
				children, err := c.listNodes(ctx, token, cfg, n.NodeID)
				if err != nil {
					return nil, err
				}
				refs = append(refs, children...)
			}
		}
		if !out.HasMore {
			return refs, nil
		}
		nextToken = out.NextToken
	}
}

func (c *DingTalkConnector) get(ctx context.Context, token, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base()+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-acs-dingtalk-access-token", token)
	return doJSON(c.HTTP, req, out)
}
