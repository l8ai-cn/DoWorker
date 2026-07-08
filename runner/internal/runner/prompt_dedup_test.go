package runner

import "testing"

func TestPromptDedupRing_EvictsOldest(t *testing.T) {
	r := newPromptDedupRing(2)
	r.add("first")
	r.add("second")
	if !r.seen("first") || !r.seen("second") {
		t.Fatal("expected both ids present")
	}
	r.add("third")
	if r.seen("first") {
		t.Fatal("expected first id evicted")
	}
	if !r.seen("third") {
		t.Fatal("expected third id retained")
	}
}
