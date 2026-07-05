package coordinator

import "testing"

func TestProject_ValidateIdentifiers_RejectsInvalidSlug(t *testing.T) {
	project := &Project{Slug: "Bad.Slug", AgentSlug: "do-agent"}

	if err := project.ValidateIdentifiers(); err == nil {
		t.Fatalf("ValidateIdentifiers accepted invalid slug")
	}
}

func TestProject_ValidateIdentifiers_RejectsInvalidAgentSlug(t *testing.T) {
	project := &Project{Slug: "auto-project", AgentSlug: "do_agent"}

	if err := project.ValidateIdentifiers(); err == nil {
		t.Fatalf("ValidateIdentifiers accepted invalid agent slug")
	}
}

func TestProject_ValidateIdentifiers_AcceptsValidIdentifiers(t *testing.T) {
	project := &Project{Slug: "auto-project", AgentSlug: "do-agent"}

	if err := project.ValidateIdentifiers(); err != nil {
		t.Fatalf("ValidateIdentifiers rejected valid identifiers: %v", err)
	}
}
