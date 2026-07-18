package workerconfigmigration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func decodeObject(raw []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil || object == nil {
		return nil, fmt.Errorf("must contain a JSON object")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("must contain one JSON object")
	}
	return object, nil
}

func requiredObject(document map[string]any, path ...string) (map[string]any, error) {
	current := document
	for _, part := range path {
		value, found := current[part]
		if !found {
			return nil, fmt.Errorf("%s is required", part)
		}
		next, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s must be an object", part)
		}
		current = next
	}
	return current, nil
}

func requiredString(document map[string]any, path ...string) (string, error) {
	if len(path) == 0 {
		return "", fmt.Errorf("string path is empty")
	}
	parent, err := requiredObject(document, path[:len(path)-1]...)
	if err != nil {
		return "", err
	}
	value, ok := parent[path[len(path)-1]].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("%s must be a non-empty string", path[len(path)-1])
	}
	return value, nil
}

func requiredSlice(value any, field string) ([]any, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an array", field)
	}
	return items, nil
}

func positiveInt64(value any, field string) (int64, error) {
	number, ok := value.(json.Number)
	if !ok {
		return 0, fmt.Errorf("%s must contain integer bundle IDs", field)
	}
	parsed, err := number.Int64()
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must contain positive bundle IDs", field)
	}
	return parsed, nil
}

func hasLegacyTemplateConfig(raw []byte) (bool, error) {
	document, err := decodeObject(raw)
	if err != nil {
		return false, err
	}
	workspace, err := requiredObject(document, "spec", "workspace")
	if err != nil {
		return false, err
	}
	_, found := workspace["configBundleRefs"]
	return found, nil
}
