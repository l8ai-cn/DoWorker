package tools

import (
	"strings"
	"testing"
)

func TestAppliedResourceSummaryFormatText(t *testing.T) {
	resource := &AppliedResourceSummary{
		Kind: "Workflow", Name: "nightly-review", Revision: 3,
		WorkerSpecSnapshotID: 42,
	}

	text := resource.FormatText()

	for _, expected := range []string{
		"Workflow/nightly-review@r3",
		"Snapshot: 42",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("FormatText() = %q, missing %q", text, expected)
		}
	}
}

func TestWorkflowCreateResultIncludesAppliedResource(t *testing.T) {
	result := &WorkflowCreateResult{
		Workflow: &WorkflowSummary{
			Slug: "nightly-review", Name: "Nightly Review",
			Status: "disabled", ExecutionMode: "direct",
		},
		Resource: &AppliedResourceSummary{
			Kind: "Workflow", Name: "nightly-review", Revision: 1,
			WorkerSpecSnapshotID: 43,
		},
	}

	text := result.FormatText()

	if !strings.Contains(text, "Workflow/nightly-review@r1") ||
		!strings.Contains(text, "Snapshot: 43") {
		t.Fatalf("FormatText() = %q", text)
	}
}
