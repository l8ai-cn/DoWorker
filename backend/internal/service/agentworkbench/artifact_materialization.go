package agentworkbench

import (
	"context"
	"errors"

	domainworkbench "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	sessionfilesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessionfile"
	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
)

type ArtifactMaterialization struct {
	Batch   *agentworkbenchv2.RunnerWorkbenchEventBatch
	Files   []domainworkbench.ArtifactFile
	service *sessionfilesvc.Service
	uploads []*sessionfilesvc.ArtifactUpload
}

func (materialization *ArtifactMaterialization) Abort(
	ctx context.Context,
) error {
	if materialization == nil || materialization.service == nil {
		return nil
	}
	var cleanup error
	for _, upload := range materialization.uploads {
		cleanup = errors.Join(
			cleanup,
			materialization.service.AbortArtifactUpload(ctx, upload),
		)
	}
	return cleanup
}

func (materialization *ArtifactMaterialization) Reconcile(
	ctx context.Context,
) error {
	if materialization == nil || materialization.service == nil {
		return nil
	}
	var cleanup error
	for _, upload := range materialization.uploads {
		cleanup = errors.Join(
			cleanup,
			materialization.service.ReconcileArtifactUpload(ctx, upload),
		)
	}
	return cleanup
}

func artifactFileFromUpload(
	upload *sessionfilesvc.ArtifactUpload,
) domainworkbench.ArtifactFile {
	file := upload.File
	return domainworkbench.ArtifactFile{
		ID: file.ID, SessionID: file.SessionID,
		Filename: file.Filename, Bytes: file.Bytes,
		ContentType: file.ContentType, MinioKey: file.MinioKey,
		CreatedAt: file.CreatedAt,
	}
}
