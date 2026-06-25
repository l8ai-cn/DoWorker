package git

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCNBGetProjectMapsHTTPSOnlyRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/owner/repo" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": 456,
			"name": "repo",
			"path": "owner/repo",
			"description": "CNB repository",
			"default_branch": "main",
			"html_url": "https://cnb.cool/owner/repo",
			"clone_url": "https://cnb.cool/owner/repo",
			"visibility": "private"
		}`))
	}))
	defer server.Close()

	provider, err := NewCNBProvider(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewCNBProvider: %v", err)
	}
	project, err := provider.GetProject(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if project.Slug != "owner/repo" {
		t.Fatalf("Slug = %q, want owner/repo", project.Slug)
	}
	if project.HttpCloneURL != "https://cnb.cool/owner/repo" {
		t.Fatalf("HttpCloneURL = %q", project.HttpCloneURL)
	}
	if project.SSHCloneURL != "" {
		t.Fatalf("CNB must not expose SSH clone URL, got %q", project.SSHCloneURL)
	}
}

func TestCNBListProjectsUsesUserReposEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/repos" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" || r.URL.Query().Get("page_size") != "30" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"id": 1,
			"name": "repo",
			"path": "owner/repo",
			"html_url": "https://cnb.cool/owner/repo",
			"clone_url": "https://cnb.cool/owner/repo"
		}]`))
	}))
	defer server.Close()

	provider, err := NewCNBProvider(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewCNBProvider: %v", err)
	}
	projects, err := provider.ListProjects(context.Background(), 2, 30)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 1 || projects[0].Slug != "owner/repo" {
		t.Fatalf("unexpected projects: %+v", projects)
	}
}
