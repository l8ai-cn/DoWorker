package runner

import (
	"os"
	"strings"
)

func gitProcessIsolationEnv(credentialType string) map[string]string {
	switch credentialType {
	case "none", "oauth", "pat", "ssh_key":
		return map[string]string{
			"GIT_CONFIG_GLOBAL":   os.DevNull,
			"GIT_CONFIG_NOSYSTEM": "1",
			"GIT_ASKPASS":         "",
			"SSH_ASKPASS":         "",
			"SSH_AUTH_SOCK":       "",
		}
	default:
		return nil
	}
}

func gitProcessIsolationUnsetEnv(credentialType string) []string {
	if !usesExplicitGitProcessIsolation(credentialType) {
		return nil
	}
	names := []string{
		"GIT_CONFIG_GLOBAL",
		"GIT_CONFIG_NOSYSTEM",
		"GIT_CONFIG_COUNT",
		"GIT_SSH_COMMAND",
		"GIT_ASKPASS",
		"SSH_ASKPASS",
		"SSH_AUTH_SOCK",
	}
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		seen[name] = struct{}{}
	}
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if isGitProcessCredentialEnv(name) {
			if _, found := seen[name]; !found {
				names = append(names, name)
				seen[name] = struct{}{}
			}
		}
	}
	return names
}

func enforceGitProcessIsolation(env map[string]string, credentialType string) {
	if !usesExplicitGitProcessIsolation(credentialType) {
		return
	}
	for key := range env {
		if isGitProcessCredentialEnv(key) {
			delete(env, key)
		}
	}
	for key, value := range gitProcessIsolationEnv(credentialType) {
		env[key] = value
	}
}

func usesExplicitGitProcessIsolation(credentialType string) bool {
	switch credentialType {
	case "none", "oauth", "pat", "ssh_key":
		return true
	default:
		return false
	}
}

func isGitProcessCredentialEnv(key string) bool {
	normalized := strings.ToUpper(key)
	switch normalized {
	case "GIT_CONFIG_GLOBAL", "GIT_CONFIG_NOSYSTEM", "GIT_CONFIG_COUNT",
		"GIT_SSH_COMMAND", "GIT_ASKPASS", "SSH_ASKPASS", "SSH_AUTH_SOCK":
		return true
	default:
		return strings.HasPrefix(normalized, "GIT_CONFIG_KEY_") ||
			strings.HasPrefix(normalized, "GIT_CONFIG_VALUE_")
	}
}
