package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// FeishuConnector syncs a Feishu (Lark) wiki space. Docs are exported via the
// docx raw_content API, which returns plain text — good enough as the
// immutable raw/ layer; the wiki/ compilation pass restructures it anyway.
//
// source_config: {"app_id": "...", "app_secret": "...", "space_id": "..."}
type FeishuConnector struct {
	HTTP *http.Client
	// BaseURL overrides the API host in tests; empty = production endpoint.
	BaseURL string
}

type feishuConfig struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
	SpaceID   string `json:"space_id"`
}

func (c *FeishuConnector) SourceType() string { return "feishu" }

func (c *FeishuConnector) base() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return "https://open.feishu.cn"
}

func (c *FeishuConnector) ListDocs(ctx context.Context, config json.RawMessage) ([]DocRef, error) {
	cfg, token, err := c.auth(ctx, config)
	if err != nil {
		return nil, err
	}
	return c.listNodes(ctx, token, cfg.SpaceID, "")
}

func (c *FeishuConnector) FetchDoc(ctx context.Context, config json.RawMessage, ref DocRef) (*Doc, error) {
	_, token, err := c.auth(ctx, config)
	if err != nil {
		return nil, err
	}
	var out struct {
		Data struct {
			Content string `json:"content"`
		} `json:"data"`
	}
	path := fmt.Sprintf("/open-apis/docx/v1/documents/%s/raw_content", url.PathEscape(ref.ID))
	if err := c.get(ctx, token, path, &out); err != nil {
		return nil, err
	}
	md := "# " + ref.Title + "\n\n" + out.Data.Content
	return &Doc{Ref: ref, Markdown: md}, nil
}

func (c *FeishuConnector) auth(ctx context.Context, config json.RawMessage) (*feishuConfig, string, error) {
	var cfg feishuConfig
	if err := decodeConfig(config, &cfg); err != nil {
		return nil, "", err
	}
	if cfg.AppID == "" || cfg.AppSecret == "" || cfg.SpaceID == "" {
		return nil, "", fmt.Errorf("%w: feishu requires app_id/app_secret/space_id", ErrBadConfig)
	}
	body, _ := json.Marshal(map[string]string{"app_id": cfg.AppID, "app_secret": cfg.AppSecret})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.base()+"/open-apis/auth/v3/tenant_access_token/internal", bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	var out struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
	}
	if err := doJSON(c.HTTP, req, &out); err != nil {
		return nil, "", fmt.Errorf("feishu auth: %w", err)
	}
	if out.Code != 0 {
		return nil, "", fmt.Errorf("feishu auth: code=%d msg=%s", out.Code, out.Msg)
	}
	return &cfg, out.TenantAccessToken, nil
}

// listNodes walks the wiki space tree depth-first, collecting docx nodes.
func (c *FeishuConnector) listNodes(ctx context.Context, token, spaceID, parent string) ([]DocRef, error) {
	var refs []DocRef
	pageToken := ""
	for {
		q := url.Values{"page_size": {"50"}}
		if parent != "" {
			q.Set("parent_node_token", parent)
		}
		if pageToken != "" {
			q.Set("page_token", pageToken)
		}
		var out struct {
			Data struct {
				Items []struct {
					NodeToken string `json:"node_token"`
					ObjToken  string `json:"obj_token"`
					ObjType   string `json:"obj_type"`
					Title     string `json:"title"`
					HasChild  bool   `json:"has_child"`
				} `json:"items"`
				HasMore   bool   `json:"has_more"`
				PageToken string `json:"page_token"`
			} `json:"data"`
		}
		path := fmt.Sprintf("/open-apis/wiki/v2/spaces/%s/nodes?%s", url.PathEscape(spaceID), q.Encode())
		if err := c.get(ctx, token, path, &out); err != nil {
			return nil, err
		}
		for _, item := range out.Data.Items {
			if item.ObjType == "docx" {
				refs = append(refs, DocRef{
					ID:    item.ObjToken,
					Title: item.Title,
					Path:  DocPath(item.Title, item.ObjToken),
				})
			}
			if item.HasChild {
				children, err := c.listNodes(ctx, token, spaceID, item.NodeToken)
				if err != nil {
					return nil, err
				}
				refs = append(refs, children...)
			}
		}
		if !out.Data.HasMore {
			return refs, nil
		}
		pageToken = out.Data.PageToken
	}
}

func (c *FeishuConnector) get(ctx context.Context, token, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base()+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return doJSON(c.HTTP, req, out)
}
