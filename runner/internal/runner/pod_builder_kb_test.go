package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func TestKBAuthURL(t *testing.T) {
	assert.Equal(t,
		"https://x-access-token:tok@gitea.local/am-kb/docs.git",
		kbAuthURL("https://gitea.local/am-kb/docs.git", "tok"))
	assert.Equal(t,
		"http://x-access-token:tok@gitea.local/am-kb/docs.git",
		kbAuthURL("http://gitea.local/am-kb/docs.git", "tok"))
	assert.Equal(t,
		"https://gitea.local/am-kb/docs.git",
		kbAuthURL("https://gitea.local/am-kb/docs.git", ""))
	assert.Equal(t,
		"git@gitea.local:am-kb/docs.git",
		kbAuthURL("git@gitea.local:am-kb/docs.git", "tok"))
}

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

func TestCloneKnowledgeMount_ROStripsTokenFromRemote(t *testing.T) {
	origin := initKBOriginRepo(t)
	sandbox := t.TempDir()

	b := &PodBuilder{cmd: &runnerv1.CreatePodCommand{PodKey: "pod-1"}}
	m := &runnerv1.KnowledgeMount{
		Slug: "docs", HttpCloneUrl: origin, Branch: "main", Mode: "ro", GitToken: "secret",
	}
	require.NoError(t, b.cloneKnowledgeMount(context.Background(), sandbox, m))

	cfg, err := os.ReadFile(filepath.Join(sandbox, "kb", "docs", ".git", "config"))
	require.NoError(t, err)
	assert.NotContains(t, string(cfg), "secret", "ro mount must not retain the clone token")
	// Git for Windows normalizes path separators in .git/config; compare slash-folded.
	assert.Contains(t, filepath.ToSlash(string(cfg)), filepath.ToSlash(origin))
}
