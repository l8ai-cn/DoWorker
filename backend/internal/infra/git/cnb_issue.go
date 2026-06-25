package git

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (p *CNBProvider) ListIssues(ctx context.Context, repo string, opts IssueListOptions) ([]*Issue, error) {
	cleanRepo := strings.Trim(repo, "/")
	query := url.Values{}
	if opts.State != "" {
		query.Set("state", opts.State)
	}
	if len(opts.Labels) > 0 {
		query.Set("labels", strings.Join(opts.Labels, ","))
	}
	if opts.Priority != "" {
		query.Set("priority", opts.Priority)
	}
	if opts.Assignees != "" {
		query.Set("assignees", opts.Assignees)
	}
	if opts.Keyword != "" {
		query.Set("keyword", opts.Keyword)
	}
	query.Set("page", strconv.Itoa(maxInt(opts.Page, 1)))
	query.Set("page_size", strconv.Itoa(clampPageSize(opts.PageSize)))

	resp, err := p.doRequest(ctx, http.MethodGet, fmt.Sprintf("/%s/-/issues?%s", cleanRepo, query.Encode()), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raws []cnbIssue
	if err := json.NewDecoder(resp.Body).Decode(&raws); err != nil {
		return nil, err
	}
	issues := make([]*Issue, 0, len(raws))
	for i := range raws {
		issues = append(issues, raws[i].toIssue(p))
	}
	return issues, nil
}

func (p *CNBProvider) GetIssue(ctx context.Context, repo string, number int) (*Issue, error) {
	cleanRepo := strings.Trim(repo, "/")
	resp, err := p.doRequest(ctx, http.MethodGet, fmt.Sprintf("/%s/-/issues/%d", cleanRepo, number), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw cnbIssue
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if raw.Number == 0 {
		raw.Number = number
	}
	return raw.toIssue(p), nil
}

func (p *CNBProvider) ListIssueComments(ctx context.Context, repo string, number int) ([]*IssueComment, error) {
	cleanRepo := strings.Trim(repo, "/")
	resp, err := p.doRequest(ctx, http.MethodGet, fmt.Sprintf("/%s/-/issues/%d/comments?page=1&page_size=100", cleanRepo, number), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raws []cnbComment
	if err := json.NewDecoder(resp.Body).Decode(&raws); err != nil {
		return nil, err
	}
	comments := make([]*IssueComment, 0, len(raws))
	for i := range raws {
		comments = append(comments, raws[i].toComment())
	}
	return comments, nil
}

func (p *CNBProvider) PostIssueComment(ctx context.Context, repo string, number int, body string) (*IssueComment, error) {
	cleanRepo := strings.Trim(repo, "/")
	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return nil, err
	}
	resp, err := p.doRequest(ctx, http.MethodPost, fmt.Sprintf("/%s/-/issues/%d/comments", cleanRepo, number), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw cnbComment
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		// CNB may return an empty/non-JSON body on success; treat as posted.
		return &IssueComment{Body: body, CreatedAt: time.Now().UTC()}, nil
	}
	comment := raw.toComment()
	if comment.Body == "" {
		comment.Body = body
	}
	return comment, nil
}

type cnbIssue struct {
	Number    int             `json:"number"`
	State     string          `json:"state"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Priority  string          `json:"priority"`
	Labels    json.RawMessage `json:"labels"`
	Assignees json.RawMessage `json:"assignees"`
	Author    json.RawMessage `json:"author"`
	URL       string          `json:"url"`
	WebURL    string          `json:"web_url"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (raw *cnbIssue) toIssue(p *CNBProvider) *Issue {
	webURL := firstNonEmpty(raw.WebURL, raw.URL)
	authors := decodeCNBUsers(raw.Author)
	author := ""
	if len(authors) > 0 {
		author = authors[0]
	}
	return &Issue{
		Number:    raw.Number,
		Title:     raw.Title,
		Body:      raw.Body,
		State:     raw.State,
		Labels:    decodeCNBLabels(raw.Labels),
		Priority:  raw.Priority,
		Assignees: decodeCNBUsers(raw.Assignees),
		Author:    author,
		WebURL:    webURL,
		CreatedAt: raw.CreatedAt,
		UpdatedAt: raw.UpdatedAt,
	}
}

type cnbComment struct {
	ID        json.RawMessage `json:"id"`
	Body      string          `json:"body"`
	Author    json.RawMessage `json:"author"`
	CreatedAt time.Time       `json:"created_at"`
}

func (raw *cnbComment) toComment() *IssueComment {
	authors := decodeCNBUsers(raw.Author)
	author := ""
	if len(authors) > 0 {
		author = authors[0]
	}
	return &IssueComment{
		ID:        decodeFlexID(raw.ID),
		Body:      raw.Body,
		Author:    author,
		CreatedAt: raw.CreatedAt,
	}
}

func decodeCNBLabels(raw json.RawMessage) []string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var stringsOnly []string
	if err := json.Unmarshal(raw, &stringsOnly); err == nil {
		return stringsOnly
	}
	var objects []struct {
		Name string `json:"name"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &objects); err != nil {
		return nil
	}
	labels := make([]string, 0, len(objects))
	for _, object := range objects {
		if label := firstNonEmpty(object.Name, object.Text); label != "" {
			labels = append(labels, label)
		}
	}
	return labels
}

func decodeCNBUsers(raw json.RawMessage) []string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var single struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Nickname string `json:"nickname"`
	}
	if err := json.Unmarshal(raw, &single); err == nil && (single.Username != "" || single.Name != "" || single.Nickname != "") {
		return []string{firstNonEmpty(single.Username, single.Name, single.Nickname)}
	}
	var objects []struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Nickname string `json:"nickname"`
	}
	if err := json.Unmarshal(raw, &objects); err == nil {
		users := make([]string, 0, len(objects))
		for _, object := range objects {
			if user := firstNonEmpty(object.Username, object.Name, object.Nickname); user != "" {
				users = append(users, user)
			}
		}
		return users
	}
	var stringsOnly []string
	if err := json.Unmarshal(raw, &stringsOnly); err == nil {
		return stringsOnly
	}
	return nil
}

func decodeFlexID(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString
	}
	var asNumber json.Number
	if err := json.Unmarshal(raw, &asNumber); err == nil {
		return asNumber.String()
	}
	return strings.Trim(string(raw), `"`)
}

func clampPageSize(size int) int {
	if size <= 0 {
		return 100
	}
	if size > 100 {
		return 100
	}
	return size
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
