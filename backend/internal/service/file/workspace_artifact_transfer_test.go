package file

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareWorkspaceArtifactTransferAcceptsLargeVideo(t *testing.T) {
	store := storage.NewMockStorage()
	service := NewService(store, config.StorageConfig{})

	transfer, err := service.PrepareWorkspaceArtifactTransfer(
		context.Background(),
		7,
		"seedance.mp4",
		"video/mp4",
		32<<20,
	)

	require.NoError(t, err)
	assert.Contains(t, transfer.Key, "workspace-artifacts/orgs/7/")
	assert.Contains(t, transfer.PutURL, "upload=true")
	assert.Equal(t, "video/mp4", transfer.ContentType)
	assert.Equal(t, int64(32<<20), transfer.Size)
}

func TestPrepareWorkspaceArtifactTransferRejectsOversizedFile(t *testing.T) {
	service := NewService(storage.NewMockStorage(), config.StorageConfig{})

	_, err := service.PrepareWorkspaceArtifactTransfer(
		context.Background(),
		7,
		"seedance.mp4",
		"video/mp4",
		MaxWorkspaceArtifactBytes+1,
	)

	require.ErrorIs(t, err, ErrFileTooLarge)
}

func TestDeleteWorkspaceArtifactRemovesTemporaryObject(t *testing.T) {
	store := storage.NewMockStorage()
	service := NewService(store, config.StorageConfig{})
	transfer := &WorkspaceArtifactTransfer{Key: "workspace-artifact"}
	store.PutFile(transfer.Key)

	require.NoError(t, service.DeleteWorkspaceArtifact(context.Background(), transfer))
	exists, err := store.Exists(context.Background(), transfer.Key)
	require.NoError(t, err)
	assert.False(t, exists)
}
