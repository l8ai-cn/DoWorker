package policy

import "testing"

func TestEvaluate_NoRulesAsk(t *testing.T) {
	if got := Evaluate(nil, "Bash", "/tmp"); got != VerdictAsk {
		t.Fatalf("got %q want ask", got)
	}
}

func TestEvaluate_PriorityWins(t *testing.T) {
	rules := []Rule{
		{ToolPattern: "Bash", Verdict: VerdictAllow, Priority: 1},
		{ToolPattern: "Bash", Verdict: VerdictDeny, Priority: 10},
	}
	if got := Evaluate(rules, "Bash", ""); got != VerdictDeny {
		t.Fatalf("got %q want deny", got)
	}
}
