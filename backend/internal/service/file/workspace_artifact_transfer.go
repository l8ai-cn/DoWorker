package file

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

const MaxWorkspaceArtifactBytes int64 = 128 << 20

type WorkspaceArtifactTransfer struct {
	Key         string
	PutURL      string
	ContentType string
	Size        int64
}

func (s *Service) PrepareWorkspaceArtifactTransfer(
	ctx context.Context,
	organizationID int64,
	filename string,
	contentType string,
	size int64,
) (*WorkspaceArtifactTransfer, error) {
	if s == nil || s.storage == nil {
		return nil, ErrStorageError
	}
	if size < 0 || size > MaxWorkspaceArtifactBytes {
		return nil, fmt.Errorf("%w: max workspace artifact size is %d MB", ErrFileTooLarge, MaxWorkspaceArtifactBytes>>20)
	}
	contentType = workspaceArtifactContentType(filename, contentType)
	key := workspaceArtifactKey(organizationID, filename)
	putURL, err := s.storage.InternalPresignPutURL(ctx, key, contentType, size, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStorageError, err)
	}
	return &WorkspaceArtifactTransfer{
		Key: key, PutURL: putURL, ContentType: contentType, Size: size,
	}, nil
}

func (s *Service) OpenWorkspaceArtifact(
	ctx context.Context,
	transfer *WorkspaceArtifactTransfer,
) (io.ReadCloser, int64, error) {
	if transfer == nil {
		return nil, 0, ErrStorageError
	}
	return s.OpenObject(ctx, transfer.Key)
}

func (s *Service) DeleteWorkspaceArtifact(
	ctx context.Context,
	transfer *WorkspaceArtifactTransfer,
) error {
	if s == nil || s.storage == nil || transfer == nil {
		return ErrStorageError
	}
	return s.storage.Delete(ctx, transfer.Key)
}

func workspaceArtifactKey(organizationID int64, filename string) string {
	ext := path.Ext(path.Base(filename))
	now := time.Now()
	return fmt.Sprintf(
		"workspace-artifacts/orgs/%d/%d/%02d/%s%s",
		organizationID,
		now.Year(),
		now.Month(),
		uuid.NewString(),
		ext,
	)
}

func workspaceArtifactContentType(filename, contentType string) string {
	contentType = strings.TrimSpace(strings.Split(contentType, ";")[0])
	if contentType != "" {
		return contentType
	}
	if detected := mime.TypeByExtension(strings.ToLower(path.Ext(filename))); detected != "" {
		return detected
	}
	return "application/octet-stream"
}
