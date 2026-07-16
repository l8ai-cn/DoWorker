package codex

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxToolOutputBytes        = 512 << 10
	toolOutputTruncatedMarker = "\n[Runner truncated tool output]\n"
)

type toolOutputBuffer struct {
	text      string
	truncated bool
}

func commandArgumentsJSON(item codexItem) string {
	arguments := map[string]any{}
	if command, ok := rawJSONValue(item.Command); ok {
		arguments["command"] = command
	}
	if item.CWD != "" {
		arguments["cwd"] = item.CWD
	}
	return marshalArguments(arguments)
}

func fileChangeArgumentsJSON(changes []fileUpdateChange) string {
	if len(changes) == 0 {
		return ""
	}
	return marshalArguments(map[string]any{"changes": changes})
}

func rawArgumentsJSON(value json.RawMessage) string {
	parsed, ok := rawJSONValue(value)
	if !ok {
		return ""
	}
	return marshalArguments(parsed)
}

func rawJSONValue(value json.RawMessage) (any, bool) {
	if len(value) == 0 || string(value) == "null" {
		return nil, false
	}
	var parsed any
	if err := json.Unmarshal(value, &parsed); err != nil {
		return string(value), true
	}
	return parsed, true
}

func marshalArguments(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func (t *transport) appendToolOutput(itemID, delta string) {
	if itemID == "" || delta == "" {
		return
	}
	t.toolMu.Lock()
	current := t.toolOutputs[itemID]
	if current.truncated {
		t.toolMu.Unlock()
		return
	}
	remaining := maxToolOutputBytes - len(current.text)
	if len(delta) > remaining {
		current.text += validUTF8Prefix(delta, remaining)
		current.truncated = true
	} else {
		current.text += delta
	}
	t.toolOutputs[itemID] = current
	t.toolMu.Unlock()
}

func (t *transport) takeToolOutput(itemID string) string {
	t.toolMu.Lock()
	defer t.toolMu.Unlock()
	buffer := t.toolOutputs[itemID]
	delete(t.toolOutputs, itemID)
	if buffer.truncated {
		return buffer.text + toolOutputTruncatedMarker
	}
	return buffer.text
}

func (t *transport) clearToolOutputs() {
	t.toolMu.Lock()
	clear(t.toolOutputs)
	t.toolMu.Unlock()
}

func validUTF8Prefix(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if len(value) <= limit {
		return value
	}
	prefix := value[:limit]
	for !utf8.ValidString(prefix) {
		prefix = prefix[:len(prefix)-1]
	}
	return prefix
}

func (t *transport) commandOutput(item codexItem) string {
	streamed := t.takeToolOutput(item.ID)
	if item.AggregatedOutput != "" {
		return item.AggregatedOutput
	}
	return streamed
}

func (t *transport) fileChangeOutput(item codexItem) string {
	if output := t.takeToolOutput(item.ID); output != "" {
		return output
	}
	if len(item.Changes) == 0 && item.FilePath != "" {
		return item.FilePath
	}
	lines := make([]string, 0, len(item.Changes))
	for _, change := range item.Changes {
		if change.Path == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s %s", fileChangeKind(change.Kind), change.Path))
	}
	return strings.Join(lines, "\n")
}

func fileChangeKind(value json.RawMessage) string {
	var kind struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(value, &kind) != nil || kind.Type == "" {
		return "Changed"
	}
	return strings.ToUpper(kind.Type[:1]) + kind.Type[1:]
}
