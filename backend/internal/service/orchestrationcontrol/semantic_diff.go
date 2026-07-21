package orchestrationcontrol

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
	"strings"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
)

func semanticChanges(
	current *control.ResourceRevision,
	proposed json.RawMessage,
) ([]control.SemanticChange, error) {
	after, err := authoringProjection(proposed)
	if err != nil {
		return nil, err
	}
	if current == nil {
		digest, err := semanticValueDigest(after)
		if err != nil {
			return nil, err
		}
		return []control.SemanticChange{{
			Operation: control.SemanticChangeAdd,
			Path:      "/", After: control.ChangeValue{Digest: digest},
		}}, nil
	}
	before, err := authoringProjection(current.CanonicalManifest)
	if err != nil {
		return nil, err
	}
	changes := make([]control.SemanticChange, 0)
	if err := appendSemanticChanges(&changes, "", before, after); err != nil {
		return nil, err
	}
	sort.Slice(changes, func(left, right int) bool {
		if changes[left].Path != changes[right].Path {
			return changes[left].Path < changes[right].Path
		}
		return changes[left].Operation < changes[right].Operation
	})
	return changes, nil
}

func authoringProjection(raw json.RawMessage) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var document map[string]any
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	delete(document, "status")
	if metadata, ok := document["metadata"].(map[string]any); ok {
		delete(metadata, "uid")
		delete(metadata, "resourceVersion")
		delete(metadata, "generation")
	}
	return document, nil
}

func appendSemanticChanges(
	changes *[]control.SemanticChange,
	path string,
	before, after any,
) error {
	if reflect.DeepEqual(before, after) {
		return nil
	}
	beforeMap, beforeIsMap := before.(map[string]any)
	afterMap, afterIsMap := after.(map[string]any)
	if beforeIsMap && afterIsMap {
		keys := mapKeys(beforeMap, afterMap)
		for _, key := range keys {
			beforeValue, beforeExists := beforeMap[key]
			afterValue, afterExists := afterMap[key]
			childPath := path + "/" + escapeJSONPointer(key)
			if err := appendSemanticEntry(
				changes,
				childPath,
				beforeValue,
				afterValue,
				beforeExists,
				afterExists,
			); err != nil {
				return err
			}
		}
		return nil
	}
	return appendSemanticEntry(changes, normalizedDiffPath(path), before, after, true, true)
}

func appendSemanticEntry(
	changes *[]control.SemanticChange,
	path string,
	before, after any,
	beforeExists, afterExists bool,
) error {
	change := control.SemanticChange{Path: normalizedDiffPath(path)}
	switch {
	case !beforeExists:
		change.Operation = control.SemanticChangeAdd
		digest, err := semanticValueDigest(after)
		if err != nil {
			return err
		}
		change.After.Digest = digest
	case !afterExists:
		change.Operation = control.SemanticChangeRemove
		digest, err := semanticValueDigest(before)
		if err != nil {
			return err
		}
		change.Before.Digest = digest
	default:
		if beforeMap, ok := before.(map[string]any); ok {
			if afterMap, matches := after.(map[string]any); matches {
				return appendSemanticChanges(changes, path, beforeMap, afterMap)
			}
		}
		change.Operation = control.SemanticChangeReplace
		beforeDigest, err := semanticValueDigest(before)
		if err != nil {
			return err
		}
		afterDigest, err := semanticValueDigest(after)
		if err != nil {
			return err
		}
		change.Before.Digest = beforeDigest
		change.After.Digest = afterDigest
	}
	*changes = append(*changes, change)
	return nil
}

func semanticValueDigest(value any) (string, error) {
	return control.DigestCanonicalJSON(map[string]any{"value": value})
}

func mapKeys(left, right map[string]any) []string {
	keys := make(map[string]struct{}, len(left)+len(right))
	for key := range left {
		keys[key] = struct{}{}
	}
	for key := range right {
		keys[key] = struct{}{}
	}
	sorted := make([]string, 0, len(keys))
	for key := range keys {
		sorted = append(sorted, key)
	}
	sort.Strings(sorted)
	return sorted
}

func escapeJSONPointer(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "~", "~0"), "/", "~1")
}

func normalizedDiffPath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}
