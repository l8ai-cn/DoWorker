package runner

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteSandboxSSHKeyWindowsACLFailureFailsClosed(t *testing.T) {
	restore := forceWindowsACL(t, errors.New("acl denied"))
	defer restore()
	sandbox := t.TempDir()
	builder := &PodBuilder{cmd: &runnerv1.CreatePodCommand{PodKey: "acl-main"}}

	_, err := builder.writeSandboxSSHKey(context.Background(), sandbox, "private-key")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set SSH key ACL")
	assert.NoFileExists(t, filepath.Join(sandbox, ".ssh_key"))
}

func TestWriteKnowledgeMountKeyWindowsACLFailureCleansTemporaryKey(t *testing.T) {
	restore := forceWindowsACL(t, errors.New("acl denied"))
	defer restore()
	sandbox := t.TempDir()

	_, err := writeKnowledgeMountKey(sandbox, "private-key", kbKnownHosts)

	require.Error(t, err)
	assert.Empty(t, globKnowledgeKeys(t, sandbox))
}

func forceWindowsACL(t *testing.T, err error) func() {
	t.Helper()
	oldGOOS, oldICACLS := runnerGOOS, runICACLS
	runnerGOOS = "windows"
	t.Setenv("USERNAME", "runner")
	runICACLS = func(_, _ string) error { return err }
	return func() {
		runnerGOOS = oldGOOS
		runICACLS = oldICACLS
	}
}

func globKnowledgeKeys(t *testing.T, sandbox string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(sandbox, ".agentsmesh-kb-key-*"))
	require.NoError(t, err)
	return matches
}
