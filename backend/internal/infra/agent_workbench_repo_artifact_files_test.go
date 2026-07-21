package infra

import (
	"context"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	sessionfiledomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/sessionfile"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAgentWorkbenchArtifactFilesAreTransactional(t *testing.T) {
	db, repo := agentWorkbenchPostgresRepository(t)
	sessionID := insertAgentWorkbenchSession(t, db)
	ctx := context.Background()
	require.NoError(t, appendArtifactFileRevision(
		ctx,
		repo,
		sessionID,
		0,
		1,
		nil,
	))

	staleFile := artifactFileRecord(
		sessionID,
		"file_stalerollback",
		"sessions/"+sessionID+"/artifacts/file_stalerollback/stale.pdf",
	)
	err := appendArtifactFileRevision(
		ctx,
		repo,
		sessionID,
		0,
		1,
		[]agentworkbench.ArtifactFile{staleFile},
	)

	require.ErrorIs(t, err, agentworkbench.ErrRevisionConflict)
	assertArtifactFileCount(t, db, staleFile.ID, 0)
}

func TestAgentWorkbenchArtifactFileConflictKeepsCanonicalObject(t *testing.T) {
	db, repo := agentWorkbenchPostgresRepository(t)
	sessionID := insertAgentWorkbenchSession(t, db)
	ctx := context.Background()
	canonical := artifactFileRecord(
		sessionID,
		"file_sameartifact",
		"sessions/"+sessionID+"/artifacts/file_sameartifact/canonical.pdf",
	)
	require.NoError(t, appendArtifactFileRevision(
		ctx,
		repo,
		sessionID,
		0,
		1,
		[]agentworkbench.ArtifactFile{canonical},
	))
	duplicate := canonical
	duplicate.MinioKey =
		"sessions/" + sessionID + "/artifacts/file_sameartifact/duplicate.pdf"
	require.NoError(t, appendArtifactFileRevision(
		ctx,
		repo,
		sessionID,
		1,
		2,
		[]agentworkbench.ArtifactFile{duplicate},
	))

	var stored sessionfiledomain.File
	require.NoError(t, db.Where(
		"id = ? AND session_id = ?",
		canonical.ID,
		sessionID,
	).Take(&stored).Error)
	require.Equal(t, canonical.MinioKey, stored.MinioKey)
}

func appendArtifactFileRevision(
	ctx context.Context,
	repo agentworkbench.Repository,
	sessionID string,
	expectedRevision uint64,
	revision uint64,
	files []agentworkbench.ArtifactFile,
) error {
	event := newAgentWorkbenchEvent(
		sessionID,
		"epoch-artifact",
		revision,
		revision,
		[]byte{byte(revision)},
	)
	_, err := repo.Append(ctx, agentworkbench.AppendRequest{
		SessionID: sessionID, ExpectedRevision: expectedRevision,
		ArtifactFiles: files,
		Events:        []agentworkbench.Event{event},
		Projection: agentworkbench.SessionState{
			SessionID: sessionID, StreamEpoch: "epoch-artifact",
			Revision: revision, LatestSequence: revision,
			Projection: []byte{byte(revision)}, Digest: agentWorkbenchDigest,
		},
	})
	return err
}

func artifactFileRecord(
	sessionID string,
	fileID string,
	minioKey string,
) agentworkbench.ArtifactFile {
	return agentworkbench.ArtifactFile{
		ID: fileID, SessionID: sessionID,
		Filename: "report.pdf", Bytes: 42,
		ContentType: "application/pdf", MinioKey: minioKey,
		CreatedAt: time.Now().UTC(),
	}
}

func assertArtifactFileCount(
	t *testing.T,
	db *gorm.DB,
	fileID string,
	expected int64,
) {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&sessionfiledomain.File{}).
		Where("id = ?", fileID).
		Count(&count).Error)
	require.Equal(t, expected, count)
}
