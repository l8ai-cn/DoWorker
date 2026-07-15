package orchestrationcontrol

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func CanonicalJSON(value any) ([]byte, error) {
	return canonicalJSON(value, 0)
}

func CanonicalJSONObject(value any) ([]byte, error) {
	return canonicalJSON(value, '{')
}

func CanonicalJSONArray(value any) ([]byte, error) {
	return canonicalJSON(value, '[')
}

func DigestCanonicalJSON(value any) (string, error) {
	canonical, err := CanonicalJSON(value)
	if err != nil {
		return "", err
	}
	return digestBytes(canonical), nil
}

func canonicalJSON(value any, expected json.Delim) ([]byte, error) {
	raw, err := marshalCanonicalInput(value)
	if err != nil {
		return nil, invalid("json", "cannot be encoded")
	}
	decoded, root, err := decodeJSONValue(raw)
	if err != nil {
		return nil, invalid("json", "must contain one valid JSON value")
	}
	if root != '{' && root != '[' {
		return nil, invalid("json", "root must be an object or array")
	}
	if expected != 0 && root != expected {
		return nil, invalid("json", "root has the wrong container shape")
	}
	if err := normalizeJSONNumbers(decoded); err != nil {
		return nil, invalid("json", "contains an unsupported number")
	}
	canonical, err := json.Marshal(decoded)
	if err != nil {
		return nil, invalid("json", "cannot be encoded")
	}
	if len(canonical) > maxCanonicalJSONBytes {
		return nil, invalid("json", "canonical form is too large")
	}
	return canonical, nil
}

func canonicalAnyJSON(raw json.RawMessage) ([]byte, error) {
	decoded, _, err := decodeJSONValue(raw)
	if err != nil {
		return nil, invalid("json", "must contain one valid JSON value")
	}
	if err := normalizeJSONNumbers(decoded); err != nil {
		return nil, invalid("json", "contains an unsupported number")
	}
	canonical, err := json.Marshal(decoded)
	if err != nil {
		return nil, invalid("json", "cannot be encoded")
	}
	if len(canonical) > maxCanonicalJSONBytes {
		return nil, invalid("json", "canonical form is too large")
	}
	return canonical, nil
}

func marshalCanonicalInput(value any) ([]byte, error) {
	switch raw := value.(type) {
	case json.RawMessage:
		return bytes.Clone(raw), nil
	case []byte:
		return bytes.Clone(raw), nil
	default:
		return json.Marshal(value)
	}
}

func decodeJSONValue(raw []byte) (any, json.Delim, error) {
	if err := validateCanonicalJSONStructure(raw); err != nil {
		return nil, 0, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return nil, 0, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, 0, errors.New("trailing JSON data")
	}
	switch decoded.(type) {
	case map[string]any:
		return decoded, '{', nil
	case []any:
		return decoded, '[', nil
	default:
		return decoded, 0, nil
	}
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return fmt.Sprintf("sha256:%x", sum)
}
