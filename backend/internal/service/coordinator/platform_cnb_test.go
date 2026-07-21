package coordinator

import (
	"context"
	"testing"
	"time"

	coordinatordom "github.com/l8ai-cn/agentcloud/backend/internal/domain/coordinator"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/git"
)

type fakeIssueClient struct {
	issues   []*git.Issue
	comments map[int][]*git.IssueComment
	posted   map[int][]string
}

func newFakeIssueClient() *fakeIssueClient {
	return &fakeIssueClient{
		comments: map[int][]*git.IssueComment{},
		posted:   map[int][]string{},
	}
}

func (f *fakeIssueClient) ListIssues(_ context.Context, _ string, _ git.IssueListOptions) ([]*git.Issue, error) {
	return f.issues, nil
}

func (f *fakeIssueClient) GetIssue(_ context.Context, _ string, number int) (*git.Issue, error) {
	for _, issue := range f.issues {
		if issue.Number == number {
			return issue, nil
		}
	}
	return nil, git.ErrNotFound
}

func (f *fakeIssueClient) ListIssueComments(_ context.Context, _ string, number int) ([]*git.IssueComment, error) {
	return f.comments[number], nil
}

func (f *fakeIssueClient) PostIssueComment(_ context.Context, _ string, number int, body string) (*git.IssueComment, error) {
	f.posted[number] = append(f.posted[number], body)
	comment := &git.IssueComment{Body: body, CreatedAt: time.Now().Add(time.Duration(len(f.comments[number])) * time.Second)}
	f.comments[number] = append(f.comments[number], comment)
	return comment, nil
}

func TestCNBPlatformDiscoverMapsIssues(t *testing.T) {
	fake := newFakeIssueClient()
	fake.issues = []*git.Issue{
		{Number: 1, Title: "Crash", Body: "boom", State: "open", Labels: []string{"bug"}},
	}
	platform := NewCNBPlatform(fake)
	tasks, err := platform.DiscoverTasks(context.Background(), "owner/repo", coordinatordom.ClaimPolicy{})
	if err != nil {
		t.Fatalf("DiscoverTasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ExternalID != "issue:1" || tasks[0].Type != "bug" {
		t.Fatalf("unexpected tasks: %+v", tasks)
	}
}

func TestCNBPlatformClaimPostsMarkerAndIsIdempotent(t *testing.T) {
	fake := newFakeIssueClient()
	platform := NewCNBPlatform(fake)
	task := ExternalTask{ExternalID: "issue:5", Number: 5}

	result, err := platform.TryClaim(context.Background(), "owner/repo", task, "project=1 ticket=9")
	if err != nil {
		t.Fatalf("TryClaim: %v", err)
	}
	if !result.Claimed {
		t.Fatalf("expected claim to succeed, got %+v", result)
	}
	if len(fake.posted[5]) != 1 {
		t.Fatalf("expected exactly 1 marker comment, got %d", len(fake.posted[5]))
	}

	again, err := platform.TryClaim(context.Background(), "owner/repo", task, "project=1 ticket=9")
	if err != nil {
		t.Fatalf("TryClaim idempotent: %v", err)
	}
	if !again.Claimed || again.Reason != "idempotent claim" {
		t.Fatalf("expected idempotent claim, got %+v", again)
	}
	if len(fake.posted[5]) != 1 {
		t.Fatalf("idempotent claim must not post another marker, got %d", len(fake.posted[5]))
	}
}

func TestCNBPlatformClaimRejectsForeignClaim(t *testing.T) {
	fake := newFakeIssueClient()
	fake.comments[7] = []*git.IssueComment{
		{Body: `<!-- agentcloud-coordinator:claim key="project=2 ticket=3" -->`, CreatedAt: time.Now()},
	}
	platform := NewCNBPlatform(fake)
	task := ExternalTask{ExternalID: "issue:7", Number: 7}

	result, err := platform.TryClaim(context.Background(), "owner/repo", task, "project=1 ticket=9")
	if err != nil {
		t.Fatalf("TryClaim: %v", err)
	}
	if result.Claimed {
		t.Fatalf("expected claim rejection for foreign marker, got %+v", result)
	}
	if len(fake.posted[7]) != 0 {
		t.Fatalf("must not post marker when already claimed, got %d", len(fake.posted[7]))
	}
}
