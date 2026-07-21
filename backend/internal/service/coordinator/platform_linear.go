package coordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	coordinatordom "github.com/l8ai-cn/agentcloud/backend/internal/domain/coordinator"
)

const defaultLinearEndpoint = "https://api.linear.app/graphql"

// linearPlatform implements TaskPlatform over the Linear GraphQL API. The "repo"
// argument is interpreted as a Linear team key. Claiming reuses the same marker
// convention as CNB, stored as an issue comment.
type linearPlatform struct {
	endpoint string
	token    string
	client   *http.Client
}

func NewLinearPlatform(token, endpoint string) TaskPlatform {
	if strings.TrimSpace(endpoint) == "" {
		endpoint = defaultLinearEndpoint
	}
	return &linearPlatform{endpoint: endpoint, token: token, client: &http.Client{Timeout: 30 * time.Second}}
}

func (p *linearPlatform) PlatformType() string { return coordinatordom.PlatformTypeLinear }

func (p *linearPlatform) DiscoverTasks(ctx context.Context, team string, _ coordinatordom.ClaimPolicy) ([]ExternalTask, error) {
	query := `query($team:String!){issues(filter:{team:{key:{eq:$team}}}){nodes{id identifier title description url state{name type} labels{nodes{name}} assignees:assignee{name}}}}`
	var resp struct {
		Data struct {
			Issues struct{ Nodes []linearIssue } `json:"issues"`
		} `json:"data"`
	}
	if err := p.do(ctx, query, map[string]any{"team": team}, &resp); err != nil {
		return nil, err
	}
	tasks := make([]ExternalTask, 0, len(resp.Data.Issues.Nodes))
	for _, n := range resp.Data.Issues.Nodes {
		tasks = append(tasks, n.toTask())
	}
	return tasks, nil
}

func (p *linearPlatform) TryClaim(ctx context.Context, _ string, task ExternalTask, claimKey string) (ClaimResult, error) {
	active, err := p.activeClaim(ctx, task.ExternalID)
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
	if err := p.comment(ctx, task.ExternalID, body); err != nil {
		return ClaimResult{}, err
	}
	return ClaimResult{Claimed: true, Marker: body}, nil
}

func (p *linearPlatform) PostFeedback(ctx context.Context, _ string, task ExternalTask, body string) error {
	return p.comment(ctx, task.ExternalID, body)
}

func (p *linearPlatform) activeClaim(ctx context.Context, issueID string) (string, error) {
	query := `query($id:String!){issue(id:$id){comments{nodes{body createdAt}}}}`
	var resp struct {
		Data struct {
			Issue struct {
				Comments struct {
					Nodes []struct {
						Body      string    `json:"body"`
						CreatedAt time.Time `json:"createdAt"`
					}
				} `json:"comments"`
			} `json:"issue"`
		} `json:"data"`
	}
	if err := p.do(ctx, query, map[string]any{"id": issueID}, &resp); err != nil {
		return "", err
	}
	nodes := resp.Data.Issue.Comments.Nodes
	sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].CreatedAt.Before(nodes[j].CreatedAt) })
	for _, n := range nodes {
		if key, ok := parseClaimMarker(n.Body); ok {
			return key, nil
		}
	}
	return "", nil
}

func (p *linearPlatform) comment(ctx context.Context, issueID, body string) error {
	mutation := `mutation($id:String!,$body:String!){commentCreate(input:{issueId:$id,body:$body}){success}}`
	var resp struct {
		Data struct {
			CommentCreate struct {
				Success bool `json:"success"`
			} `json:"commentCreate"`
		} `json:"data"`
	}
	if err := p.do(ctx, mutation, map[string]any{"id": issueID, "body": body}, &resp); err != nil {
		return err
	}
	if !resp.Data.CommentCreate.Success {
		return fmt.Errorf("linear: commentCreate returned success=false")
	}
	return nil
}

func (p *linearPlatform) do(ctx context.Context, query string, variables map[string]any, out any) error {
	payload, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", p.token)
	res, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("linear: graphql status %d", res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(out)
}

type linearIssue struct {
	ID         string `json:"id"`
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	Body       string `json:"description"`
	URL        string `json:"url"`
	State      struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"state"`
	Labels struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"labels"`
	Assignee struct {
		Name string `json:"name"`
	} `json:"assignees"`
}

func (n linearIssue) toTask() ExternalTask {
	labels := make([]string, 0, len(n.Labels.Nodes))
	for _, l := range n.Labels.Nodes {
		labels = append(labels, l.Name)
	}
	var assignees []string
	if n.Assignee.Name != "" {
		assignees = []string{n.Assignee.Name}
	}
	return ExternalTask{
		ExternalID:  n.ID,
		Kind:        "issue",
		Title:       n.Title,
		Description: n.Body,
		State:       n.State.Name,
		Type:        n.State.Type,
		Labels:      labels,
		Assignees:   assignees,
		URL:         n.URL,
	}
}
