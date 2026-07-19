package workerdependency

import (
	"encoding/json"
	"net/url"
	"strings"
	"unicode"
)

func isSensitiveFieldName(value string) bool {
	normalized := strings.Map(func(character rune) rune {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			return unicode.ToLower(character)
		}
		return -1
	}, value)
	for _, marker := range []string{
		"password", "passwd", "secret", "token", "privatekey",
		"clientsecret", "credential", "apikey", "accesskey", "authkey",
	} {
		if normalized == marker || strings.HasSuffix(normalized, marker) {
			return true
		}
	}
	return false
}

func containsRawSecretText(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{
		"password=", "passwd=", "secret=", "token=", "api_key=",
		"apikey=", "private_key=", "client_secret=",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	for _, marker := range []string{
		"sk-", "ghp_", "github_pat_", "begin private key",
	} {
		if containsSecretTokenPrefix(lower, marker) {
			return true
		}
	}
	return containsSensitiveJSON(value)
}

func containsSecretTokenPrefix(value, marker string) bool {
	index := 0
	for {
		position := strings.Index(value[index:], marker)
		if position < 0 {
			return false
		}
		absolute := index + position
		if absolute == 0 || !isTokenCharacter(rune(value[absolute-1])) {
			return true
		}
		index = absolute + len(marker)
	}
}

func isTokenCharacter(character rune) bool {
	return unicode.IsLetter(character) ||
		unicode.IsDigit(character) ||
		character == '_'
}

func containsURLUserInfo(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.User != nil
}

func containsSensitiveJSON(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || (trimmed[0] != '{' && trimmed[0] != '[') {
		return false
	}
	var decoded any
	if json.Unmarshal([]byte(trimmed), &decoded) != nil {
		return false
	}
	return containsSensitiveJSONValue("", decoded)
}

func containsSensitiveJSONValue(key string, value any) bool {
	if isSensitiveFieldName(key) {
		return true
	}
	switch typed := value.(type) {
	case map[string]any:
		for childKey, child := range typed {
			if containsSensitiveJSONValue(childKey, child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if containsSensitiveJSONValue("", child) {
				return true
			}
		}
	case string:
		return containsRawSecretText(typed)
	}
	return false
}
