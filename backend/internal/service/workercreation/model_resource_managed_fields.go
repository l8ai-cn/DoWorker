package workercreation

import "sort"

var modelResourceManagedTypeFields = map[string]map[string]struct{}{
	"codex-cli": {
		"model":           {},
		"OPENAI_API_KEY":  {},
		"OPENAI_BASE_URL": {},
		"OPENAI_MODEL":    {},
	},
	"claude-code": {
		"model":                {},
		"ANTHROPIC_API_KEY":    {},
		"ANTHROPIC_AUTH_TOKEN": {},
		"ANTHROPIC_BASE_URL":   {},
	},
	"gemini-cli": {
		"model":          {},
		"GEMINI_API_KEY": {},
		"GOOGLE_API_KEY": {},
	},
}

func isModelResourceManagedTypeField(workerType, field string) bool {
	fields := modelResourceManagedTypeFields[workerType]
	_, exists := fields[field]
	return exists
}

func modelResourceManagedRuntimeField(
	workerType string,
	values map[string]string,
) string {
	fields := make([]string, 0, len(values))
	for field := range values {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		if isModelResourceManagedTypeField(workerType, field) {
			return field
		}
	}
	return ""
}
