package coordinator

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	coordinatordom "github.com/l8ai-cn/agentcloud/backend/internal/domain/coordinator"
)

func TestLinearDiscoverTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "tok" {
			t.Errorf("missing auth header, got %q", r.Header.Get("Authorization"))
		}
		_, _ = io.WriteString(w, `{"data":{"issues":{"nodes":[
			{"id":"uuid-1","identifier":"ENG-1","title":"Fix","description":"body","url":"u","state":{"name":"Todo","type":"unstarted"},"labels":{"nodes":[{"name":"bug"}]},"assignees":{"name":"alice"}}
		]}}}`)
	}))
	defer server.Close()

	p := NewLinearPlatform("tok", server.URL)
	tasks, err := p.DiscoverTasks(context.Background(), "ENG", coordinatordom.ClaimPolicy{})
	if err != nil {
		t.Fatalf("DiscoverTasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("tasks = %d, want 1", len(tasks))
	}
	if tasks[0].ExternalID != "uuid-1" || tasks[0].Title != "Fix" {
		t.Fatalf("unexpected task: %+v", tasks[0])
	}
	if len(tasks[0].Labels) != 1 || tasks[0].Labels[0] != "bug" {
		t.Fatalf("labels = %v", tasks[0].Labels)
	}
	if len(tasks[0].Assignees) != 1 || tasks[0].Assignees[0] != "alice" {
		t.Fatalf("assignees = %v", tasks[0].Assignees)
	}
}

func TestLinearTryClaimAndFeedback(t *testing.T) {
	var mutations int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Query string `json:"query"`
		}
		_ = json.Unmarshal(body, &req)
		switch {
		case strings.Contains(req.Query, "comments"):
			_, _ = io.WriteString(w, `{"data":{"issue":{"comments":{"nodes":[]}}}}`)
		case strings.Contains(req.Query, "commentCreate"):
			mutations++
			_, _ = io.WriteString(w, `{"data":{"commentCreate":{"success":true}}}`)
		default:
			t.Errorf("unexpected query: %s", req.Query)
		}
	}))
	defer server.Close()

	p := NewLinearPlatform("tok", server.URL)
	task := ExternalTask{ExternalID: "uuid-1", Title: "t"}

	claim, err := p.TryClaim(context.Background(), "ENG", task, "k1")
	if err != nil {
		t.Fatalf("TryClaim: %v", err)
	}
	if !claim.Claimed {
		t.Fatalf("expected claim to succeed")
	}
	if err := p.PostFeedback(context.Background(), "ENG", task, "done"); err != nil {
		t.Fatalf("PostFeedback: %v", err)
	}
	if mutations != 2 {
		t.Fatalf("commentCreate calls = %d, want 2 (claim + feedback)", mutations)
	}
}
