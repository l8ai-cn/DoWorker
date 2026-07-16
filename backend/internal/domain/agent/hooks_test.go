package agent

import "testing"

func TestAgent_ValidateIdentifiers_RejectsInvalidSlug(t *testing.T) {
	a := &Agent{Slug: "Bad.Slug"}
	if err := a.ValidateIdentifiers(); err == nil {
		t.Fatal("validator should reject slug with dot")
	}
}

func TestAgent_ValidateIdentifiers_AcceptsValidSlug(t *testing.T) {
	a := &Agent{Slug: "claude-code", AdapterID: "claude-stream-json"}
	if err := a.ValidateIdentifiers(); err != nil {
		t.Errorf("validator rejected valid slug: %v", err)
	}
}

func TestAgent_ValidateIdentifiers_RejectsInvalidAdapterID(t *testing.T) {
	a := &Agent{Slug: "claude-code", AdapterID: "Claude.Stream"}
	if err := a.ValidateIdentifiers(); err == nil {
		t.Fatal("validator should reject adapter_id with punctuation")
	}
}
