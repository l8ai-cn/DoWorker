package codeximport

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// rolloutLine is the envelope of one line in a rollout-*.jsonl transcript.
type rolloutLine struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// convertRollout parses a Codex rollout transcript into normalized items.
func convertRollout(path string) (*Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("codeximport: open rollout: %w", err)
	}
	defer f.Close()

	res := &Result{Kind: KindRollout, SourcePath: path}
	// json.Decoder handles newline-delimited JSON with no per-line size cap,
	// which matters because some lines (base_instructions, image results) are
	// multi-megabyte.
	dec := json.NewDecoder(f)
	for {
		var line rolloutLine
		if err := dec.Decode(&line); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("codeximport: decode rollout line: %w", err)
		}
		switch line.Type {
		case "session_meta":
			var meta struct {
				SessionID string `json:"session_id"`
			}
			_ = json.Unmarshal(line.Payload, &meta)
			if meta.SessionID != "" {
				res.SourceID = meta.SessionID
			}
		case "response_item":
			if item, ok := mapResponseItem(line.Payload); ok {
				res.Items = append(res.Items, item)
			}
		default:
			// event_msg, turn_context, compacted, etc. carry no persistable
			// transcript content — skip.
		}
	}

	res.Title = deriveRolloutTitle(res.Items, res.SourceID)
	return res, nil
}

// mapResponseItem converts one Codex response_item payload into a normalized
// Item. The second return is false when the item carries nothing worth
// persisting (encrypted reasoning, injected developer/system messages, empty
// content).
func mapResponseItem(raw json.RawMessage) (Item, bool) {
	var head struct {
		Type string `json:"type"`
		Role string `json:"role"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return Item{}, false
	}

	switch head.Type {
	case "message":
		return mapMessage(raw, head.Role)
	case "function_call":
		return mapFunctionCall(raw)
	case "custom_tool_call":
		return mapCustomToolCall(raw)
	case "function_call_output", "custom_tool_call_output":
		return mapFunctionCallOutput(raw)
	case "image_generation_call":
		return mapImageGenerationCall(raw)
	case "reasoning":
		// Codex reasoning items are encrypted/opaque with empty summaries —
		// nothing renderable.
		return Item{}, false
	default:
		return Item{}, false
	}
}

func mapMessage(raw json.RawMessage, role string) (Item, bool) {
	// Injected system prompt / permission scaffolding rides on the
	// "developer" role and is not part of the visible conversation.
	if role == "developer" || role == "system" {
		return Item{}, false
	}
	var msg struct {
		Content []map[string]any `json:"content"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return Item{}, false
	}
	content := sanitizeContent(msg.Content)
	if len(content) == 0 {
		return Item{}, false
	}
	normRole := role
	if normRole != "assistant" {
		normRole = "user"
	}
	payload := map[string]any{
		"type":    "message",
		"role":    normRole,
		"content": content,
	}
	// Injected context (AGENTS.md, environment/permissions blocks) rides on the
	// user role but is not a real prompt. Flag it is_meta so the web renderer
	// hides it from the main flow and it never anchors the history window.
	meta := normRole == "user" && isScaffolding(firstBlockText(content))
	if meta {
		payload["is_meta"] = true
	}
	return Item{
		Type:       "message",
		Status:     "completed",
		StartsTurn: normRole == "user" && !meta,
		Payload:    payload,
	}, true
}

// isScaffolding reports whether a user message is injected agent context rather
// than a genuine user prompt.
func isScaffolding(text string) bool {
	t := strings.TrimSpace(text)
	if t == "" {
		return true
	}
	prefixes := []string{
		"# AGENTS.md instructions",
		"<environment_context>",
		"<permissions instructions>",
		"<app-context>",
		"<collaboration_mode>",
		"<skills_instructions>",
		"<plugins_instructions>",
		"<user_instructions>",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	return false
}

func firstBlockText(content []map[string]any) string {
	for _, b := range content {
		if txt, ok := b["text"].(string); ok && txt != "" {
			return txt
		}
	}
	return ""
}

// allowedBlockTypes are the message content block types the web renderer
// understands (see clients/web-user/src/lib/blocks.ts MessageContentBlock).
var allowedBlockTypes = map[string]bool{
	"input_text":  true,
	"output_text": true,
	"input_image": true,
	"input_file":  true,
}

// sanitizeContent keeps only renderable content blocks and strips transport
// noise. A bare "text" block (rare) is normalized to input_text.
func sanitizeContent(blocks []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(blocks))
	for _, b := range blocks {
		t, _ := b["type"].(string)
		switch {
		case t == "text":
			if txt, ok := b["text"].(string); ok && txt != "" {
				out = append(out, map[string]any{"type": "input_text", "text": txt})
			}
		case t == "input_text" || t == "output_text":
			if txt, ok := b["text"].(string); ok && txt != "" {
				out = append(out, map[string]any{"type": t, "text": txt})
			}
		case allowedBlockTypes[t]:
			clean := map[string]any{"type": t}
			if v, ok := b["file_id"]; ok {
				clean["file_id"] = v
			}
			if v, ok := b["filename"]; ok {
				clean["filename"] = v
			}
			out = append(out, clean)
		}
	}
	return out
}

func mapFunctionCall(raw json.RawMessage) (Item, bool) {
	var fc struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
		CallID    string `json:"call_id"`
		ID        string `json:"id"`
	}
	if err := json.Unmarshal(raw, &fc); err != nil || fc.Name == "" {
		return Item{}, false
	}
	callID := firstNonEmpty(fc.CallID, fc.ID)
	return Item{
		Type:   "function_call",
		Status: "completed",
		Payload: map[string]any{
			"type":      "function_call",
			"name":      fc.Name,
			"arguments": fc.Arguments,
			"call_id":   callID,
		},
	}, true
}

// mapCustomToolCall maps Codex custom tools (e.g. apply_patch) onto the
// generic function_call item shape so they render as tool calls. Codex carries
// the argument text in "input" rather than "arguments".
func mapCustomToolCall(raw json.RawMessage) (Item, bool) {
	var ct struct {
		Name   string `json:"name"`
		Input  string `json:"input"`
		CallID string `json:"call_id"`
		ID     string `json:"id"`
	}
	if err := json.Unmarshal(raw, &ct); err != nil || ct.Name == "" {
		return Item{}, false
	}
	callID := firstNonEmpty(ct.CallID, ct.ID)
	return Item{
		Type:   "function_call",
		Status: "completed",
		Payload: map[string]any{
			"type":      "function_call",
			"name":      ct.Name,
			"arguments": ct.Input,
			"call_id":   callID,
		},
	}, true
}

func mapFunctionCallOutput(raw json.RawMessage) (Item, bool) {
	var fo struct {
		CallID string          `json:"call_id"`
		Output json.RawMessage `json:"output"`
	}
	if err := json.Unmarshal(raw, &fo); err != nil || fo.CallID == "" {
		return Item{}, false
	}
	return Item{
		Type:   "function_call_output",
		Status: "completed",
		Payload: map[string]any{
			"type":    "function_call_output",
			"call_id": fo.CallID,
			"output":  outputToString(fo.Output),
		},
	}, true
}

// mapImageGenerationCall keeps the provider-native image item but drops the
// (often multi-megabyte) base64 "result" blob so persisted payloads stay
// small; the revised prompt and status remain for display.
func mapImageGenerationCall(raw json.RawMessage) (Item, bool) {
	var ig struct {
		ID            string `json:"id"`
		Status        string `json:"status"`
		RevisedPrompt string `json:"revised_prompt"`
	}
	if err := json.Unmarshal(raw, &ig); err != nil {
		return Item{}, false
	}
	payload := map[string]any{"type": "image_generation_call"}
	if ig.RevisedPrompt != "" {
		payload["revised_prompt"] = ig.RevisedPrompt
	}
	status := ig.Status
	if status == "" {
		status = "completed"
	}
	payload["status"] = status
	return Item{
		Type:    "image_generation_call",
		Status:  status,
		Payload: payload,
	}, true
}

// outputToString coerces a function-call output (which may be a JSON string or
// an object) into a plain string for the function_call_output item.
func outputToString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

// deriveRolloutTitle uses the first visible user prompt as the session title,
// trimmed to a single short line. Falls back to the Codex session id.
func deriveRolloutTitle(items []Item, sourceID string) string {
	for _, it := range items {
		if it.Type != "message" {
			continue
		}
		if role, _ := it.Payload["role"].(string); role != "user" {
			continue
		}
		if meta, _ := it.Payload["is_meta"].(bool); meta {
			continue
		}
		if text := firstText(it.Payload); text != "" {
			return truncateTitle(extractUserRequest(text))
		}
	}
	if sourceID != "" {
		return "Codex " + sourceID
	}
	return "Imported Codex conversation"
}

// extractUserRequest pulls the human ask out of a Codex desktop user message
// that wraps the prompt under a "My request for Codex:" heading (the rest is
// injected file context). Returns the input unchanged when no marker is found.
func extractUserRequest(text string) string {
	for _, marker := range []string{"## My request for Codex:", "My request for Codex:"} {
		if i := strings.LastIndex(text, marker); i >= 0 {
			rest := strings.TrimSpace(text[i+len(marker):])
			if rest != "" {
				return rest
			}
		}
	}
	return text
}

func firstText(payload map[string]any) string {
	blocks, _ := payload["content"].([]map[string]any)
	for _, b := range blocks {
		if txt, ok := b["text"].(string); ok && strings.TrimSpace(txt) != "" {
			return txt
		}
	}
	return ""
}

func truncateTitle(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	const max = 80
	if len([]rune(s)) > max {
		return string([]rune(s)[:max]) + "…"
	}
	return s
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
