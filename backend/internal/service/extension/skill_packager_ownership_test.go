package extension

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func TestPrepareCatalogFromDirUsesIsolatedNamespace(t *testing.T) {
	packager := NewSkillPackager(nil, &svcMockStorage{})
	dir := ownershipTestSkillDir(t)

	direct, err := packager.PrepareFromDir(context.Background(), dir)
	require.NoError(t, err)
	catalog, err := packager.PrepareCatalogFromDir(
		context.Background(),
		dir,
		"am-skills/org7-ownership-test",
	)
	require.NoError(t, err)

	assert.Equal(t, direct.ContentSha, catalog.ContentSha)
	assert.NotEqual(t, direct.StorageKey, catalog.StorageKey)
	assert.Equal(t, "skills/direct/ownership-test/"+direct.ContentSha+".tar.gz", direct.StorageKey)
	parts := strings.Split(catalog.StorageKey, "/")
	require.Len(t, parts, 4)
	assert.Equal(t, []string{"skills", "catalog"}, parts[:2])
	assert.Regexp(t, `^[a-f0-9]{64}$`, parts[2])
	assert.Equal(t, catalog.ContentSha+".tar.gz", parts[3])
}

func TestPrepareCatalogFromDirSeparatesInternalRepositories(t *testing.T) {
	packager := NewSkillPackager(nil, &svcMockStorage{})
	dir := ownershipTestSkillDir(t)

	first, err := packager.PrepareCatalogFromDir(
		context.Background(),
		dir,
		"am-skills/org7-ownership-test",
	)
	require.NoError(t, err)
	second, err := packager.PrepareCatalogFromDir(
		context.Background(),
		dir,
		"../../am-skills/org8-ownership-test",
	)
	require.NoError(t, err)

	assert.Equal(t, first.ContentSha, second.ContentSha)
	assert.NotEqual(t, first.StorageKey, second.StorageKey)
	assert.NotContains(t, second.StorageKey, "..")
}

func TestStorePreparedReusesExistingObject(t *testing.T) {
	var uploads int
	store := &svcMockStorage{
		existsFn: func(context.Context, string) (bool, error) {
			return true, nil
		},
		downloadFn: func(context.Context, string) (io.ReadCloser, int64, error) {
			return io.NopCloser(strings.NewReader("stored package")), 731, nil
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
	assert.Equal(t, int64(731), pkg.PackageSize)
	assert.Zero(t, uploads)
}

func TestStorePreparedRejectsEmptyExistingObject(t *testing.T) {
	store := &svcMockStorage{
		existsFn: func(context.Context, string) (bool, error) {
			return true, nil
		},
		downloadFn: func(context.Context, string) (io.ReadCloser, int64, error) {
			return io.NopCloser(strings.NewReader("")), 0, nil
		},
	}
	packager := NewSkillPackager(nil, store)
	prepared, err := packager.PrepareFromDir(context.Background(), ownershipTestSkillDir(t))
	require.NoError(t, err)

	_, err = packager.StorePrepared(context.Background(), prepared)

	require.EqualError(t, err, "existing package is empty")
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
