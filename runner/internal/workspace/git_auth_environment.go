package workspace

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/l8ai-cn/agentcloud/runner/internal/envfilter"
)

func (m *Manager) setGitAuthEnv(cmd *exec.Cmd, opts *WorktreeOptions) {
	env := baseGitEnv(opts)
	if opts != nil && opts.SSHKeyPath != "" {
		env = append(env, "GIT_SSH_COMMAND="+sshKeyCommand(opts.SSHKeyPath, false))
	} else if opts != nil && (opts.AnonymousAuth || opts.GitToken != "") {
		env = append(env, "GIT_SSH_COMMAND="+anonymousSSHCommand(false))
	}
	cmd.Env = env
}

func (m *Manager) setLocalGitEnv(cmd *exec.Cmd) {
	env := removeInheritedGitAuthEnv(envfilter.FilterEnv(os.Environ()))
	env = append(env,
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
		"SSH_ASKPASS=",
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"SSH_AUTH_SOCK=",
		"GIT_SSH_COMMAND="+anonymousSSHCommand(false),
	)
	cmd.Env = appendGitEnvConfig(env, []gitEnvConfig{
		{key: "credential.helper", value: ""},
		{key: "http.extraHeader", value: ""},
	})
}

func (m *Manager) setProbeEnv(cmd *exec.Cmd, opts *WorktreeOptions) {
	env := baseGitEnv(opts)
	if opts != nil && opts.SSHKeyPath != "" {
		env = append(env, "GIT_SSH_COMMAND="+sshKeyCommand(opts.SSHKeyPath, true))
	} else if opts != nil && (opts.AnonymousAuth || opts.GitToken != "") {
		env = append(env, "GIT_SSH_COMMAND="+anonymousSSHCommand(true))
	} else {
		env = append(env, "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no -o BatchMode=yes -o ConnectTimeout=30")
	}
	cmd.Env = env
}

func baseGitEnv(opts *WorktreeOptions) []string {
	env := envfilter.FilterEnv(os.Environ())
	env = append(env, "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=")
	if opts == nil || (!opts.AnonymousAuth && opts.GitToken == "" && opts.SSHKeyPath == "") {
		return env
	}
	env = removeInheritedGitAuthEnv(env)
	env = append(env, "GIT_ASKPASS=", "SSH_ASKPASS=")
	configs := []gitEnvConfig{
		{key: "credential.helper", value: ""},
		{key: "http.extraHeader", value: ""},
	}
	if opts.GitToken != "" {
		configs = append(configs, gitEnvConfig{key: "http.extraHeader", value: gitBasicAuthorization(opts)})
	}
	env = append(env,
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"SSH_AUTH_SOCK=",
	)
	return appendGitEnvConfig(env, configs)
}

func removeInheritedGitAuthEnv(env []string) []string {
	result := env[:0]
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if isGitAuthEnvKey(key) {
			continue
		}
		result = append(result, entry)
	}
	return result
}

func isGitAuthEnvKey(key string) bool {
	normalized := strings.ToUpper(key)
	return normalized == "GIT_CONFIG_GLOBAL" || normalized == "GIT_CONFIG_NOSYSTEM" ||
		normalized == "GIT_CONFIG_COUNT" || normalized == "GIT_SSH_COMMAND" ||
		normalized == "SSH_AUTH_SOCK" || normalized == "GIT_ASKPASS" || normalized == "SSH_ASKPASS" ||
		strings.HasPrefix(normalized, "GIT_CONFIG_KEY_") ||
		strings.HasPrefix(normalized, "GIT_CONFIG_VALUE_")
}

type gitEnvConfig struct {
	key   string
	value string
}

func appendGitEnvConfig(env []string, configs []gitEnvConfig) []string {
	env = append(env, fmt.Sprintf("GIT_CONFIG_COUNT=%d", len(configs)))
	for i, config := range configs {
		env = append(env,
			fmt.Sprintf("GIT_CONFIG_KEY_%d=%s", i, config.key),
			fmt.Sprintf("GIT_CONFIG_VALUE_%d=%s", i, config.value),
		)
	}
	return env
}

func sshKeyCommand(keyPath string, withTimeout bool) string {
	command := fmt.Sprintf(
		`ssh -F %s -o IdentityFile=none -i %s -o IdentitiesOnly=yes -o IdentityAgent=none -o StrictHostKeyChecking=no -o BatchMode=yes`,
		gitSSHQuote(os.DevNull),
		gitSSHQuote(keyPath),
	)
	if withTimeout {
		command += " -o ConnectTimeout=30"
	}
	return command
}

func anonymousSSHCommand(withTimeout bool) string {
	command := fmt.Sprintf("ssh -F %s -o IdentityFile=%s -o IdentitiesOnly=yes -o IdentityAgent=none -o StrictHostKeyChecking=no -o BatchMode=yes",
		gitSSHQuote(os.DevNull), gitSSHQuote(os.DevNull))
	if withTimeout {
		command += " -o ConnectTimeout=30"
	}
	return command
}

func gitSSHQuote(path string) string {
	escaped := strings.ReplaceAll(path, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

func (m *Manager) redactGitOutput(opts *WorktreeOptions, output []byte) string {
	text := strings.TrimSpace(string(output))
	if opts == nil || opts.GitToken == "" {
		return text
	}
	text = strings.ReplaceAll(text, gitBasicAuthorization(opts), "Authorization: Basic [REDACTED]")
	return strings.ReplaceAll(text, opts.GitToken, "[REDACTED]")
}

func gitBasicAuthorization(opts *WorktreeOptions) string {
	username := opts.GitUsername
	if username == "" {
		username = "x-access-token"
	}
	value := base64.StdEncoding.EncodeToString([]byte(username + ":" + opts.GitToken))
	return "Authorization: Basic " + value
}
