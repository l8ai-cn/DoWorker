package runner

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testBuilder(cmd *runnerv1.CreatePodCommand) *PodBuilder {
	return NewPodBuilder(PodBuilderDeps{}).WithCommand(cmd)
}

const rootTpl = "{{.sandbox.root_path}}"

// --- createFiles tests ---

func TestCreateFiles_NormalFile(t *testing.T) {
	sandbox := t.TempDir()
	b := testBuilder(&runnerv1.CreatePodCommand{
		FilesToCreate: []*runnerv1.FileToCreate{
			{Path: rootTpl + "/hello.txt", Content: "world"},
		},
	})
	require.NoError(t, b.createFiles(sandbox, sandbox))
	data, err := os.ReadFile(filepath.Join(sandbox, "hello.txt"))
	require.NoError(t, err)
	assert.Equal(t, "world", string(data))
}

func TestCreateFiles_Directory(t *testing.T) {
	sandbox := t.TempDir()
	b := testBuilder(&runnerv1.CreatePodCommand{
		FilesToCreate: []*runnerv1.FileToCreate{
			{Path: rootTpl + "/subdir", IsDirectory: true},
		},
	})
	require.NoError(t, b.createFiles(sandbox, sandbox))
	info, err := os.Stat(filepath.Join(sandbox, "subdir"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestCreateFiles_AllowsExplicitWorkingDirectory(t *testing.T) {
	sandbox := t.TempDir()
	workDir := t.TempDir()
	b := testBuilder(&runnerv1.CreatePodCommand{
		FilesToCreate: []*runnerv1.FileToCreate{
			{Path: "{{.sandbox.work_dir}}/.codex/mcp.json", Content: "{}"},
		},
	})

	require.NoError(t, b.createFiles(sandbox, workDir))
	require.FileExists(t, filepath.Join(workDir, ".codex", "mcp.json"))
}

func TestCreateFiles_PathTraversal_DotDot(t *testing.T) {
	sandbox := t.TempDir()
	b := testBuilder(&runnerv1.CreatePodCommand{
		FilesToCreate: []*runnerv1.FileToCreate{
			{Path: rootTpl + "/../../../etc/passwd", Content: "evil"},
		},
	})
	err := b.createFiles(sandbox, sandbox)
	require.Error(t, err)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeFileCreate, podErr.Code)
}

func TestCreateFiles_PathTraversal_AbsolutePath(t *testing.T) {
	sandbox := t.TempDir()
	b := testBuilder(&runnerv1.CreatePodCommand{
		FilesToCreate: []*runnerv1.FileToCreate{
			{Path: "/tmp/evil.txt", Content: "evil"},
		},
	})
	err := b.createFiles(sandbox, sandbox)
	require.Error(t, err)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeFileCreate, podErr.Code)
}

func TestCreateFiles_PathTraversal_Symlink(t *testing.T) {
	sandbox := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.Symlink(outside, filepath.Join(sandbox, "outside")))
	b := testBuilder(&runnerv1.CreatePodCommand{
		FilesToCreate: []*runnerv1.FileToCreate{
			{Path: rootTpl + "/outside/escape.txt", Content: "evil"},
		},
	})

	err := b.createFiles(sandbox, sandbox)

	require.Error(t, err)
	require.NoFileExists(t, filepath.Join(outside, "escape.txt"))
}

func TestCreateFiles_CustomMode(t *testing.T) {
	sandbox := t.TempDir()
	b := testBuilder(&runnerv1.CreatePodCommand{
		FilesToCreate: []*runnerv1.FileToCreate{
			{Path: rootTpl + "/script.sh", Content: "#!/bin/sh", Mode: 0755},
		},
	})
	require.NoError(t, b.createFiles(sandbox, sandbox))
	info, err := os.Stat(filepath.Join(sandbox, "script.sh"))
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
	}
}

func TestCreateFiles_NestedDirectories(t *testing.T) {
	sandbox := t.TempDir()
	b := testBuilder(&runnerv1.CreatePodCommand{
		FilesToCreate: []*runnerv1.FileToCreate{
			{Path: rootTpl + "/a/b/c/d.txt", Content: "deep"},
		},
	})
	require.NoError(t, b.createFiles(sandbox, sandbox))
	data, err := os.ReadFile(filepath.Join(sandbox, "a", "b", "c", "d.txt"))
	require.NoError(t, err)
	assert.Equal(t, "deep", string(data))
}

func TestCreateFiles_EmptyList(t *testing.T) {
	sandbox := t.TempDir()
	b := testBuilder(&runnerv1.CreatePodCommand{})
	require.NoError(t, b.createFiles(sandbox, sandbox))
}

// --- createFilesFromProto tests ---

func TestCreateFilesFromProto_Normal(t *testing.T) {
	sandbox := t.TempDir()
	target := filepath.Join(sandbox, "proto.txt")
	b := testBuilder(&runnerv1.CreatePodCommand{})
	err := b.createFilesFromProto([]*runnerv1.FileToCreate{
		{Path: target, Content: "from proto"},
	}, sandbox, sandbox)
	require.NoError(t, err)
	data, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "from proto", string(data))
}

func TestCreateFilesFromProto_PathTraversal(t *testing.T) {
	sandbox := t.TempDir()
	b := testBuilder(&runnerv1.CreatePodCommand{})
	err := b.createFilesFromProto([]*runnerv1.FileToCreate{
		{Path: "/tmp/evil.txt", Content: "evil"},
	}, sandbox, sandbox)
	require.Error(t, err)
	var podErr *client.PodError
	require.ErrorAs(t, err, &podErr)
	assert.Equal(t, client.ErrCodeFileCreate, podErr.Code)
}

func TestCreateFilesFromProto_AllowsExplicitWorkingDirectory(t *testing.T) {
	sandbox := t.TempDir()
	workDir := t.TempDir()
	target := filepath.Join(workDir, ".codex", "mcp.json")
	b := testBuilder(&runnerv1.CreatePodCommand{})

	err := b.createFilesFromProto([]*runnerv1.FileToCreate{
		{Path: target, Content: "{}"},
	}, sandbox, workDir)

	require.NoError(t, err)
	require.FileExists(t, target)
}

func TestCreateFilesFromProto_PathTraversal_Symlink(t *testing.T) {
	sandbox := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.Symlink(outside, filepath.Join(sandbox, "outside")))
	target := filepath.Join(sandbox, "outside", "escape.txt")
	b := testBuilder(&runnerv1.CreatePodCommand{})

	err := b.createFilesFromProto([]*runnerv1.FileToCreate{
		{Path: target, Content: "evil"},
	}, sandbox, sandbox)

	require.Error(t, err)
	require.NoFileExists(t, filepath.Join(outside, "escape.txt"))
}

func TestCreateFilesFromProto_EmptyList(t *testing.T) {
	sandbox := t.TempDir()
	b := testBuilder(&runnerv1.CreatePodCommand{})
	require.NoError(t, b.createFilesFromProto(nil, sandbox, sandbox))
	require.NoError(t, b.createFilesFromProto([]*runnerv1.FileToCreate{}, sandbox, sandbox))
}
