package git

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (p *CNBProvider) GetProject(ctx context.Context, projectID string) (*Project, error) {
	cleanID := strings.Trim(projectID, "/")
	resp, err := p.doRequest(ctx, http.MethodGet, "/"+cleanID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var repo cnbRepository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, err
	}
	return p.toProject(cleanID, &repo), nil
}

func (p *CNBProvider) ListProjects(ctx context.Context, page, perPage int) ([]*Project, error) {
	resp, err := p.doRequest(ctx, http.MethodGet, fmt.Sprintf("/user/repos?page=%d&page_size=%d", page, perPage), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var repos []cnbRepository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}
	projects := make([]*Project, 0, len(repos))
	for i := range repos {
		projects = append(projects, p.toProject("", &repos[i]))
	}
	return projects, nil
}

func (p *CNBProvider) SearchProjects(ctx context.Context, query string, page, perPage int) ([]*Project, error) {
	resp, err := p.doRequest(ctx, http.MethodGet, fmt.Sprintf(
		"/user/repos?page=%d&page_size=%d&keyword=%s",
		page,
		perPage,
		url.QueryEscape(query),
	), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var repos []cnbRepository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}
	projects := make([]*Project, 0, len(repos))
	for i := range repos {
		projects = append(projects, p.toProject("", &repos[i]))
	}
	return projects, nil
}

type cnbRepository struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	DefaultBranch string    `json:"default_branch"`
	HTMLURL       string    `json:"html_url"`
	CloneURL      string    `json:"clone_url"`
	Private       bool      `json:"private"`
	Visibility    string    `json:"visibility"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	PushedAt      time.Time `json:"pushed_at"`
}

func (p *CNBProvider) toProject(fallbackPath string, repo *cnbRepository) *Project {
	slug := firstNonEmpty(repo.Path, repo.FullName, fallbackPath)
	webURL := repo.HTMLURL
	if webURL == "" && slug != "" {
		webURL = p.webBaseURL + "/" + slug
	}
	cloneURL := repo.CloneURL
	if cloneURL == "" {
		cloneURL = webURL
	}
	visibility := repo.Visibility
	if visibility == "" {
		if repo.Private {
			visibility = "private"
		} else {
			visibility = "public"
		}
	}
	updatedAt := repo.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = repo.PushedAt
	}

	return &Project{
		ID:            fmt.Sprintf("%d", repo.ID),
		Name:          firstNonEmpty(repo.Name, pathBase(slug)),
		Slug:          slug,
		Description:   repo.Description,
		DefaultBranch: repo.DefaultBranch,
		WebURL:        webURL,
		HttpCloneURL:  cloneURL,
		SSHCloneURL:   "",
		Visibility:    visibility,
		CreatedAt:     repo.CreatedAt,
		UpdatedAt:     updatedAt,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func pathBase(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
