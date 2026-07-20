package sessionapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	workbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	sessionfiledomain "github.com/anthropics/agentsmesh/backend/internal/domain/sessionfile"
	"github.com/anthropics/agentsmesh/backend/internal/infra/storage"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	filesvc "github.com/anthropics/agentsmesh/backend/internal/service/file"
	sessionfilesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionfile"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func TestArtifactRepresentationRequiresSessionOwner(t *testing.T) {
	deps := readOnlySessionPermissionDeps(t)
	response := artifactRepresentationRequest(t, deps, 12)

	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestArtifactRepresentationOwnerReachesArtifactService(t *testing.T) {
	deps := ownerSessionPermissionDeps(t)
	response := artifactRepresentationRequest(t, deps, 11)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestArtifactDownloadGrantScope(t *testing.T) {
	now := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)
	minimum := uint64(3)
	maximum := uint64(3)
	expires := now.Add(time.Minute).Format(time.RFC3339Nano)
	artifact := &agentworkbenchv2.ArtifactDescriptor{
		Revision: 3,
		Grants: []*agentworkbenchv2.ArtifactGrant{{
			GrantId:           "grant-1",
			Actions:           []string{"artifact.download"},
			RepresentationIds: []string{"preview"},
			MinimumRevision:   &minimum,
			MaximumRevision:   &maximum,
			ExpiresAt:         &expires,
		}},
	}

	assert.True(t, artifactDownloadGranted(artifact, "preview", now))
	assert.False(t, artifactDownloadGranted(artifact, "original", now))
	artifact.Grants[0].Actions = []string{"image.edit"}
	assert.False(t, artifactDownloadGranted(artifact, "preview", now))
	artifact.Grants[0].Actions = []string{"artifact.download"}
	assert.False(t, artifactDownloadGranted(
		artifact,
		"preview",
		now.Add(2*time.Minute),
	))
}

func TestArtifactRepresentationServesRangeFromSessionFile(t *testing.T) {
	deps, db := ownerArtifactRepresentationDeps(t)
	objects := storage.NewMockStorage()
	counted := &artifactRangeStorage{Storage: objects}
	deps.SessionFiles = sessionfilesvc.NewService(
		db,
		filesvc.NewService(counted, config.StorageConfig{}),
	)
	deps.WorkbenchRepo = artifactRepresentationRepo(t)
	require.NoError(t, db.AutoMigrate(&sessionfiledomain.File{}))
	require.NoError(t, db.Create(&sessionfiledomain.File{
		ID: "file_artifact", SessionID: "conv_read", Filename: "seedance.mp4",
		Bytes: 10, ContentType: "video/mp4", MinioKey: "objects/seedance.mp4",
	}).Error)
	objects.PutFileData("objects/seedance.mp4", []byte("0123456789"), "video/mp4")

	response := artifactRepresentationRequestWithRange(t, deps, "bytes=2-5")

	assert.Equal(t, http.StatusPartialContent, response.Code)
	assert.Equal(t, "bytes 2-5/10", response.Header().Get("Content-Range"))
	assert.Equal(t, "bytes", response.Header().Get("Accept-Ranges"))
	assert.Equal(t, "2345", response.Body.String())
	assert.Zero(t, counted.downloads)
	assert.Equal(t, 1, counted.ranges)
}

func TestArtifactRepresentationRejectsUnsatisfiableRange(t *testing.T) {
	deps, db := ownerArtifactRepresentationDeps(t)
	objects := storage.NewMockStorage()
	deps.SessionFiles = sessionfilesvc.NewService(
		db,
		filesvc.NewService(objects, config.StorageConfig{}),
	)
	deps.WorkbenchRepo = artifactRepresentationRepo(t)
	require.NoError(t, db.AutoMigrate(&sessionfiledomain.File{}))
	require.NoError(t, db.Create(&sessionfiledomain.File{
		ID: "file_artifact", SessionID: "conv_read", Filename: "seedance.mp4",
		Bytes: 10, ContentType: "video/mp4", MinioKey: "objects/seedance.mp4",
	}).Error)
	objects.PutFileData("objects/seedance.mp4", []byte("0123456789"), "video/mp4")

	response := artifactRepresentationRequestWithRange(t, deps, "bytes=20-25")

	assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, response.Code)
	assert.Equal(t, "bytes */10", response.Header().Get("Content-Range"))
}

func artifactRepresentationRequest(
	t *testing.T,
	deps *Deps,
	userID int64,
) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/v1/sessions/conv_read/artifacts/representation",
		nil,
	)
	ctx.Params = gin.Params{{Key: "id", Value: "conv_read"}}
	ctx.Set("tenant", &middleware.TenantContext{
		OrganizationID: 21,
		UserID:         userID,
	})
	deps.handleGetSessionArtifactRepresentation(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}

func artifactRepresentationRequestWithRange(
	t *testing.T,
	deps *Deps,
	rangeHeader string,
) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/v1/sessions/conv_read/artifacts/content?artifact_id=artifact-1&representation_id=playable&revision=1&digest=sha256:artifact",
		nil,
	)
	ctx.Request.Header.Set("Range", rangeHeader)
	ctx.Params = gin.Params{{Key: "id", Value: "conv_read"}}
	ctx.Set("tenant", &middleware.TenantContext{
		OrganizationID: 21,
		UserID:         11,
	})
	deps.handleGetSessionArtifactRepresentation(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}

func ownerArtifactRepresentationDeps(t *testing.T) (*Deps, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := setupSessionByPodTestDB(t)
	insertSessionByPodTestRow(t, db, "conv_read", "read-pod", 21, 11)
	return &Deps{Sessions: sessionsvc.NewService(db)}, db
}

type artifactRepresentationSnapshotRepo struct {
	state *workbenchdomain.SessionState
}

type artifactRangeStorage struct {
	storage.Storage
	downloads int
	ranges    int
}

func (s *artifactRangeStorage) Download(
	ctx context.Context,
	key string,
) (io.ReadCloser, int64, error) {
	s.downloads++
	return s.Storage.Download(ctx, key)
}

func (s *artifactRangeStorage) DownloadRange(
	ctx context.Context,
	key string,
	start int64,
	end int64,
) (io.ReadCloser, int64, error) {
	s.ranges++
	return s.Storage.DownloadRange(ctx, key, start, end)
}

func (repo artifactRepresentationSnapshotRepo) Append(
	context.Context,
	workbenchdomain.AppendRequest,
) (workbenchdomain.AppendResult, error) {
	return workbenchdomain.AppendResult{}, nil
}

func (repo artifactRepresentationSnapshotRepo) GetSnapshot(
	context.Context,
	string,
) (*workbenchdomain.SessionState, error) {
	return repo.state, nil
}

func (repo artifactRepresentationSnapshotRepo) ListAfter(
	context.Context,
	string,
	string,
	uint64,
	int,
) ([]workbenchdomain.Event, error) {
	return nil, nil
}

func (repo artifactRepresentationSnapshotRepo) PutCommandReceipt(
	context.Context,
	workbenchdomain.CommandReceipt,
) (*workbenchdomain.CommandReceipt, error) {
	return nil, nil
}

func (repo artifactRepresentationSnapshotRepo) GetCommandReceipt(
	context.Context,
	string,
	string,
) (*workbenchdomain.CommandReceipt, error) {
	return nil, nil
}

func artifactRepresentationRepo(t *testing.T) artifactRepresentationSnapshotRepo {
	t.Helper()
	bytes := uint64(10)
	digest := "sha256:artifact"
	snapshot := &agentworkbenchv2.SessionSnapshot{
		SessionId: "conv_read",
		Capabilities: &agentworkbenchv2.SupportCapabilities{
			ArtifactOperations: []string{"artifact.download"},
		},
		Artifacts: []*agentworkbenchv2.ArtifactDescriptor{{
			ArtifactId: "artifact-1", Revision: 1, Filename: "seedance.mp4",
			Grants: []*agentworkbenchv2.ArtifactGrant{{
				GrantId: "grant-1", Actions: []string{"artifact.download"},
			}},
			Representations: []*agentworkbenchv2.ArtifactRepresentation{{
				RepresentationId: "playable", Revision: 1,
				MediaType: "video/mp4", Status: agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
				ByteSize: &bytes, Digest: &digest,
				Transport: &agentworkbenchv2.ArtifactTransport{
					Transport: &agentworkbenchv2.ArtifactTransport_ResourceId{
						ResourceId: "session-file:file_artifact",
					},
				},
			}},
		}},
	}
	raw, err := proto.Marshal(snapshot)
	require.NoError(t, err)
	return artifactRepresentationSnapshotRepo{
		state: &workbenchdomain.SessionState{
			SessionID:  "conv_read",
			Projection: raw,
		},
	}
}
