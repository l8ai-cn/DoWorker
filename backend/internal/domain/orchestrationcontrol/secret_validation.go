package orchestrationcontrol

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func rejectRawSecretJSON(raw json.RawMessage) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return invalid("json", "must be valid")
	}
	if containsRawSecret("", value) {
		return invalid("json", "must not contain raw secret-like data")
	}
	return nil
}

func containsRawSecret(key string, value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		if isSensitiveKey(key) {
			return !isReferenceIdentityObject(typed)
		}
		for childKey, child := range typed {
			if containsRawSecret(childKey, child) {
				return true
			}
		}
	case []any:
		if isSensitiveKey(key) {
			return true
		}
		for _, child := range typed {
			if containsRawSecret(key, child) {
				return true
			}
		}
	case string:
		return looksLikeStrongSecret(typed) ||
			(isSensitiveKey(key) && !isRedactedValue(typed))
	case json.Number, bool, nil:
		return isSensitiveKey(key)
	}
	return false
}

func isReferenceIdentityObject(value map[string]any) bool {
	kind, kindOK := value["kind"].(string)
	if !kindOK || !validReferenceIdentityKind(kind) {
		return false
	}
	if name, exists := value["name"]; exists {
		return validNamedReferenceIdentity(value, name)
	}
	id, exists := value["id"]
	return exists && validIDReferenceIdentity(value, id)
}

func validNamedReferenceIdentity(value map[string]any, name any) bool {
	nameValue, ok := name.(string)
	if !ok || slugkit.Validate(nameValue) != nil {
		return false
	}
	allowed := map[string]struct{}{
		"apiVersion": {}, "kind": {}, "namespace": {}, "name": {},
		"uid": {}, "revision": {}, "digest": {},
	}
	for key, field := range value {
		if _, ok := allowed[key]; !ok || !validReferenceIdentityField(key, field) {
			return false
		}
	}
	return true
}

func validIDReferenceIdentity(value map[string]any, id any) bool {
	if len(value) != 2 {
		return false
	}
	number, ok := id.(json.Number)
	if !ok {
		return false
	}
	parsed, err := strconv.ParseInt(number.String(), 10, 64)
	return err == nil && parsed > 0
}

func validReferenceIdentityKind(value string) bool {
	if slugkit.Validate(value) == nil {
		return true
	}
	return (orchestrationresource.TypeMeta{
		APIVersion: orchestrationresource.APIVersionV1Alpha1,
		Kind:       value,
	}).Validate() == nil
}

func validReferenceIdentityField(key string, value any) bool {
	switch key {
	case "apiVersion", "kind", "uid", "digest":
		_, ok := value.(string)
		return ok
	case "namespace", "name":
		text, ok := value.(string)
		return ok && (text == "" || slugkit.Validate(text) == nil)
	case "revision":
		number, ok := value.(json.Number)
		if !ok {
			return false
		}
		parsed, err := strconv.ParseInt(number.String(), 10, 64)
		return err == nil && parsed >= 0
	default:
		return false
	}
}

func isSensitiveKey(value string) bool {
	normalized := strings.Map(func(char rune) rune {
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			return char
		}
		if char >= 'A' && char <= 'Z' {
			return char + ('a' - 'A')
		}
		return -1
	}, value)
	if strings.HasSuffix(normalized, "ref") ||
		strings.HasSuffix(normalized, "refs") ||
		strings.HasSuffix(normalized, "reference") ||
		strings.HasSuffix(normalized, "references") {
		return false
	}
	for _, marker := range []string{
		"password", "passwd", "secret", "token", "privatekey",
		"clientsecret", "credential", "apikey", "accesskey",
	} {
		if normalized == marker || strings.HasSuffix(normalized, marker) {
			return true
		}
	}
	return false
}

func containsSecretLikeText(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{
		"password=", "passwd=", "secret=", "token=", "api_key=",
		"apikey=", "private_key=", "client_secret=",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return looksLikeStrongSecret(value)
}

func looksLikeStrongSecret(value string) bool {
	trimmed := strings.TrimSpace(value)
	return strings.HasPrefix(trimmed, "sk-") ||
		strings.HasPrefix(trimmed, "ghp_") ||
		strings.HasPrefix(trimmed, "github_pat_") ||
		strings.Contains(trimmed, "BEGIN PRIVATE KEY")
}

func isRedactedValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "[redacted]", "<redacted>", "***", "redacted":
		return true
	default:
		return false
	}
}
