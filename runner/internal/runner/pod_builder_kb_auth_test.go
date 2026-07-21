package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneKnowledgeMountPublicHTTPUsesAnonymousGitIsolation(t *testing.T) {
	origin := initKBOriginRepo(t)
	sandbox := t.TempDir()
	commitSHA := kbHead(t, origin)
	realGit, logPath, globalConfig := installPublicKBGitHarness(t, origin)
	t.Setenv("GIT_SSH_COMMAND", "ssh -i /runner/key")
	t.Setenv("GIT_ASKPASS", "forbidden-global-askpass")
	t.Setenv("SSH_ASKPASS", "forbidden-global-ssh-askpass")
	mount := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: "https://gitea.test/am-kb/docs.git", Branch: "main", CommitSha: commitSHA, Mode: "ro",
	}
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{PodKey: "public-kb"}}

	require.NoError(t, builder.cloneKnowledgeMount(context.Background(), sandbox, mount))

	log := readKBFile(t, logPath)
	assert.Contains(t, log, "git_config_global="+os.DevNull)
	assert.Contains(t, log, "ssh_auth_sock=")
	assert.Contains(t, log, "askpass=")
	assert.Contains(t, log, "ssh_askpass=")
	assert.NotContains(t, log, "forbidden-global")
	assert.NotContains(t, log, "/runner/key")
	assert.Contains(t, log, "config_key_0=credential.helper")
	assert.Contains(t, log, "config_key_1=http.extraHeader")
	assert.Contains(t, log, "ssh=ssh -F")
	assert.Equal(t, initialKBGlobalConfig, readKBFile(t, globalConfig))
	dest := filepath.Join(sandbox, "kb", "docs")
	assert.Equal(t, "https://gitea.test/am-kb/docs.git", runKBGit(t, realGit, dest, "remote", "get-url", "origin"))
}

func TestKnowledgeMountBaseGitEnvironmentRemovesInheritedCredentials(t *testing.T) {
	t.Setenv("GIT_SSH_COMMAND", "ssh -i /runner/key")
	t.Setenv("GIT_ASKPASS", "forbidden-global-askpass")
	t.Setenv("SSH_ASKPASS", "forbidden-global-ssh-askpass")

	env := envSliceMap(knowledgeMountGitConfigEnv())

	assert.NotContains(t, env, "GIT_SSH_COMMAND")
	assert.Empty(t, env["GIT_ASKPASS"])
	assert.Empty(t, env["SSH_ASKPASS"])
	assert.Empty(t, env["SSH_AUTH_SOCK"])
}

func TestKnowledgeMountRejectsHTTPQueryCredentialsBeforeGit(t *testing.T) {
	sandbox := t.TempDir()
	marker := installKnowledgeMountNoGitHarness(t)
	mount := &runnerv1.KnowledgeMount{
		Slug:         "docs",
		HttpCloneUrl: "https://gitea.test/am-kb/docs.git?access_token=query-secret#fragment-secret",
		CommitSha:    testCommitSHA,
		Mode:         "ro",
	}
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{PodKey: "invalid-kb-url"}}

	err := builder.cloneKnowledgeMount(context.Background(), sandbox, mount)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not contain query or fragment")
	assert.NotContains(t, err.Error(), "query-secret")
	assert.NotContains(t, err.Error(), "fragment-secret")
	assert.NoFileExists(t, marker)
}

func TestKnowledgeMountSSHOnlyReadOnlyResumeKeepsSSHRemote(t *testing.T) {
	origin := initKBOriginRepo(t)
	sandbox := t.TempDir()
	realGit, _, _ := installKBGitHarness(t, origin, sandbox)
	sshURL := "ssh://git@gitea.test/am-kb/docs.git"
	mount := &runnerv1.KnowledgeMount{
		Slug: "docs", SshCloneUrl: sshURL, GitPrivateKey: "ro-key", GitKnownHosts: kbKnownHosts,
		Branch: "main", CommitSha: kbHead(t, origin), Mode: "ro",
	}
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{
		PodKey:        "ssh-ro",
		SandboxConfig: &runnerv1.SandboxConfig{KnowledgeMounts: []*runnerv1.KnowledgeMount{mount}},
	}}

	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))
	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))

	dest := filepath.Join(sandbox, "kb", "docs")
	assert.Equal(t, sshURL, runKBGit(t, realGit, dest, "remote", "get-url", "origin"))
	assert.Equal(t, kbHead(t, origin), kbHead(t, dest))
}

func TestKnowledgeMountReadOnlyToReadWriteResumeChecksOutBranch(t *testing.T) {
	origin := initKBOriginRepo(t)
	sandbox := t.TempDir()
	commitSHA := kbHead(t, origin)
	realGit, logPath, _ := installKBGitHarness(t, origin, sandbox)
	sshURL := "ssh://git@gitea.test/am-kb/docs.git"
	mount := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: "https://gitea.test/am-kb/docs.git", SshCloneUrl: sshURL,
		GitPrivateKey: "rw-key", GitKnownHosts: kbKnownHosts, Branch: "main", CommitSha: commitSHA, Mode: "ro",
	}
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{
		PodKey:        "ro-to-rw",
		SandboxConfig: &runnerv1.SandboxConfig{KnowledgeMounts: []*runnerv1.KnowledgeMount{mount}},
	}}
	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))
	mount.Mode = "rw"

	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))
	assert.NotContains(t, readKBFile(t, logPath), "/runner/key")

	dest := filepath.Join(sandbox, "kb", "docs")
	assert.Equal(t, "refs/heads/main", kbSymbolicRef(t, dest))
	require.NoError(t, os.WriteFile(filepath.Join(dest, "rw.txt"), []byte("rw"), 0644))
	runKBGit(t, realGit, dest, "add", ".")
	runKBGit(t, realGit, dest, "commit", "-m", "rw")
}

func installPublicKBGitHarness(t *testing.T, origin string) (string, string, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("public KB harness requires a POSIX shell")
	}
	realGit, err := exec.LookPath("git")
	require.NoError(t, err)
	dir := t.TempDir()
	logPath := filepath.Join(dir, "git.log")
	globalConfig := filepath.Join(dir, "global.gitconfig")
	require.NoError(t, os.WriteFile(globalConfig, []byte(initialKBGlobalConfig), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "git"), []byte(publicKBGitHarnessScript), 0755))
	t.Setenv("KB_REAL_GIT", realGit)
	t.Setenv("KB_ORIGIN", origin)
	t.Setenv("KB_GIT_LOG", logPath)
	t.Setenv("GIT_CONFIG_GLOBAL", globalConfig)
	t.Setenv("SSH_AUTH_SOCK", "forbidden-agent")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return realGit, logPath, globalConfig
}

func installKnowledgeMountNoGitHarness(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	marker := filepath.Join(dir, "git-called")
	name := "git"
	content := "#!/bin/sh\n: > \"$KB_GIT_MARKER\"\nexit 42\n"
	if runtime.GOOS == "windows" {
		name = "git.bat"
		content = "@echo off\r\necho called>\"%KB_GIT_MARKER%\"\r\nexit /b 42\r\n"
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0755))
	t.Setenv("KB_GIT_MARKER", marker)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return marker
}

const publicKBGitHarnessScript = `#!/bin/sh
set -eu
{
	printf 'ssh=%s\n' "${GIT_SSH_COMMAND:-}"
	printf 'ssh_auth_sock=%s\n' "${SSH_AUTH_SOCK:-}"
	printf 'askpass=%s\n' "${GIT_ASKPASS:-}"
	printf 'ssh_askpass=%s\n' "${SSH_ASKPASS:-}"
	printf 'git_config_global=%s\n' "${GIT_CONFIG_GLOBAL:-}"
	printf 'config_key_0=%s\n' "${GIT_CONFIG_KEY_0:-}"
	printf 'config_key_1=%s\n' "${GIT_CONFIG_KEY_1:-}"
	for arg in "$@"; do printf 'arg=%s\n' "$arg"; done
} >> "$KB_GIT_LOG"

case "${1:-}" in
clone)
	previous=
	last=
	for arg in "$@"; do previous=$last; last=$arg; done
	"$KB_REAL_GIT" clone "$KB_ORIGIN" "$last"
	"$KB_REAL_GIT" -C "$last" remote set-url origin "$previous"
	;;
fetch)
	rebuilt=
	for arg in "$@"; do
		if [ "$arg" = "origin" ]; then rebuilt="$rebuilt '$KB_ORIGIN'"; else rebuilt="$rebuilt '$arg'"; fi
	done
	eval "exec \"$KB_REAL_GIT\" $rebuilt"
	;;
*)
	exec "$KB_REAL_GIT" "$@"
	;;
esac
`

var _ = strings.TrimSpace
