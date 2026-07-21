package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func TestWorkerSkillDiscoverRequiresAllowedRelativeRoot(t *testing.T) {
	workspace := t.TempDir()
	writeWorkerSkillFixture(t, workspace, "skills/release-notes/SKILL.md")
	handler := &RunnerMessageHandler{}

	tests := []struct {
		name string
		path string
		err  string
	}{
		{name: "absolute path", path: filepath.Join(workspace, "skills/release-notes"), err: "worker skill path must be relative"},
		{name: "workspace escape", path: "../skills/release-notes", err: "worker skill path escapes workspace"},
		{name: "outside allowed root", path: "templates/release-notes", err: "worker skill path is outside allowed roots"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.dispatchSandboxFsOp(workspace, &runnerv1.SandboxFsCommand{
				Op:   "skill_discover",
				Path: tt.path,
			})

			require.NoError(t, err)
			require.Equal(t, tt.err, result.Error)
		})
	}
}

func TestWorkerSkillDiscoverAcceptsSkillDirectory(t *testing.T) {
	workspace := t.TempDir()
	writeWorkerSkillFixture(t, workspace, "skills/release-notes/SKILL.md")

	result, err := (&RunnerMessageHandler{}).dispatchSandboxFsOp(workspace, &runnerv1.SandboxFsCommand{
		Op:   "skill_discover",
		Path: "skills/release-notes",
	})

	require.NoError(t, err)
	require.Empty(t, result.Error)
	require.Len(t, result.Entries, 1)
	require.Equal(t, "skills/release-notes", result.Entries[0].Path)
}

func TestWorkerSkillDiscoverRejectsInvalidFrontmatter(t *testing.T) {
	workspace := t.TempDir()
	handler := &RunnerMessageHandler{}

	tests := []struct {
		name    string
		content string
		err     string
	}{
		{name: "missing frontmatter", content: "# Release notes\n", err: "worker skill frontmatter is invalid"},
		{name: "invalid name", content: "---\nname: Release Notes\ndescription: Draft release notes.\n---\n", err: "worker skill name is invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(workspace, "skills", tt.name, "SKILL.md")
			require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
			require.NoError(t, os.WriteFile(path, []byte(tt.content), 0o644))

			result, err := handler.dispatchSandboxFsOp(workspace, &runnerv1.SandboxFsCommand{
				Op:   "skill_discover",
				Path: filepath.ToSlash(filepath.Join("skills", tt.name)),
			})

			require.NoError(t, err)
			require.Equal(t, tt.err, result.Error)
		})
	}
}

func writeWorkerSkillFixture(t *testing.T, workspace, path string) {
	t.Helper()
	absolutePath := filepath.Join(workspace, path)
	require.NoError(t, os.MkdirAll(filepath.Dir(absolutePath), 0o755))
	require.NoError(t, os.WriteFile(absolutePath, []byte("---\nname: release-notes\ndescription: Draft release notes.\n---\n"), 0o644))
}
