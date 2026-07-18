package agentworkbench

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	domainworkbench "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	sessionfiledomain "github.com/anthropics/agentsmesh/backend/internal/domain/sessionfile"
	"github.com/anthropics/agentsmesh/backend/internal/infra/storage"
	filesvc "github.com/anthropics/agentsmesh/backend/internal/service/file"
	sessionfilesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionfile"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type artifactMaterializerSandbox struct {
	commands []*runnerv1.SandboxFsCommand
	digest   string
	objects  *storage.MockStorage
	size     int64
	data     []byte
}

func (*artifactMaterializerSandbox) IsConnected(int64) bool {
	return true
}

func (sandbox *artifactMaterializerSandbox) Exec(
	_ context.Context,
	_ int64,
	command *runnerv1.SandboxFsCommand,
) (*runnerv1.SandboxFsResultEvent, error) {
	sandbox.commands = append(sandbox.commands, command)
	if sandbox.objects != nil {
		parsed, _ := url.Parse(command.GetPayload())
		key := strings.TrimPrefix(parsed.Path, "/")
		sandbox.objects.PutFileData(key, sandbox.data, "application/pdf")
	}
	return &runnerv1.SandboxFsResultEvent{
		ContentDigest: sandbox.digest,
		FileBytes:     sandbox.size,
	}, nil
}

func TestSessionFileArtifactMaterializerPersistsAndReusesArtifact(t *testing.T) {
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.AutoMigrate(&sessionfiledomain.File{}))
	objects := storage.NewMockStorage()
	sessionFiles := sessionfilesvc.NewService(
		db,
		filesvc.NewService(objects, config.StorageConfig{}),
	)
	sandbox := &artifactMaterializerSandbox{
		digest:  "sha256:c7c5c1d70c5dec4416ab6158afd0b223ef40c29b1dc1f97ed9428b94d4cadb1c",
		objects: objects,
		size:    8,
		data:    []byte("artifact"),
	}
	materializer := NewSessionFileArtifactMaterializer(sessionFiles, sandbox)
	batch := artifactMaterializerBatch(sandbox.digest, 8)

	first, err := materializer.Materialize(
		context.Background(),
		17,
		"session-1",
		"pod-1",
		batch,
	)
	require.NoError(t, err)
	resourceID := first.Batch.GetMutations()[0].GetArtifact().
		GetRepresentations()[0].GetTransport().GetResourceId()
	require.Regexp(t, `^session-file:file_[a-f0-9]{32}$`, resourceID)
	require.Len(t, first.Files, 1)
	require.Len(t, sandbox.commands, 1)
	require.Equal(t, "upload", sandbox.commands[0].GetOp())
	require.Equal(t, "output/report.pdf", sandbox.commands[0].GetPath())
	persistMaterializedArtifactFiles(t, db, first.Files)
	require.NoError(t, first.Reconcile(context.Background()))

	second, err := materializer.Materialize(
		context.Background(),
		17,
		"session-1",
		"pod-1",
		batch,
	)
	require.NoError(t, err)
	require.Equal(t, resourceID, second.Batch.GetMutations()[0].GetArtifact().
		GetRepresentations()[0].GetTransport().GetResourceId())
	require.Empty(t, second.Files)
	require.Len(t, sandbox.commands, 1)
}

func TestSessionFileArtifactMaterializerRejectsChangedUpload(t *testing.T) {
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.AutoMigrate(&sessionfiledomain.File{}))
	objects := storage.NewMockStorage()
	sessionFiles := sessionfilesvc.NewService(
		db,
		filesvc.NewService(objects, config.StorageConfig{}),
	)
	materializer := NewSessionFileArtifactMaterializer(
		sessionFiles,
		&artifactMaterializerSandbox{
			digest:  "sha256:changed",
			objects: objects,
			size:    8,
			data:    []byte("artifact"),
		},
	)

	_, err := materializer.Materialize(
		context.Background(),
		17,
		"session-1",
		"pod-1",
		artifactMaterializerBatch("sha256:artifact", 8),
	)

	require.ErrorContains(t, err, "artifact changed during upload")
	var count int64
	require.NoError(t, db.Model(&sessionfiledomain.File{}).Count(&count).Error)
	require.Zero(t, count)
}

func persistMaterializedArtifactFiles(
	t *testing.T,
	db *gorm.DB,
	files []domainworkbench.ArtifactFile,
) {
	t.Helper()
	for _, file := range files {
		require.NoError(t, db.Create(&sessionfiledomain.File{
			ID: file.ID, SessionID: file.SessionID,
			Filename: file.Filename, Bytes: file.Bytes,
			ContentType: file.ContentType, MinioKey: file.MinioKey,
			CreatedAt: file.CreatedAt,
		}).Error)
	}
}

func artifactMaterializerBatch(
	digest string,
	byteSize uint64,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	return &agentworkbenchv2.RunnerWorkbenchEventBatch{
		PodKey: "pod-1",
		Mutations: []*agentworkbenchv2.RunnerWorkbenchMutation{{
			Mutation: &agentworkbenchv2.RunnerWorkbenchMutation_Artifact{
				Artifact: &agentworkbenchv2.ArtifactDescriptor{
					ArtifactId: "artifact-1",
					Revision:   2,
					Filename:   "report.pdf",
					Representations: []*agentworkbenchv2.ArtifactRepresentation{{
						RepresentationId: "preview",
						Revision:         2,
						MediaType:        "application/pdf",
						Filename:         stringPointer("report.pdf"),
						Status:           agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
						ByteSize:         &byteSize,
						Digest:           &digest,
						Transport: &agentworkbenchv2.ArtifactTransport{
							Transport: &agentworkbenchv2.ArtifactTransport_ResourceId{
								ResourceId: "workspace:output/report.pdf",
							},
						},
					}},
				},
			},
		}},
	}
}
