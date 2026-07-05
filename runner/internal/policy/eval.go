package policy

// Verdict is the outcome of evaluating a tool permission against org rules.
type Verdict string

const (
	VerdictAllow Verdict = "allow"
	VerdictDeny  Verdict = "deny"
	VerdictAsk   Verdict = "ask"
)

// Rule is a snapshot policy row shipped with CreatePodCommand.
type Rule struct {
	ToolPattern string `json:"tool_pattern"`
	PathPattern string `json:"path_pattern,omitempty"`
	Verdict     Verdict `json:"verdict"`
	Priority    int    `json:"priority"`
}

// Evaluate returns the highest-priority matching verdict, or ASK when no rule matches.
// Fail-safe: unreachable policy data must not auto-allow.
func Evaluate(rules []Rule, toolName, path string) Verdict {
	var best *Rule
	for i := range rules {
		r := &rules[i]
		if !matchPattern(r.ToolPattern, toolName) {
			continue
		}
		if r.PathPattern != "" && !matchPattern(r.PathPattern, path) {
			continue
		}
		if best == nil || r.Priority > best.Priority {
			best = r
		}
	}
	if best == nil {
		return VerdictAsk
	}
	return best.Verdict
}

func matchPattern(pattern, value string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	return pattern == value
}
