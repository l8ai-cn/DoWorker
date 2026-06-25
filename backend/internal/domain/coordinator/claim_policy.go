package coordinator

import "strings"

// Candidate is the platform-agnostic view of a discovered task that ClaimPolicy
// is evaluated against.
type Candidate struct {
	Title       string
	Description string
	State       string
	Type        string
	Priority    string
	Labels      []string
	Assignees   []string
}

// Matches ports auto-harness coordinator.matchesClaimPolicy: every configured
// dimension must match (case-insensitive); empty dimensions are wildcards.
func (p ClaimPolicy) Matches(c Candidate) (bool, string) {
	if len(p.States) > 0 && !containsFold(p.States, c.State) {
		return false, "state does not match"
	}
	if len(p.TaskTypes) > 0 && !containsFold(p.TaskTypes, c.Type) {
		return false, "task type does not match"
	}
	if len(p.Priorities) > 0 && !containsFold(p.Priorities, c.Priority) {
		return false, "priority does not match"
	}
	if p.UnassignedOnly && len(c.Assignees) > 0 {
		return false, "task is assigned"
	}
	for _, label := range p.Labels {
		if !containsFold(c.Labels, label) {
			return false, "label does not match"
		}
	}
	if len(p.TitleKeywords) > 0 && !containsKeyword(c.Title, p.TitleKeywords) {
		return false, "title keyword does not match"
	}
	if len(p.BodyKeywords) > 0 && !containsKeyword(c.Description, p.BodyKeywords) {
		return false, "body keyword does not match"
	}
	return true, ""
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(target)) {
			return true
		}
	}
	return false
}

func containsKeyword(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, keyword := range keywords {
		keyword = strings.ToLower(strings.TrimSpace(keyword))
		if keyword != "" && strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}
