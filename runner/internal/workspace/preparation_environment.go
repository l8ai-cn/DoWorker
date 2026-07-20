package workspace

import "strings"

func removePreparationEnvironmentVariables(env, names []string) []string {
	if len(names) == 0 {
		return env
	}
	excluded := make(map[string]struct{}, len(names))
	for _, name := range names {
		excluded[name] = struct{}{}
	}
	result := env[:0]
	for _, entry := range env {
		name, _, _ := strings.Cut(entry, "=")
		if _, found := excluded[name]; !found {
			result = append(result, entry)
		}
	}
	return result
}
