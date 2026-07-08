// Package codeximport converts a local Codex conversation record into the
// flattened conversation-item payloads persisted by the sessions store
// (backend/internal/domain/conversationitem). It is deliberately free of any
// database, HTTP, or gorm dependency so it can be unit-tested in isolation and
// reused by CLI tooling. The HTTP layer (session_import.go) wires the result
// into a real Worker session.
//
// Two source shapes are auto-detected (see Detect):
//
//   - KindRollout: a Codex CLI/Desktop "rollout" transcript, one JSON object
//     per line, living at ~/.codex/sessions/<yyyy>/<mm>/<dd>/rollout-*.jsonl.
//     This is the real turn-by-turn conversation record.
//   - KindOutputDir: a workflow output_* directory (conversation_input.json +
//     run_manifest.json). It is not a turn-by-turn transcript, so we synthesize
//     a short request/result conversation from its manifest.
package codeximport

// Kind identifies the detected Codex source format.
type Kind string

const (
	// KindRollout is a Codex rollout transcript (rollout-*.jsonl).
	KindRollout Kind = "codex_rollout"
	// KindOutputDir is a workflow output_* directory.
	KindOutputDir Kind = "codex_output_dir"
)

// Item is one normalized conversation item ready to be persisted. Payload is
// the flattened conversation-item payload WITHOUT the store-assigned id,
// response_id, and position — the caller fills those in at insert time so ids
// stay unique per destination session and turns group correctly.
type Item struct {
	// Type is the conversation_items.item_type (e.g. "message",
	// "function_call", "function_call_output", "image_generation_call").
	Type string
	// Status is the conversation_items.status; defaults to "completed".
	Status string
	// StartsTurn marks the first item of a new turn (a user message). The
	// caller mints a fresh response_id whenever this is true so a turn's
	// assistant reply and tool calls share the user prompt's response_id.
	StartsTurn bool
	// Payload is the flattened item payload. The caller injects "id",
	// "response_id", and "status" before serializing.
	Payload map[string]any
}

// Result is the outcome of converting one Codex source.
type Result struct {
	// Kind is the detected source format.
	Kind Kind
	// SourceID is the Codex session id (rollout) or directory basename.
	SourceID string
	// SourcePath is the concrete file/dir that was read.
	SourcePath string
	// Title is a human-readable title derived from the first user prompt (or
	// the workflow target), suitable for the session row.
	Title string
	// Items are the normalized conversation items in chronological order.
	Items []Item
}
