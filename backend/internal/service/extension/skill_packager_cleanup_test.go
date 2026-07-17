package extension

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPackagerDeletePackageDelegatesToStorage(t *testing.T) {
	var deletedKey string
	storage := &svcMockStorage{
		deleteFn: func(_ context.Context, key string) error {
			deletedKey = key
			return nil
		},
	}
	packager := NewSkillPackager(nil, storage)

	err := packager.DeletePackage(context.Background(), "skills/direct/video/package.tar.gz")

	require.NoError(t, err)
	assert.Equal(t, "skills/direct/video/package.tar.gz", deletedKey)
}

func TestSkillPackagerDeletePackageWrapsStorageError(t *testing.T) {
	storage := &svcMockStorage{
		deleteFn: func(context.Context, string) error {
			return errors.New("storage unavailable")
		},
	}
	packager := NewSkillPackager(nil, storage)

	err := packager.DeletePackage(context.Background(), "skills/direct/video/package.tar.gz")

	require.ErrorContains(t, err, "storage unavailable")
	require.ErrorContains(t, err, "skills/direct/video/package.tar.gz")
}
