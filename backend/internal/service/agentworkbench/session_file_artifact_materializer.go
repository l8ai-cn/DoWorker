package agentworkbench

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	sessionfilesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionfile"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"google.golang.org/protobuf/proto"
)

const artifactUploadTimeout = 160 * time.Second

type ArtifactSandbox interface {
	IsConnected(int64) bool
	Exec(context.Context, int64, *runnerv1.SandboxFsCommand) (*runnerv1.SandboxFsResultEvent, error)
}

type SessionFileArtifactMaterializer struct {
	files    *sessionfilesvc.Service
	sandbox  ArtifactSandbox
	verifier ArtifactVerifier
}

func NewSessionFileArtifactMaterializer(
	files *sessionfilesvc.Service,
	sandbox ArtifactSandbox,
) *SessionFileArtifactMaterializer {
	return &SessionFileArtifactMaterializer{
		files: files, sandbox: sandbox,
		verifier: NewSessionFileArtifactVerifier(files),
	}
}

func (m *SessionFileArtifactMaterializer) Materialize(
	ctx context.Context,
	runnerID int64,
	sessionID string,
	podKey string,
	batch *agentworkbenchv2.RunnerWorkbenchEventBatch,
) (*ArtifactMaterialization, error) {
	if m == nil || m.files == nil || m.sandbox == nil ||
		runnerID <= 0 || sessionID == "" || podKey == "" || batch == nil {
		return nil, ErrIngressConfiguration
	}
	result := &ArtifactMaterialization{
		Batch:   proto.Clone(batch).(*agentworkbenchv2.RunnerWorkbenchEventBatch),
		service: m.files,
	}
	for _, mutation := range result.Batch.GetMutations() {
		artifact := mutation.GetArtifact()
		if artifact == nil {
			continue
		}
		for _, representation := range artifact.GetRepresentations() {
			if err := m.materializeRepresentation(
				ctx,
				runnerID,
				sessionID,
				podKey,
				artifact,
				representation,
				result,
			); err != nil {
				return nil, errors.Join(
					err,
					result.Abort(context.WithoutCancel(ctx)),
				)
			}
		}
	}
	return result, nil
}

func (m *SessionFileArtifactMaterializer) materializeRepresentation(
	ctx context.Context,
	runnerID int64,
	sessionID string,
	podKey string,
	artifact *agentworkbenchv2.ArtifactDescriptor,
	representation *agentworkbenchv2.ArtifactRepresentation,
	result *ArtifactMaterialization,
) error {
	if representation.GetStatus() !=
		agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY {
		return nil
	}
	resourceID := representation.GetTransport().GetResourceId()
	if resourceID == "" || strings.HasPrefix(resourceID, "session-file:") {
		return nil
	}
	if !strings.HasPrefix(resourceID, "workspace:") {
		return fmt.Errorf(
			"artifact %q representation %q uses non-durable resource %q",
			artifact.GetArtifactId(),
			representation.GetRepresentationId(),
			resourceID,
		)
	}
	path := strings.TrimPrefix(resourceID, "workspace:")
	if path == "" || representation.ByteSize == nil ||
		representation.GetDigest() == "" {
		return fmt.Errorf("artifact representation metadata is incomplete")
	}
	filename := representation.GetFilename()
	if filename == "" {
		filename = artifact.GetFilename()
	}
	upload, err := m.files.PrepareArtifact(ctx, sessionfilesvc.ArtifactInput{
		SessionID: sessionID, ArtifactID: artifact.GetArtifactId(),
		RepresentationID: representation.GetRepresentationId(),
		Revision:         artifact.GetRevision(), Filename: filepath.Base(filename),
		ContentType: representation.GetMediaType(),
		Digest:      representation.GetDigest(), Size: int64(representation.GetByteSize()),
	})
	if err != nil {
		return err
	}
	if !upload.Stored {
		if !m.sandbox.IsConnected(runnerID) {
			return fmt.Errorf("runner unavailable")
		}
		uploadCtx, cancel := context.WithTimeout(ctx, artifactUploadTimeout)
		result, execErr := m.sandbox.Exec(uploadCtx, runnerID, &runnerv1.SandboxFsCommand{
			Op: "upload", PodKey: podKey, Path: path, Payload: upload.PutURL,
		})
		cancel()
		if execErr != nil {
			return execErr
		}
		if result == nil || result.GetError() != "" {
			return fmt.Errorf("artifact upload failed: %s", resultError(result))
		}
		if result.GetFileBytes() != int64(representation.GetByteSize()) ||
			result.GetContentDigest() != representation.GetDigest() {
			return fmt.Errorf("artifact changed during upload")
		}
	}
	if err := m.verifier.Verify(ctx, upload.File, representation); err != nil {
		return err
	}
	if !upload.Stored {
		for _, file := range result.Files {
			if file.ID == upload.File.ID {
				return fmt.Errorf("artifact representation is duplicated")
			}
		}
		result.uploads = append(result.uploads, upload)
		result.Files = append(result.Files, artifactFileFromUpload(upload))
	}
	representation.Transport = &agentworkbenchv2.ArtifactTransport{
		Transport: &agentworkbenchv2.ArtifactTransport_ResourceId{
			ResourceId: "session-file:" + upload.File.ID,
		},
	}
	return nil
}

func resultError(result *runnerv1.SandboxFsResultEvent) string {
	if result == nil || result.GetError() == "" {
		return "runner returned no result"
	}
	return result.GetError()
}
