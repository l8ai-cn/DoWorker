package runner

import "strings"

func removeKnowledgeMountInheritedGitEnv(env []string) []string {
	result := env[:0]
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		normalized := strings.ToUpper(key)
		if normalized == "GIT_CONFIG_GLOBAL" || normalized == "GIT_CONFIG_NOSYSTEM" ||
			normalized == "GIT_CONFIG_COUNT" || normalized == "GIT_SSH_COMMAND" ||
			normalized == "SSH_AUTH_SOCK" || normalized == "GIT_ASKPASS" || normalized == "SSH_ASKPASS" ||
			strings.HasPrefix(normalized, "GIT_CONFIG_KEY_") ||
			strings.HasPrefix(normalized, "GIT_CONFIG_VALUE_") {
			continue
		}
		result = append(result, entry)
	}
	return result
}
