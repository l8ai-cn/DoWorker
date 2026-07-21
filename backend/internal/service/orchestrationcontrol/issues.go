package orchestrationcontrol

import (
	"sort"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
)

func staleOptionsIssue() control.PlanIssue {
	return control.PlanIssue{
		Severity: control.PlanIssueBlocking,
		Path:     "/spec",
		Code:     "stale-options",
		Message:  "Available options changed; refresh and plan again.",
	}
}

func normalizePlanIssues(
	source []control.PlanIssue,
) ([]control.PlanIssue, error) {
	issues := append([]control.PlanIssue{}, source...)
	for index := range issues {
		if err := issues[index].Validate(); err != nil {
			return nil, err
		}
	}
	sort.Slice(issues, func(left, right int) bool {
		if issues[left].Severity != issues[right].Severity {
			return issues[left].Severity < issues[right].Severity
		}
		if issues[left].Path != issues[right].Path {
			return issues[left].Path < issues[right].Path
		}
		if issues[left].Code != issues[right].Code {
			return issues[left].Code < issues[right].Code
		}
		return issues[left].Message < issues[right].Message
	})
	return issues, nil
}

func hasBlockingIssues(issues []control.PlanIssue) bool {
	for _, issue := range issues {
		if issue.Severity == control.PlanIssueBlocking {
			return true
		}
	}
	return false
}
