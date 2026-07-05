package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// GoogleDriveConnector syncs Google Docs from a Drive folder, exported as
// markdown (Drive natively supports text/markdown export for Docs).
//
// source_config: {"access_token": "...", "folder_id": "..."}
// The access token must carry drive.readonly scope; refresh handling is the
// caller's concern (store a long-lived token or re-provision periodically).
type GoogleDriveConnector struct {
	HTTP    *http.Client
	BaseURL string // test override; empty = https://www.googleapis.com
}

type googleConfig struct {
	AccessToken string `json:"access_token"`
	FolderID    string `json:"folder_id"`
}

func (c *GoogleDriveConnector) SourceType() string { return "google" }

func (c *GoogleDriveConnector) base() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return "https://www.googleapis.com"
}

func (c *GoogleDriveConnector) ListDocs(ctx context.Context, config json.RawMessage) ([]DocRef, error) {
	cfg, err := c.config(config)
	if err != nil {
		return nil, err
	}
	var refs []DocRef
	pageToken := ""
	for {
		q := url.Values{
			"q": {fmt.Sprintf("'%s' in parents and mimeType='application/vnd.google-apps.document' and trashed=false",
				cfg.FolderID)},
			"fields":   {"nextPageToken,files(id,name)"},
			"pageSize": {"100"},
		}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		var out struct {
			NextPageToken string `json:"nextPageToken"`
			Files         []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"files"`
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			c.base()+"/drive/v3/files?"+q.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
		if err := doJSON(c.HTTP, req, &out); err != nil {
			return nil, fmt.Errorf("google drive list: %w", err)
		}
		for _, f := range out.Files {
			refs = append(refs, DocRef{ID: f.ID, Title: f.Name, Path: DocPath(f.Name, f.ID)})
		}
		if out.NextPageToken == "" {
			return refs, nil
		}
		pageToken = out.NextPageToken
	}
}

func (c *GoogleDriveConnector) FetchDoc(ctx context.Context, config json.RawMessage, ref DocRef) (*Doc, error) {
	cfg, err := c.config(config)
	if err != nil {
		return nil, err
	}
	exportURL := fmt.Sprintf("%s/drive/v3/files/%s/export?mimeType=%s",
		c.base(), url.PathEscape(ref.ID), url.QueryEscape("text/markdown"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, exportURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google drive export: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google drive export: status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	return &Doc{Ref: ref, Markdown: string(body)}, nil
}

func (c *GoogleDriveConnector) config(config json.RawMessage) (*googleConfig, error) {
	var cfg googleConfig
	if err := decodeConfig(config, &cfg); err != nil {
		return nil, err
	}
	if cfg.AccessToken == "" || cfg.FolderID == "" {
		return nil, fmt.Errorf("%w: google requires access_token/folder_id", ErrBadConfig)
	}
	return &cfg, nil
}
