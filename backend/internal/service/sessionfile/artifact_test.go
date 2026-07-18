package sessionfile

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/sessionfile"
	"github.com/anthropics/agentsmesh/backend/internal/infra/storage"
	filesvc "github.com/anthropics/agentsmesh/backend/internal/service/file"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/stretchr/testify/require"
)

func TestPrepareArtifactUsesUniqueObjectKeyPerAttempt(t *testing.T) {
	service, _ := artifactTestService(t)
	input := artifactTestInput()

	first, err := service.PrepareArtifact(context.Background(), input)
	require.NoError(t, err)
	second, err := service.PrepareArtifact(context.Background(), input)
	require.NoError(t, err)

	require.Equal(t, first.File.ID, second.File.ID)
	require.NotEqual(t, first.File.MinioKey, second.File.MinioKey)
}

func TestAbortArtifactUploadDeletesCandidateObject(t *testing.T) {
	service, objects := artifactTestService(t)
	upload, err := service.PrepareArtifact(
		context.Background(),
		artifactTestInput(),
	)
	require.NoError(t, err)
	objects.PutFile(upload.File.MinioKey)

	require.NoError(t, service.AbortArtifactUpload(
		context.Background(),
		upload,
	))
	require.Zero(t, objects.FileCount())
}

func TestReconcileArtifactUploadDeletesLosingObject(t *testing.T) {
	service, objects := artifactTestService(t)
	input := artifactTestInput()
	winner, err := service.PrepareArtifact(context.Background(), input)
	require.NoError(t, err)
	loser, err := service.PrepareArtifact(context.Background(), input)
	require.NoError(t, err)
	require.NoError(t, service.db.Create(winner.File).Error)
	objects.PutFile(winner.File.MinioKey)
	objects.PutFile(loser.File.MinioKey)

	require.NoError(t, service.ReconcileArtifactUpload(
		context.Background(),
		loser,
	))

	_, winnerExists := objects.GetFile(winner.File.MinioKey)
	_, loserExists := objects.GetFile(loser.File.MinioKey)
	require.True(t, winnerExists)
	require.False(t, loserExists)
}

func artifactTestService(t *testing.T) (*Service, *storage.MockStorage) {
	t.Helper()
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.File{}))
	objects := storage.NewMockStorage()
	return NewService(
		db,
		filesvc.NewService(objects, config.StorageConfig{}),
	), objects
}

func artifactTestInput() ArtifactInput {
	return ArtifactInput{
		SessionID: "session-1", ArtifactID: "artifact-1",
		RepresentationID: "preview", Revision: 2,
		Filename: "report.pdf", ContentType: "application/pdf",
		Digest: "sha256:artifact", Size: 42,
	}
}
