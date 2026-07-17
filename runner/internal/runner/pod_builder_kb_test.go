package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/client"
)

func initKBOriginRepo(t *testing.T) string {
	t.Helper()
	origin := t.TempDir()
	run := func(dir string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
	}
	run(origin, "init", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(origin, "llms.txt"), []byte("# KB\n"), 0644))
	run(origin, "add", ".")
	run(origin, "commit", "-m", "init")
	return origin
}

func TestSetupKnowledgeMounts_CloneAndSkipExisting(t *testing.T) {
	origin := initKBOriginRepo(t)
	sandbox := t.TempDir()

	b := &PodBuilder{cmd: &runnerv1.CreatePodCommand{
		PodKey: "pod-1",
		SandboxConfig: &runnerv1.SandboxConfig{
			KnowledgeMounts: []*runnerv1.KnowledgeMount{
				{Slug: "docs", HttpCloneUrl: origin, Branch: "main", MountPath: "kb/docs", Mode: "ro"},
			},
		},
	}}

	require.NoError(t, b.setupKnowledgeMounts(context.Background(), sandbox))
	assert.FileExists(t, filepath.Join(sandbox, "kb", "docs", "llms.txt"))

	// Second run (resume) must be a no-op, not a clone failure.
	require.NoError(t, b.setupKnowledgeMounts(context.Background(), sandbox))
}

func TestCloneKnowledgeMount_PublicHTTPCloneKeepsRemote(t *testing.T) {
	origin := initKBOriginRepo(t)
	sandbox := t.TempDir()

	b := &PodBuilder{cmd: &runnerv1.CreatePodCommand{PodKey: "pod-1"}}
	m := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: origin, Branch: "main", Mode: "ro",
	}
	require.NoError(t, b.cloneKnowledgeMount(context.Background(), sandbox, m))

	dest := filepath.Join(sandbox, "kb", "docs")
	got := strings.ReplaceAll(filepath.ToSlash(runKBGit(t, "git", dest, "remote", "get-url", "origin")), "//", "/")
	want := strings.ReplaceAll(filepath.ToSlash(origin), "//", "/")
	assert.Equal(t, want, got)
}

func TestCloneKnowledgeMount_SSHReadOnlyDeletesKeyAndRetainsNoCredential(t *testing.T) {
	origin := initKBOriginRepo(t)
	sandbox := t.TempDir()
	realGit, logPath, globalConfig := installKBGitHarness(t, origin, sandbox)
	privateKey := "-----BEGIN OPENSSH PRIVATE KEY-----\nread-only-secret\n-----END OPENSSH PRIVATE KEY-----\n"
	sshURL := "ssh://git@gitea.test/am-kb/docs.git"
	httpURL := "https://gitea.test/am-kb/docs.git"
	m := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: httpURL, SshCloneUrl: sshURL,
		GitPrivateKey: privateKey, GitKnownHosts: kbKnownHosts, Branch: "main", Mode: "ro",
	}

	b := &PodBuilder{cmd: &runnerv1.CreatePodCommand{PodKey: "pod-1"}}
	require.NoError(t, b.cloneKnowledgeMount(context.Background(), sandbox, m))

	dest := filepath.Join(sandbox, "kb", "docs")
	log := readKBFile(t, logPath)
	temporaryKey := kbLogValue(t, log, "key=")
	assert.Equal(t, "600", kbLogValue(t, log, "mode="))
	assert.NoFileExists(t, temporaryKey)
	assert.NoFileExists(t, temporaryKey+temporaryKnownHostsFileSuffix)
	assert.NoFileExists(t, filepath.Join(dest, ".git", knowledgeMountDeployKey))
	assert.NoFileExists(t, filepath.Join(dest, ".git", knowledgeMountKnownHosts))
	assert.Equal(t, httpURL, runKBGit(t, realGit, dest, "remote", "get-url", "origin"))
	cfg := readKBFile(t, filepath.Join(dest, ".git", "config"))
	assert.NotContains(t, cfg, "sshCommand")
	assert.NotContains(t, cfg, privateKey)
	assert.Contains(t, log, "arg="+sshURL)
	assert.NotContains(t, log, privateKey)
	assert.Contains(t, log, "StrictHostKeyChecking=yes")
	assert.NotContains(t, log, "StrictHostKeyChecking=no")
	assert.Contains(t, log, "git_config_global="+os.DevNull)
	assert.Equal(t, initialKBGlobalConfig, readKBFile(t, globalConfig))
}

func TestCloneKnowledgeMount_SSHReadWritePersistsKeyAndLocalPushConfig(t *testing.T) {
	origin := initKBOriginRepo(t)
	sandbox := t.TempDir()
	realGit, logPath, globalConfig := installKBGitHarness(t, origin, sandbox)
	privateKey := "-----BEGIN OPENSSH PRIVATE KEY-----\nread-write-secret\n-----END OPENSSH PRIVATE KEY-----\n"
	sshURL := "git@gitea.test:am-kb/docs.git"
	m := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: "https://gitea.test/am-kb/docs.git", SshCloneUrl: sshURL,
		GitPrivateKey: privateKey, GitKnownHosts: kbKnownHosts, Branch: "main", Mode: "rw",
	}

	b := &PodBuilder{cmd: &runnerv1.CreatePodCommand{PodKey: "pod-1"}}
	require.NoError(t, b.cloneKnowledgeMount(context.Background(), sandbox, m))

	dest := filepath.Join(sandbox, "kb", "docs")
	keyPath := filepath.Join(dest, ".git", knowledgeMountDeployKey)
	knownHostsPath := filepath.Join(dest, ".git", knowledgeMountKnownHosts)
	info, err := os.Stat(keyPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	assert.Equal(t, privateKey, readKBFile(t, keyPath))
	knownHostsInfo, err := os.Stat(knownHostsPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), knownHostsInfo.Mode().Perm())
	assert.Equal(t, kbKnownHosts+"\n", readKBFile(t, knownHostsPath))
	log := readKBFile(t, logPath)
	temporaryKey := kbLogValue(t, log, "key=")
	assert.NoFileExists(t, temporaryKey)
	assert.NoFileExists(t, temporaryKey+temporaryKnownHostsFileSuffix)
	assert.Equal(t, sshURL, runKBGit(t, realGit, dest, "remote", "get-url", "origin"))
	sshCommand := runKBGit(t, realGit, dest, "config", "--local", "--get", "core.sshCommand")
	assert.Contains(t, sshCommand, keyPath)
	assert.Contains(t, sshCommand, knownHostsPath)
	assert.Contains(t, sshCommand, "IdentitiesOnly=yes")
	assert.Contains(t, sshCommand, "StrictHostKeyChecking=yes")
	assert.NotContains(t, sshCommand, "StrictHostKeyChecking=no")
	assert.NotContains(t, sshCommand, privateKey)
	assert.NotContains(t, readKBFile(t, filepath.Join(dest, ".git", "config")), privateKey)
	assert.NotContains(t, log, privateKey)
	assert.Equal(t, initialKBGlobalConfig, readKBFile(t, globalConfig))
}

func TestCloneKnowledgeMount_SSHFailureCleansAndRedactsPrivateKey(t *testing.T) {
	origin := initKBOriginRepo(t)
	sandbox := t.TempDir()
	_, logPath, globalConfig := installKBGitHarness(t, origin, sandbox)
	t.Setenv("KB_GIT_FAIL", "1")
	privateKey := "failure-private-key-secret"
	m := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: origin, SshCloneUrl: "ssh://git@gitea.test/am-kb/docs.git",
		GitPrivateKey: privateKey, GitKnownHosts: kbKnownHosts, Branch: "main", Mode: "ro",
	}

	b := &PodBuilder{cmd: &runnerv1.CreatePodCommand{PodKey: "pod-1"}}
	err := b.cloneKnowledgeMount(context.Background(), sandbox, m)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeGitClone, podErr.Code)
	assert.NotContains(t, err.Error(), privateKey)
	assert.Contains(t, err.Error(), "[REDACTED]")
	log := readKBFile(t, logPath)
	assert.NotContains(t, log, privateKey)
	temporaryKey := kbLogValue(t, log, "key=")
	assert.NoFileExists(t, temporaryKey)
	assert.NoFileExists(t, temporaryKey+temporaryKnownHostsFileSuffix)
	assert.NoDirExists(t, filepath.Join(sandbox, "kb", "docs"))
	assert.Equal(t, initialKBGlobalConfig, readKBFile(t, globalConfig))
}

func TestCloneKnowledgeMount_PrivateKeyRequiresSSHURL(t *testing.T) {
	sandbox := t.TempDir()
	privateKey := "must-not-leak"
	m := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: initKBOriginRepo(t), GitPrivateKey: privateKey, Mode: "ro",
	}

	b := &PodBuilder{cmd: &runnerv1.CreatePodCommand{PodKey: "pod-1"}}
	err := b.cloneKnowledgeMount(context.Background(), sandbox, m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ssh_clone_url and git_known_hosts are required")
	assert.NotContains(t, err.Error(), privateKey)
	assert.NoDirExists(t, filepath.Join(sandbox, "kb", "docs"))
}

func TestSetupKnowledgeMounts_ReconcilesResumeCredentialModes(t *testing.T) {
	origin := initKBOriginRepo(t)
	sandbox := t.TempDir()
	realGit, _, _ := installKBGitHarness(t, origin, sandbox)
	httpURL := "https://gitea.test/am-kb/docs.git"
	sshURL := "ssh://git@gitea.test/am-kb/docs.git"
	mount := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: httpURL, SshCloneUrl: sshURL,
		GitPrivateKey: "rw-key-one", GitKnownHosts: kbKnownHosts,
		Branch: "main", Mode: "rw",
	}
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{
		PodKey: "resume-pod",
		SandboxConfig: &runnerv1.SandboxConfig{
			KnowledgeMounts: []*runnerv1.KnowledgeMount{mount},
		},
	}}

	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))
	dest := filepath.Join(sandbox, "kb", "docs")
	keyPath := filepath.Join(dest, ".git", knowledgeMountDeployKey)
	knownHostsPath := filepath.Join(dest, ".git", knowledgeMountKnownHosts)
	assert.Equal(t, "rw-key-one", readKBFile(t, keyPath))

	mount.Mode = "ro"
	mount.GitPrivateKey = "ro-key"
	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))
	assert.NoFileExists(t, keyPath)
	assert.NoFileExists(t, knownHostsPath)
	_, err := runKBGitError(realGit, dest, "config", "--local", "--get", "core.sshCommand")
	require.Error(t, err)
	assert.Equal(t, httpURL, runKBGit(t, realGit, dest, "remote", "get-url", "origin"))

	mount.Mode = "rw"
	mount.GitPrivateKey = "rw-key-two"
	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))
	assert.Equal(t, "rw-key-two", readKBFile(t, keyPath))
	assert.Equal(t, sshURL, runKBGit(t, realGit, dest, "remote", "get-url", "origin"))

	builder.cmd.SandboxConfig.KnowledgeMounts = nil
	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))
	assert.NoFileExists(t, keyPath)
	assert.NoFileExists(t, knownHostsPath)
	assert.FileExists(t, filepath.Join(dest, "llms.txt"))
}

func TestSetupKnowledgeMounts_RejectsSymlinkedGitDirectoryOnResume(t *testing.T) {
	origin := initKBOriginRepo(t)
	sandbox := t.TempDir()
	dest := filepath.Join(sandbox, "kb", "docs")
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{
		PodKey: "resume-pod",
		SandboxConfig: &runnerv1.SandboxConfig{
			KnowledgeMounts: []*runnerv1.KnowledgeMount{{
				Slug: "docs", HttpCloneUrl: origin, Branch: "main", Mode: "ro",
			}},
		},
	}}
	require.NoError(t, builder.setupKnowledgeMounts(context.Background(), sandbox))

	externalGit := t.TempDir()
	require.NoError(t, os.RemoveAll(filepath.Join(dest, ".git")))
	require.NoError(t, os.Symlink(externalGit, filepath.Join(dest, ".git")))
	mount := builder.cmd.SandboxConfig.KnowledgeMounts[0]
	mount.Mode = "rw"
	mount.SshCloneUrl = "ssh://git@gitea.test/am-kb/docs.git"
	mount.GitPrivateKey = "must-not-escape"
	mount.GitKnownHosts = kbKnownHosts

	err := builder.setupKnowledgeMounts(context.Background(), sandbox)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must contain only directories")
	assert.NoFileExists(t, filepath.Join(externalGit, knowledgeMountDeployKey))
	assert.NoFileExists(t, filepath.Join(externalGit, knowledgeMountKnownHosts))
}

func TestSetupKnowledgeMounts_RejectsSymlinkedDestinationParent(t *testing.T) {
	sandbox := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "keep.txt"), []byte("keep"), 0600))
	require.NoError(t, os.Symlink(outside, filepath.Join(sandbox, "kb")))
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{
		PodKey: "new-pod",
		SandboxConfig: &runnerv1.SandboxConfig{
			KnowledgeMounts: []*runnerv1.KnowledgeMount{{
				Slug: "docs", HttpCloneUrl: initKBOriginRepo(t), Branch: "main", Mode: "ro",
			}},
		},
	}}

	err := builder.setupKnowledgeMounts(context.Background(), sandbox)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must contain only directories")
	assert.FileExists(t, filepath.Join(outside, "keep.txt"))
	assert.NoDirExists(t, filepath.Join(outside, "docs"))
}

const initialKBGlobalConfig = "[credential]\n\thelper = forbidden-global-helper\n"
const kbKnownHosts = "gitea.test ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITestHostKey"

func installKBGitHarness(t *testing.T, origin, sandbox string) (string, string, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("SSH credential lifecycle harness requires a POSIX shell")
	}
	realGit, err := exec.LookPath("git")
	require.NoError(t, err)
	dir := t.TempDir()
	logPath := filepath.Join(dir, "git.log")
	globalConfig := filepath.Join(dir, "global.gitconfig")
	require.NoError(t, os.WriteFile(globalConfig, []byte(initialKBGlobalConfig), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "git"), []byte(kbGitHarnessScript), 0755))
	t.Setenv("KB_REAL_GIT", realGit)
	t.Setenv("KB_ORIGIN", origin)
	t.Setenv("KB_SANDBOX", sandbox)
	t.Setenv("KB_GIT_LOG", logPath)
	t.Setenv("GIT_CONFIG_GLOBAL", globalConfig)
	t.Setenv("GIT_ASKPASS", "forbidden-global-askpass")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return realGit, logPath, globalConfig
}

func runKBGit(t *testing.T, gitPath, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(gitPath, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, output)
	return strings.TrimSpace(string(output))
}

func runKBGitError(gitPath, dir string, args ...string) ([]byte, error) {
	cmd := exec.Command(gitPath, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func readKBFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}

func kbLogValue(t *testing.T, log, prefix string) string {
	t.Helper()
	for _, line := range strings.Split(log, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimPrefix(line, prefix)
		}
	}
	t.Fatalf("missing %q in git log:\n%s", prefix, log)
	return ""
}

const kbGitHarnessScript = `#!/bin/sh
set -eu
{
	printf 'ssh=%s\n' "${GIT_SSH_COMMAND:-}"
	printf 'git_config_global=%s\n' "${GIT_CONFIG_GLOBAL:-}"
	printf 'askpass=%s\n' "${GIT_ASKPASS:-}"
	for arg in "$@"; do printf 'arg=%s\n' "$arg"; done
} >> "$KB_GIT_LOG"

if [ "${1:-}" != "clone" ]; then
	exec "$KB_REAL_GIT" "$@"
fi

key=
for candidate in "$KB_SANDBOX"/.agentsmesh-kb-key-*; do
	[ -f "$candidate" ] || continue
	key=$candidate
	break
done
[ -n "$key" ] || { printf 'temporary key missing\n' >&2; exit 31; }
mode=$(stat -f '%Lp' "$key" 2>/dev/null || stat -c '%a' "$key")
printf 'key=%s\nmode=%s\n' "$key" "$mode" >> "$KB_GIT_LOG"

if [ "${KB_GIT_FAIL:-}" = "1" ]; then
	printf 'fatal: ' >&2
	cat "$key" >&2
	exit 23
fi

previous=
last=
for arg in "$@"; do
	previous=$last
	last=$arg
done
"$KB_REAL_GIT" clone "$KB_ORIGIN" "$last"
"$KB_REAL_GIT" -C "$last" remote set-url origin "$previous"
`
