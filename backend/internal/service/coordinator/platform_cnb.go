package coordinator

import (
	"context"
	"fmt"
	"sort"
	"strings"

	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
	"github.com/anthropics/agentsmesh/backend/internal/infra/git"
)

const claimMarkerPrefix = "agentsmesh-coordinator:claim"

// cnbPlatform implements TaskPlatform over the CNB issues REST API. Claiming is
// a marker comment (ported from auto-harness cnb.Driver): the first comment that
// carries a claim marker wins, and the coordinator treats its own key as an
// idempotent re-claim.
type cnbPlatform struct {
	client git.IssueClient
}

func NewCNBPlatform(client git.IssueClient) TaskPlatform {
	return &cnbPlatform{client: client}
}

func (p *cnbPlatform) PlatformType() string { return coordinatordom.PlatformTypeCNB }

func (p *cnbPlatform) DiscoverTasks(ctx context.Context, repo string, policy coordinatordom.ClaimPolicy) ([]ExternalTask, error) {
	opts := git.IssueListOptions{PageSize: 100}
	if len(policy.States) == 1 {
		opts.State = policy.States[0]
	}
	if len(policy.Labels) > 0 {
		opts.Labels = policy.Labels
	}
	if policy.UnassignedOnly {
		opts.Assignees = "-"
	}
	issues, err := p.client.ListIssues(ctx, repo, opts)
	if err != nil {
		return nil, err
	}
	tasks := make([]ExternalTask, 0, len(issues))
	for _, issue := range issues {
		tasks = append(tasks, issueToTask(issue))
	}
	return tasks, nil
}

func (p *cnbPlatform) TryClaim(ctx context.Context, repo string, task ExternalTask, claimKey string) (ClaimResult, error) {
	active, err := p.activeClaim(ctx, repo, task.Number)
	if err != nil {
		return ClaimResult{}, err
	}
	if active != "" && active != claimKey {
		return ClaimResult{Claimed: false, Reason: "task already claimed", Marker: active}, nil
	}
	if active == claimKey {
		return ClaimResult{Claimed: true, Reason: "idempotent claim", Marker: active}, nil
	}
	body := claimBody(claimKey, task)
	if _, err := p.client.PostIssueComment(ctx, repo, task.Number, body); err != nil {
		return ClaimResult{}, err
	}
	active, err = p.activeClaim(ctx, repo, task.Number)
	if err != nil {
		return ClaimResult{}, err
	}
	if active == claimKey {
		return ClaimResult{Claimed: true, Marker: body}, nil
	}
	if active != "" {
		return ClaimResult{Claimed: false, Reason: "lost claim race", Marker: active}, nil
	}
	return ClaimResult{Claimed: false, Reason: "claim marker not visible after posting"}, nil
}

func (p *cnbPlatform) PostFeedback(ctx context.Context, repo string, task ExternalTask, body string) error {
	_, err := p.client.PostIssueComment(ctx, repo, task.Number, body)
	return err
}

// activeClaim returns the claim key of the earliest unreleased claim marker, or
// "" when the issue is unclaimed.
func (p *cnbPlatform) activeClaim(ctx context.Context, repo string, number int) (string, error) {
	comments, err := p.client.ListIssueComments(ctx, repo, number)
	if err != nil {
		return "", err
	}
	sort.SliceStable(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})
	for _, comment := range comments {
		if key, ok := parseClaimMarker(comment.Body); ok {
			return key, nil
		}
	}
	return "", nil
}

func issueToTask(issue *git.Issue) ExternalTask {
	return ExternalTask{
		ExternalID:  fmt.Sprintf("issue:%d", issue.Number),
		Number:      issue.Number,
		Kind:        "issue",
		Title:       issue.Title,
		Description: issue.Body,
		State:       issue.State,
		Type:        inferType(issue.Labels),
		Priority:    issue.Priority,
		Labels:      issue.Labels,
		Assignees:   issue.Assignees,
		URL:         issue.WebURL,
	}
}

func inferType(labels []string) string {
	for _, label := range labels {
		switch strings.ToLower(label) {
		case "bug", "defect":
			return "bug"
		case "feature", "enhancement":
			return "feature"
		}
	}
	return "issue"
}

func claimBody(claimKey string, task ExternalTask) string {
	return fmt.Sprintf("<!-- %s key=%q -->\nDo Worker Coordinator claimed this task (%s).",
		claimMarkerPrefix, claimKey, task.ExternalID)
}

func parseClaimMarker(body string) (string, bool) {
	start := strings.Index(body, claimMarkerPrefix)
	if start < 0 {
		return "", false
	}
	line := body[start:]
	if end := strings.Index(line, "-->"); end >= 0 {
		line = line[:end]
	}
	const keyToken = "key="
	idx := strings.Index(line, keyToken)
	if idx < 0 {
		return "", false
	}
	rest := strings.TrimSpace(line[idx+len(keyToken):])
	rest = strings.Trim(rest, `"`)
	if rest == "" {
		return "", false
	}
	return rest, true
}
