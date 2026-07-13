package extension

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/infra/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageDirReusesExistingObject(t *testing.T) {
	var uploads int
	store := &svcMockStorage{
		existsFn: func(context.Context, string) (bool, error) {
			return true, nil
		},
		uploadFn: func(context.Context, string, io.Reader, int64, string) (*storage.FileInfo, error) {
			uploads++
			return nil, errors.New("must not upload existing object")
		},
	}
	pkg, err := NewSkillPackager(nil, store).PackageFromDir(
		context.Background(),
		ownershipTestSkillDir(t),
	)

	require.NoError(t, err)
	assert.False(t, pkg.Created)
	assert.Zero(t, uploads)
}

func TestPackageDirMarksNewlyUploadedObjectCreated(t *testing.T) {
	var uploads int
	store := &svcMockStorage{
		existsFn: func(context.Context, string) (bool, error) {
			return false, nil
		},
		uploadFn: func(_ context.Context, key string, _ io.Reader, size int64, _ string) (*storage.FileInfo, error) {
			uploads++
			return &storage.FileInfo{Key: key, Size: size}, nil
		},
	}
	pkg, err := NewSkillPackager(nil, store).PackageFromDir(
		context.Background(),
		ownershipTestSkillDir(t),
	)

	require.NoError(t, err)
	assert.True(t, pkg.Created)
	assert.Equal(t, 1, uploads)
}

func ownershipTestSkillDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "SKILL.md"),
		[]byte("---\nname: ownership-test\n---\nbody\n"),
		0o644,
	))
	return dir
}
