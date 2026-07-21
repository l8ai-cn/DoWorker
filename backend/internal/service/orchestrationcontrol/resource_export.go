package orchestrationcontrol

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
)

type ExportResourceRequest struct {
	Scope    control.Scope
	Target   control.ResourceTarget
	Revision int64
	Format   SourceFormat
}

type ResourceExport struct {
	Format  SourceFormat
	Content []byte
}

func (service *Service) ExportResource(
	ctx context.Context,
	request ExportResourceRequest,
) (ResourceExport, error) {
	head, err := service.GetResource(ctx, request.Scope, request.Target)
	if err != nil {
		return ResourceExport{}, err
	}
	revisionNumber := request.Revision
	if revisionNumber == 0 {
		revisionNumber = head.Revision
	}
	if revisionNumber <= 0 || revisionNumber > head.Revision {
		return ResourceExport{}, control.ErrInvalid
	}
	revision, err := service.repository.GetRevision(
		ctx,
		request.Scope,
		head.ID,
		revisionNumber,
	)
	if err != nil {
		return ResourceExport{}, err
	}
	if err := validateExportRevision(request.Scope, head, revisionNumber, revision); err != nil {
		return ResourceExport{}, err
	}
	manifest, err := decodeStoredExportManifest(revision.CanonicalManifest)
	if err != nil {
		return ResourceExport{}, control.ErrCorrupt
	}
	content, err := encodeResourceExport(request.Format, manifest)
	if err != nil {
		return ResourceExport{}, err
	}
	return ResourceExport{Format: request.Format, Content: content}, nil
}

func decodeStoredExportManifest(
	source json.RawMessage,
) (resource.Manifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var manifest resource.Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return resource.Manifest{}, err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return resource.Manifest{}, control.ErrCorrupt
	}
	return manifest, nil
}

func validateExportRevision(
	scope control.Scope,
	head control.ResourceHead,
	expected int64,
	revision control.ResourceRevision,
) error {
	if err := revision.Validate(scope); err != nil ||
		revision.OrganizationID != head.OrganizationID ||
		revision.ResourceID != head.ID ||
		revision.Identity != head.Identity ||
		revision.Revision != expected ||
		revision.Revision > head.Revision ||
		revision.Generation > head.Generation ||
		revision.ResourceVersion > head.ResourceVersion {
		return control.ErrCorrupt
	}
	return nil
}

func encodeResourceExport(
	format SourceFormat,
	manifest resource.Manifest,
) ([]byte, error) {
	switch format {
	case SourceFormatJSON:
		return resource.EncodeJSON(manifest)
	case SourceFormatYAML:
		return resource.EncodeYAML(manifest)
	default:
		return nil, control.ErrInvalid
	}
}
