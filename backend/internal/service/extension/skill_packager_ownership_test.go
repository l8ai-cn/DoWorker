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

func TestPrepareFromDirDoesNotAccessStorage(t *testing.T) {
	var storageCalls int
	store := &svcMockStorage{
		existsFn: func(context.Context, string) (bool, error) {
			storageCalls++
			return false, nil
		},
		uploadFn: func(context.Context, string, io.Reader, int64, string) (*storage.FileInfo, error) {
			storageCalls++
			return nil, errors.New("unexpected upload")
		},
	}

	prepared, err := NewSkillPackager(nil, store).PrepareFromDir(
		context.Background(),
		ownershipTestSkillDir(t),
	)

	require.NoError(t, err)
	assert.NotEmpty(t, prepared.Data)
	assert.Zero(t, storageCalls)
}

func TestStorePreparedReusesExistingObject(t *testing.T) {
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
	packager := NewSkillPackager(nil, store)
	prepared, err := packager.PrepareFromDir(context.Background(), ownershipTestSkillDir(t))
	require.NoError(t, err)

	pkg, err := packager.StorePrepared(context.Background(), prepared)

	require.NoError(t, err)
	assert.False(t, pkg.Created)
	assert.Zero(t, uploads)
}

func TestStorePreparedMarksNewUploadCreated(t *testing.T) {
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
	packager := NewSkillPackager(nil, store)
	prepared, err := packager.PrepareFromDir(context.Background(), ownershipTestSkillDir(t))
	require.NoError(t, err)

	pkg, err := packager.StorePrepared(context.Background(), prepared)

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
