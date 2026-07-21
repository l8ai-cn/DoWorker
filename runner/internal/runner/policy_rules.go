package runner

import (
	"encoding/json"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/policy"
)

func policyRulesFromProto(rules []*runnerv1.PolicyRuleSnapshot) []policy.Rule {
	if len(rules) == 0 {
		return nil
	}
	out := make([]policy.Rule, 0, len(rules))
	for _, r := range rules {
		if r == nil {
			continue
		}
		out = append(out, policy.Rule{
			ToolPattern: r.GetToolPattern(),
			PathPattern: r.GetPathPattern(),
			Verdict:     policy.Verdict(r.GetVerdict()),
			Priority:    int(r.GetPriority()),
		})
	}
	return out
}

func permissionPathFromArgs(argsJSON string) string {
	if argsJSON == "" {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &m); err != nil || m == nil {
		return ""
	}
	for _, key := range []string{"file_path", "path", "filePath", "directory"} {
		if v, ok := m[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}
