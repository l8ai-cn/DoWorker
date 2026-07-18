// Package gitea is the backend's client for the internal Gitea instance that
// hosts knowledge-base repositories. Unlike infra/git (external provider
// import), this client provisions repos programmatically under a dedicated
// namespace org using an admin-scoped service token.
package gitea

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrNotConfigured = errors.New("gitea: internal gitea is not configured")

// HTTPError carries the Gitea HTTP status so higher layers can branch
// (e.g. 404 -> gitops.ErrNotFound) without string matching. Its Error()
// string is format-compatible with the previous fmt.Errorf message.
type HTTPError struct {
	StatusCode int
	Method     string
	Path       string
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("gitea: %s %s → %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

type Config struct {
	BaseURL         string // API + clone base as seen from the backend
	AdminToken      string
	Namespace       string // Gitea org that owns all KB repos
	CloneBaseURL    string // HTTP clone base as seen from runners
	SSHCloneBaseURL string // SSH clone base as seen from runners
	SSHKnownHosts   string // pinned host key distributed to runners
}

func (c Config) Enabled() bool { return c.BaseURL != "" && c.AdminToken != "" }

type Client struct {
	cfg  Config
	http *http.Client
}

func NewClient(cfg Config) *Client {
	if cfg.Namespace == "" {
		cfg.Namespace = "am-kb"
	}
	if cfg.CloneBaseURL == "" {
		cfg.CloneBaseURL = cfg.BaseURL
	}
	return &Client{cfg: cfg, http: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) Namespace() string     { return c.cfg.Namespace }
func (c *Client) SSHKnownHosts() string { return c.cfg.SSHKnownHosts }

type Repo struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+"/api/v1"+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+c.cfg.AdminToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Body:       strings.TrimSpace(string(data)),
		}
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// EnsureNamespace creates the KB namespace org if it does not exist yet.
func (c *Client) EnsureNamespace(ctx context.Context) error {
	err := c.do(ctx, http.MethodGet, "/orgs/"+c.cfg.Namespace, nil, nil)
	if err == nil {
		return nil
	}
	return c.do(ctx, http.MethodPost, "/orgs", map[string]any{
		"username":   c.cfg.Namespace,
		"visibility": "private",
	}, nil)
}

func (c *Client) CreateRepo(ctx context.Context, name, defaultBranch string) (*Repo, error) {
	var repo Repo
	err := c.do(ctx, http.MethodPost, "/orgs/"+c.cfg.Namespace+"/repos", map[string]any{
		"name":           name,
		"default_branch": defaultBranch,
		"auto_init":      false,
		"private":        true,
	}, &repo)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (c *Client) DeleteRepo(ctx context.Context, name string) error {
	err := c.do(ctx, http.MethodDelete, "/repos/"+c.cfg.Namespace+"/"+name, nil, nil)
	if IsHTTPStatus(err, http.StatusNotFound) {
		return nil
	}
	return err
}

// CloneURL returns the runner-facing HTTPS clone URL without credentials.
func (c *Client) CloneURL(name string) string {
	return fmt.Sprintf("%s/%s/%s.git", strings.TrimRight(c.cfg.CloneBaseURL, "/"), c.cfg.Namespace, name)
}

func (c *Client) SSHCloneURL(name string) string {
	if c.cfg.SSHCloneBaseURL == "" {
		return ""
	}
	return fmt.Sprintf(
		"%s/%s/%s.git",
		strings.TrimRight(c.cfg.SSHCloneBaseURL, "/"),
		c.cfg.Namespace,
		name,
	)
}

func escapePath(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		segs[i] = url.PathEscape(s)
	}
	return strings.Join(segs, "/")
}

func IsHTTPStatus(err error, status int) bool {
	var httpErr *HTTPError
	return errors.As(err, &httpErr) && httpErr.StatusCode == status
}
