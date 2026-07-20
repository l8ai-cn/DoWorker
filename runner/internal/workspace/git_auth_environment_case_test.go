package workspace

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveInheritedGitAuthEnvIgnoresKeyCase(t *testing.T) {
	env := []string{
		"PATH=/usr/bin",
		"git_config_global=/tmp/global",
		"Git_Config_NoSystem=0",
		"git_config_count=1",
		"Git_Config_Key_0=http.extraHeader",
		"git_config_value_0=Authorization: leaked",
		"git_ssh_command=ssh -i leaked",
		"ssh_auth_sock=/tmp/agent",
		"git_askpass=/tmp/askpass",
		"ssh_askpass=/tmp/ssh-askpass",
	}

	assert.Equal(t, []string{"PATH=/usr/bin"}, removeInheritedGitAuthEnv(env))
}

func TestExplicitAuthIsolatesLocalWorktreeGitCommands(t *testing.T) {
	origin, clone := createPinnedOrigin(t)
	commit := commitPinnedFile(t, clone, "isolated")
	pushPinnedBranch(t, clone)

	for _, testCase := range []struct {
		name    string
		options []WorktreeOption
	}{
		{name: "branch", options: []WorktreeOption{WithAnonymousAuth()}},
		{name: "pinned", options: []WorktreeOption{WithAnonymousAuth(), WithSourceCommitSHA(commit)}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			manager, err := NewManager(root, "")
			require.NoError(t, err)
			t.Setenv("GIT_CONFIG_COUNT", "1")
			t.Setenv("GIT_CONFIG_KEY_0", "core.repositoryformatversion")
			t.Setenv("GIT_CONFIG_VALUE_0", "999")

			result, err := manager.CreateWorktreeWithOptions(
				testGitContext(t),
				origin,
				"main",
				filepath.Join(root, "sandboxes", testCase.name, "workspace"),
				testCase.options...,
			)

			require.NoError(t, err)
			assert.DirExists(t, result.Path)
			assert.FileExists(t, filepath.Join(result.Path, "file.txt"))
		})
	}
}

func TestLocalGitEnvDoesNotExposeToken(t *testing.T) {
	manager, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "http.extraHeader")
	t.Setenv("GIT_CONFIG_VALUE_0", "Authorization: Basic inherited-secret")
	t.Setenv("GIT_ASKPASS", "/tmp/token-askpass")

	cmd := exec.Command("git", "status")
	manager.setLocalGitEnv(cmd)
	env := envValues(cmd.Env)

	assert.Equal(t, "2", env["GIT_CONFIG_COUNT"])
	assert.Equal(t, "credential.helper", env["GIT_CONFIG_KEY_0"])
	assert.Empty(t, env["GIT_CONFIG_VALUE_0"])
	assert.Equal(t, "http.extraHeader", env["GIT_CONFIG_KEY_1"])
	assert.Empty(t, env["GIT_CONFIG_VALUE_1"])
	assert.Empty(t, env["GIT_ASKPASS"])
	serialized := strings.Join(cmd.Env, "\n")
	assert.NotContains(t, serialized, "Authorization: Basic inherited-secret")
	assert.NotContains(t, serialized, "/tmp/token-askpass")
}
