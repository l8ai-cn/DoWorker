package sessionfile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"strconv"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/sessionfile"
)

type ArtifactUpload struct {
	File   *domain.File
	PutURL string
	Stored bool
}

type ArtifactInput struct {
	SessionID        string
	ArtifactID       string
	RepresentationID string
	Revision         uint64
	Filename         string
	ContentType      string
	Digest           string
	Size             int64
}

func (s *Service) PrepareArtifact(
	ctx context.Context,
	in ArtifactInput,
) (*ArtifactUpload, error) {
	if s == nil || s.files == nil || s.db == nil {
		return nil, fmt.Errorf("session file service unavailable")
	}
	if in.SessionID == "" || in.ArtifactID == "" ||
		in.RepresentationID == "" || in.Digest == "" ||
		in.Size < 0 {
		return nil, fmt.Errorf("invalid artifact upload")
	}
	id := artifactFileID(in)
	row := &domain.File{
		ID: id, SessionID: in.SessionID, Filename: in.Filename,
		Bytes: in.Size, ContentType: normalizeContentType(in.ContentType, in.Filename),
		CreatedAt: time.Now().UTC(),
	}
	stored, err := s.GetForSession(ctx, in.SessionID, id)
	if err == nil {
		if !sameArtifactIdentity(stored, row) {
			return nil, fmt.Errorf("artifact file identity conflict")
		}
		return &ArtifactUpload{File: stored, Stored: true}, nil
	}
	if err != ErrNotFound {
		return nil, err
	}
	uploadID, err := NewID()
	if err != nil {
		return nil, err
	}
	row.MinioKey = artifactObjectKey(
		in.SessionID,
		id,
		uploadID,
		in.Filename,
	)
	putURL, err := s.files.PrepareInternalUpload(
		ctx,
		row.MinioKey,
		row.ContentType,
		row.Bytes,
	)
	if err != nil {
		return nil, err
	}
	return &ArtifactUpload{File: row, PutURL: putURL}, nil
}

func (s *Service) AbortArtifactUpload(
	ctx context.Context,
	upload *ArtifactUpload,
) error {
	if s == nil || s.files == nil || upload == nil ||
		upload.Stored || upload.File == nil {
		return nil
	}
	return s.files.DeleteObject(ctx, upload.File.MinioKey)
}

func (s *Service) ReconcileArtifactUpload(
	ctx context.Context,
	upload *ArtifactUpload,
) error {
	if s == nil || upload == nil || upload.Stored || upload.File == nil {
		return nil
	}
	stored, err := s.GetForSession(
		ctx,
		upload.File.SessionID,
		upload.File.ID,
	)
	if err == ErrNotFound {
		return s.AbortArtifactUpload(ctx, upload)
	}
	if err != nil {
		return err
	}
	if !sameArtifactIdentity(stored, upload.File) {
		return fmt.Errorf("artifact file identity conflict")
	}
	if stored.MinioKey == upload.File.MinioKey {
		return nil
	}
	return s.AbortArtifactUpload(ctx, upload)
}

func artifactFileID(in ArtifactInput) string {
	hash := sha256.Sum256([]byte(
		in.SessionID + "\x00" + in.ArtifactID + "\x00" +
			in.RepresentationID + "\x00" +
			strconv.FormatUint(in.Revision, 10) + "\x00" + in.Digest,
	))
	return "file_" + hex.EncodeToString(hash[:16])
}

func sameArtifactIdentity(left, right *domain.File) bool {
	return left != nil && right != nil &&
		left.ID == right.ID &&
		left.SessionID == right.SessionID &&
		left.Filename == right.Filename &&
		left.Bytes == right.Bytes &&
		left.ContentType == right.ContentType
}

func artifactObjectKey(
	sessionID string,
	fileID string,
	uploadID string,
	filename string,
) string {
	extension := path.Ext(path.Base(filename))
	return fmt.Sprintf(
		"sessions/%s/artifacts/%s/%s%s",
		sessionID,
		fileID,
		uploadID,
		extension,
	)
}
