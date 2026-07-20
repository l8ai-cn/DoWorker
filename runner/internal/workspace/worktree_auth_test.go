package workspace

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnonymousAuthEnvDisablesRunnerCredentials(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	opts := &WorktreeOptions{AnonymousAuth: true}

	probeCmd := exec.Command("git", "ls-remote")
	mgr.setProbeEnv(probeCmd, opts)
	probeEnv := envValues(probeCmd.Env)
	assert.Equal(t, os.DevNull, probeEnv["GIT_CONFIG_GLOBAL"])
	assert.Equal(t, "1", probeEnv["GIT_CONFIG_NOSYSTEM"])
	assert.Equal(t, "2", probeEnv["GIT_CONFIG_COUNT"])
	assert.Equal(t, "credential.helper", probeEnv["GIT_CONFIG_KEY_0"])
	assert.Equal(t, "", probeEnv["GIT_CONFIG_VALUE_0"])
	assert.Equal(t, "http.extraHeader", probeEnv["GIT_CONFIG_KEY_1"])
	assert.Equal(t, "", probeEnv["GIT_CONFIG_VALUE_1"])
	assert.Equal(t, "", probeEnv["SSH_AUTH_SOCK"])
	assert.Contains(t, probeEnv["GIT_SSH_COMMAND"], "IdentityAgent=none")
	assert.Contains(t, probeEnv["GIT_SSH_COMMAND"], "-F ")
	assert.Contains(t, probeEnv["GIT_SSH_COMMAND"], "IdentityFile=")
	assert.Contains(t, probeEnv["GIT_SSH_COMMAND"], os.DevNull)

	fetchCmd := exec.Command("git", "fetch")
	mgr.setGitAuthEnv(fetchCmd, opts)
	fetchEnv := envValues(fetchCmd.Env)
	assert.Equal(t, os.DevNull, fetchEnv["GIT_CONFIG_GLOBAL"])
	assert.Equal(t, "1", fetchEnv["GIT_CONFIG_NOSYSTEM"])
	assert.Contains(t, fetchEnv["GIT_SSH_COMMAND"], "IdentityAgent=none")
	assert.Contains(t, fetchEnv["GIT_SSH_COMMAND"], "-F ")
	assert.Contains(t, fetchEnv["GIT_SSH_COMMAND"], "IdentityFile=")
	assert.Contains(t, fetchEnv["GIT_SSH_COMMAND"], os.DevNull)
}

func TestTokenAuthEnvDisablesRunnerCredentialsAndKeepsTokenOutOfArgv(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	opts := &WorktreeOptions{GitToken: "secret-token"}
	cmd := exec.Command("git", "fetch", "https://example.test/repo.git")

	assert.Equal(t, "https://example.test/repo.git", mgr.prepareAuthURL("https://example.test/repo.git", opts))
	mgr.setGitAuthEnv(cmd, opts)
	env := envValues(cmd.Env)
	assert.Equal(t, os.DevNull, env["GIT_CONFIG_GLOBAL"])
	assert.Equal(t, "1", env["GIT_CONFIG_NOSYSTEM"])
	assert.Equal(t, "", env["SSH_AUTH_SOCK"])
	assert.Equal(t, "3", env["GIT_CONFIG_COUNT"])
	assert.Equal(t, "credential.helper", env["GIT_CONFIG_KEY_0"])
	assert.Equal(t, "", env["GIT_CONFIG_VALUE_0"])
	assert.Equal(t, "http.extraHeader", env["GIT_CONFIG_KEY_1"])
	assert.Equal(t, "", env["GIT_CONFIG_VALUE_1"])
	assert.Equal(t, "http.extraHeader", env["GIT_CONFIG_KEY_2"])
	expectedAuth := "Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte("x-access-token:secret-token"))
	assert.Equal(t, expectedAuth, env["GIT_CONFIG_VALUE_2"])
	assert.Contains(t, env["GIT_SSH_COMMAND"], "IdentitiesOnly=yes")
	assert.Contains(t, env["GIT_SSH_COMMAND"], "IdentityAgent=none")
	assert.Contains(t, env["GIT_SSH_COMMAND"], "-F ")
	assert.Contains(t, env["GIT_SSH_COMMAND"], "IdentityFile=")
	assert.NotContains(t, strings.Join(cmd.Args, " "), "secret-token")
	assert.Equal(t, "fatal [REDACTED]", mgr.redactGitOutput(opts, []byte("fatal secret-token")))
}

func TestTokenProbeRedactsGitOutputAndEnvToken(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	installEchoFailGitHarness(t)

	_, err = mgr.probeRepositoryAccess(context.Background(), "https://private.test/repo.git", "", &WorktreeOptions{GitToken: "secret-token"})

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "secret-token")
	assert.Contains(t, err.Error(), "[REDACTED]")
}

func TestSSHKeyAuthEnvDisablesRunnerCredentials(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	opts := &WorktreeOptions{SSHKeyPath: filepath.Join(t.TempDir(), "id key")}

	cmd := exec.Command("git", "fetch")
	mgr.setGitAuthEnv(cmd, opts)
	env := envValues(cmd.Env)
	assert.Equal(t, os.DevNull, env["GIT_CONFIG_GLOBAL"])
	assert.Equal(t, "1", env["GIT_CONFIG_NOSYSTEM"])
	assert.Equal(t, "", env["SSH_AUTH_SOCK"])
	assert.Equal(t, "2", env["GIT_CONFIG_COUNT"])
	assert.Equal(t, "credential.helper", env["GIT_CONFIG_KEY_0"])
	assert.Equal(t, "http.extraHeader", env["GIT_CONFIG_KEY_1"])
	assert.Contains(t, env["GIT_SSH_COMMAND"], "IdentitiesOnly=yes")
	assert.Contains(t, env["GIT_SSH_COMMAND"], "IdentityAgent=none")
	assert.Contains(t, env["GIT_SSH_COMMAND"], "-F ")
	assert.Contains(t, env["GIT_SSH_COMMAND"], "IdentityFile=none")
	assert.Contains(t, env["GIT_SSH_COMMAND"], "-i "+gitSSHQuote(opts.SSHKeyPath))
}

func TestAnonymousAuthCannotUseRunnerLocalCredentials(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), "")
	require.NoError(t, err)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "")
	installProbeGitHarness(t)

	_, err = mgr.probeRepositoryAccess(context.Background(), "https://private.test/repo.git", "", &WorktreeOptions{})
	require.NoError(t, err)

	_, err = mgr.probeRepositoryAccess(context.Background(), "https://private.test/repo.git", "", &WorktreeOptions{AnonymousAuth: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP(anonymous)")
	assert.NotContains(t, err.Error(), "HTTP(local)")
}

func installProbeGitHarness(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	executable, err := os.Executable()
	require.NoError(t, err)
	name := "git"
	content := probeGitHarnessUnix(executable)
	if runtime.GOOS == "windows" {
		name = "git.bat"
		content = probeGitHarnessWindows(executable)
	}
	script := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(script, []byte(content), 0755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func envValues(env []string) map[string]string {
	values := map[string]string{}
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			values[key] = value
		}
	}
	return values
}

func probeGitHarnessUnix(executable string) string {
	return fmt.Sprintf(`#!/bin/sh
exec %s -test.run=TestGitProbeHelperProcess -- agentsmesh-git-probe-helper "$@"
`, shellSingleQuote(executable))
}

func probeGitHarnessWindows(executable string) string {
	return fmt.Sprintf(`@echo off
"%s" -test.run=TestGitProbeHelperProcess -- agentsmesh-git-probe-helper %%*
exit /b %%ERRORLEVEL%%
`, strings.ReplaceAll(executable, `"`, `""`))
}

func TestGitProbeHelperProcess(t *testing.T) {
	args := testHelperArgs(os.Args)
	if len(args) == 0 || args[0] != "agentsmesh-git-probe-helper" {
		return
	}
	gitArgs := args[1:]
	if len(gitArgs) == 0 || gitArgs[0] != "ls-remote" {
		fmt.Fprintf(os.Stderr, "unexpected git command: %s\n", strings.Join(gitArgs, " "))
		os.Exit(99)
	}
	if os.Getenv("GIT_CONFIG_NOSYSTEM") == "1" {
		fmt.Fprintln(os.Stderr, "runner local credentials unavailable")
		os.Exit(23)
	}
	fmt.Println("0123456789abcdef0123456789abcdef01234567\tHEAD")
	os.Exit(0)
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func testHelperArgs(args []string) []string {
	for index, arg := range args {
		if arg == "--" {
			return args[index+1:]
		}
	}
	return nil
}

func installEchoFailGitHarness(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	name := "git"
	content := echoFailGitHarnessUnix()
	if runtime.GOOS == "windows" {
		name = "git.bat"
		content = echoFailGitHarnessWindows()
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func echoFailGitHarnessUnix() string {
	return `#!/bin/sh
printf 'argv=%s\n' "$*" >&2
printf 'token=%s\n' "${GIT_CONFIG_VALUE_2:-}" >&2
exit 42
`
}

func echoFailGitHarnessWindows() string {
	return `@echo off
echo argv=%* 1>&2
echo token=%GIT_CONFIG_VALUE_2% 1>&2
exit /b 42
`
}
