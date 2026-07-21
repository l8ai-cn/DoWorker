package runner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeEnvVarsEnforcesExplicitGitIsolation(t *testing.T) {
	builder := gitBuilder(nil, &runnerv1.SandboxConfig{CredentialType: "oauth"})
	builder.cmd.EnvVars = map[string]string{
		"GIT_CONFIG_GLOBAL":   "/runner/global.gitconfig",
		"GIT_CONFIG_NOSYSTEM": "0",
		"GIT_CONFIG_COUNT":    "1",
		"GIT_CONFIG_KEY_0":    "http.extraHeader",
		"GIT_CONFIG_VALUE_0":  "Authorization: Basic command-secret",
		"git_config_key_7":    "credential.helper",
		"git_config_value_7":  "store --file=/runner/lowercase-credentials",
		"GIT_SSH_COMMAND":     "ssh -i /runner/key",
		"SSH_AUTH_SOCK":       "/runner/agent.sock",
	}

	env := builder.mergeEnvVars("")

	assert.Equal(t, os.DevNull, env["GIT_CONFIG_GLOBAL"])
	assert.Equal(t, "1", env["GIT_CONFIG_NOSYSTEM"])
	assert.Empty(t, env["SSH_AUTH_SOCK"])
	assert.Empty(t, env["GIT_ASKPASS"])
	assert.Empty(t, env["SSH_ASKPASS"])
	assert.NotContains(t, env, "GIT_SSH_COMMAND")
	assert.NotContains(t, env, "GIT_CONFIG_COUNT")
	assert.NotContains(t, env, "GIT_CONFIG_KEY_0")
	assert.NotContains(t, env, "GIT_CONFIG_VALUE_0")
	assert.NotContains(t, env, "git_config_key_7")
	assert.NotContains(t, env, "git_config_value_7")
}

func TestMergeEnvVarsKeepsRunnerLocalGitEnvironment(t *testing.T) {
	builder := gitBuilder(nil, &runnerv1.SandboxConfig{CredentialType: "runner_local"})
	builder.cmd.EnvVars = map[string]string{"GIT_CONFIG_GLOBAL": "/runner/global.gitconfig"}

	env := builder.mergeEnvVars("")

	assert.Equal(t, "/runner/global.gitconfig", env["GIT_CONFIG_GLOBAL"])
}

func TestPreparationScriptReceivesExplicitGitIsolation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell environment assertion requires a POSIX shell")
	}
	workspacePath := t.TempDir()
	outputPath := filepath.Join(workspacePath, "git-env.txt")
	t.Setenv("GIT_SSH_COMMAND", "ssh -i /runner/key")
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "http.extraHeader")
	t.Setenv("GIT_CONFIG_VALUE_0", "Authorization: Basic inherited-secret")
	builder := gitBuilder(nil, &runnerv1.SandboxConfig{
		CredentialType: "ssh_key",
		PreparationScript: "printf '%s|%s|%s|%s|%s' " +
			"\"$GIT_CONFIG_GLOBAL\" \"$GIT_CONFIG_NOSYSTEM\" \"${GIT_SSH_COMMAND-unset}\" " +
			"\"${GIT_CONFIG_COUNT-unset}\" \"${GIT_CONFIG_KEY_0-unset}\" > git-env.txt",
	})

	require.NoError(t, builder.runPreparationScript(
		context.Background(),
		builder.cmd.SandboxConfig,
		workspacePath,
		"main",
	))

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	parts := strings.Split(string(content), "|")
	require.Len(t, parts, 5)
	assert.Equal(t, os.DevNull, parts[0])
	assert.Equal(t, "1", parts[1])
	assert.Equal(t, "unset", parts[2])
	assert.Equal(t, "unset", parts[3])
	assert.Equal(t, "unset", parts[4])
}

func TestBuildPreventsCommandGitEnvironmentOverride(t *testing.T) {
	t.Setenv("GIT_SSH_COMMAND", "ssh -i /runner/inherited")
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "credential.helper")
	t.Setenv("GIT_CONFIG_VALUE_0", "store --file=/runner/credentials")
	t.Setenv("git_config_key_9", "core.sshCommand")
	t.Setenv("git_config_value_9", "ssh -i /runner/lowercase")
	for _, mode := range []string{InteractionModePTY, InteractionModeACP} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			builder := NewPodBuilder(PodBuilderDeps{Config: &config.Config{WorkspaceRoot: root}})
			builder.WithCommand(&runnerv1.CreatePodCommand{
				PodKey:          "git-env-" + mode,
				LaunchCommand:   "echo",
				InteractionMode: mode,
				EnvVars: map[string]string{
					"GIT_CONFIG_GLOBAL":  "{{sandbox_root}}/malicious.gitconfig",
					"GIT_CONFIG_COUNT":   "1",
					"GIT_CONFIG_KEY_0":   "http.extraHeader",
					"GIT_CONFIG_VALUE_0": "Authorization: Basic command-secret",
					"GIT_SSH_COMMAND":    "ssh -i /command/key",
					"git_config_key_7":   "credential.helper",
					"git_config_value_7": "store --file=/command/credentials",
				},
				SandboxConfig: &runnerv1.SandboxConfig{CredentialType: "oauth"},
			})

			pod, err := builder.Build(context.Background())
			require.NoError(t, err)
			defer pod.workspace.Close()

			env := envSliceMap(pod.LaunchEnv)
			assert.Equal(t, os.DevNull, env["GIT_CONFIG_GLOBAL"])
			for key := range env {
				normalized := strings.ToUpper(key)
				assert.NotEqual(t, "GIT_CONFIG_COUNT", normalized)
				assert.NotEqual(t, "GIT_SSH_COMMAND", normalized)
				assert.False(t, strings.HasPrefix(normalized, "GIT_CONFIG_KEY_"), key)
				assert.False(t, strings.HasPrefix(normalized, "GIT_CONFIG_VALUE_"), key)
			}
		})
	}
}

func TestGitProcessCredentialEnvMatchesCaseInsensitively(t *testing.T) {
	for _, key := range []string{
		"git_config_count",
		"Git_Config_Key_0",
		"gIt_CoNfIg_VaLuE_0",
		"git_ssh_command",
	} {
		assert.True(t, isGitProcessCredentialEnv(key), key)
	}
}

func envSliceMap(entries []string) map[string]string {
	result := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, found := strings.Cut(entry, "=")
		if found {
			result[key] = value
		}
	}
	return result
}
