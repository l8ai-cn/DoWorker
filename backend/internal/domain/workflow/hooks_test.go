package workflow

import "testing"

func TestLoop_ValidateIdentifiers_RejectsInvalidSlug(t *testing.T) {
	l := &Workflow{Slug: "Workflow.Bad", AgentSlug: ""}
	if err := l.ValidateIdentifiers(); err == nil {
		t.Fatal("validator should reject slug with dot")
	}
}

func TestLoop_ValidateIdentifiers_RejectsInvalidAgentSlug(t *testing.T) {
	l := &Workflow{Slug: "ok-workflow", AgentSlug: "Agent.Bad"}
	if err := l.ValidateIdentifiers(); err == nil {
		t.Fatal("validator should reject agent_slug with dot")
	}
}

func TestLoop_ValidateIdentifiers_AcceptsValidSlugs(t *testing.T) {
	l := &Workflow{Slug: "my-workflow", AgentSlug: "claude-code"}
	if err := l.ValidateIdentifiers(); err != nil {
		t.Errorf("validator rejected valid slugs: %v", err)
	}
}

func TestLoop_ValidateIdentifiers_EmptyAgentSlugAllowed(t *testing.T) {
	l := &Workflow{Slug: "my-workflow", AgentSlug: ""}
	if err := l.ValidateIdentifiers(); err != nil {
		t.Errorf("empty agent_slug (optional reference) should pass: %v", err)
	}
}
