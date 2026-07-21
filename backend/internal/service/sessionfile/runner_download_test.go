package sessionfile

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/sessionfile"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/storage"
	fileservice "github.com/l8ai-cn/agentcloud/backend/internal/service/file"
	"github.com/stretchr/testify/require"
)

func TestRunnerDownloadURLUsesInternalStorageURL(t *testing.T) {
	files := fileservice.NewService(storage.NewMockStorage(), config.StorageConfig{})
	service := NewService(nil, files)

	url, err := service.RunnerDownloadURL(
		context.Background(),
		&domain.File{MinioKey: "sessions/session-1/files/file-1.png"},
	)

	require.NoError(t, err)
	require.Contains(t, url, "mock-storage.example.com/sessions/session-1/files/file-1.png")
}
