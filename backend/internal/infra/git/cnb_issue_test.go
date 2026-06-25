package git

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCNBListIssuesBuildsQueryAndDecodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/owner/repo/-/issues" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("state") != "open" || q.Get("labels") != "bug,help" || q.Get("assignees") != "-" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		if q.Get("page") != "1" || q.Get("page_size") != "100" {
			t.Fatalf("unexpected paging: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"number": 12,
			"state": "open",
			"title": "Fix login",
			"body": "steps...",
			"priority": "high",
			"labels": [{"name": "bug"}, {"name": "help"}],
			"assignees": [],
			"author": {"username": "alice"},
			"web_url": "https://cnb.cool/owner/repo/-/issues/12"
		}]`))
	}))
	defer server.Close()

	client, err := NewIssueClient(ProviderTypeCNB, server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewIssueClient: %v", err)
	}
	issues, err := client.ListIssues(context.Background(), "owner/repo", IssueListOptions{
		State:     "open",
		Labels:    []string{"bug", "help"},
		Assignees: "-",
	})
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("len(issues) = %d, want 1", len(issues))
	}
	got := issues[0]
	if got.Number != 12 || got.Title != "Fix login" || got.Priority != "high" {
		t.Fatalf("unexpected issue: %+v", got)
	}
	if len(got.Labels) != 2 || got.Labels[0] != "bug" {
		t.Fatalf("unexpected labels: %+v", got.Labels)
	}
	if got.Author != "alice" {
		t.Fatalf("Author = %q, want alice", got.Author)
	}
}

func TestCNBGetIssueFallsBackToRequestedNumber(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/owner/repo/-/issues/7" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"title": "No number field", "state": "open"}`))
	}))
	defer server.Close()

	client, err := NewIssueClient(ProviderTypeCNB, server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewIssueClient: %v", err)
	}
	issue, err := client.GetIssue(context.Background(), "owner/repo", 7)
	if err != nil {
		t.Fatalf("GetIssue: %v", err)
	}
	if issue.Number != 7 {
		t.Fatalf("Number = %d, want fallback 7", issue.Number)
	}
}

func TestCNBPostIssueCommentSendsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/owner/repo/-/issues/3/comments" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]string
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if payload["body"] != "hello" {
			t.Fatalf("body = %q, want hello", payload["body"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id": 99, "body": "hello"}`))
	}))
	defer server.Close()

	client, err := NewIssueClient(ProviderTypeCNB, server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewIssueClient: %v", err)
	}
	comment, err := client.PostIssueComment(context.Background(), "owner/repo", 3, "hello")
	if err != nil {
		t.Fatalf("PostIssueComment: %v", err)
	}
	if comment.ID != "99" || comment.Body != "hello" {
		t.Fatalf("unexpected comment: %+v", comment)
	}
}

func TestCNBListIssueCommentsDecodesAuthor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/owner/repo/-/issues/5/comments" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id": "c1", "body": "<!-- claim -->", "author": {"username": "bot"}}]`))
	}))
	defer server.Close()

	client, err := NewIssueClient(ProviderTypeCNB, server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewIssueClient: %v", err)
	}
	comments, err := client.ListIssueComments(context.Background(), "owner/repo", 5)
	if err != nil {
		t.Fatalf("ListIssueComments: %v", err)
	}
	if len(comments) != 1 || comments[0].ID != "c1" || comments[0].Author != "bot" {
		t.Fatalf("unexpected comments: %+v", comments)
	}
}

func TestNewIssueClientRejectsNonCNB(t *testing.T) {
	if _, err := NewIssueClient(ProviderTypeGitHub, "", ""); err != ErrProviderNotSupported {
		t.Fatalf("err = %v, want ErrProviderNotSupported", err)
	}
}
