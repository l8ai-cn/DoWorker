package sessionfile

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/sessionfile"
	fileservice "github.com/anthropics/agentsmesh/backend/internal/service/file"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("session file not found")

const runnerDownloadURLExpiry = 5 * time.Minute

type Service struct {
	db    *gorm.DB
	files *fileservice.Service
}

func NewService(db *gorm.DB, files *fileservice.Service) *Service {
	return &Service{db: db, files: files}
}

func NewID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("file id: %w", err)
	}
	return "file_" + hex.EncodeToString(b), nil
}

type CreateInput struct {
	SessionID   string
	Filename    string
	ContentType string
	Reader      io.Reader
	Size        int64
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.File, error) {
	if s == nil || s.files == nil || s.db == nil {
		return nil, fmt.Errorf("session file service unavailable")
	}
	id, err := NewID()
	if err != nil {
		return nil, err
	}
	ct := normalizeContentType(in.ContentType, in.Filename)
	if err := s.files.ValidateUpload(in.Size, ct); err != nil {
		return nil, err
	}
	key := sessionObjectKey(in.SessionID, id, in.Filename)
	if err := s.files.PutObject(ctx, key, in.Reader, in.Size, ct); err != nil {
		return nil, err
	}
	row := &domain.File{
		ID: id, SessionID: in.SessionID, Filename: in.Filename,
		Bytes: in.Size, ContentType: ct, MinioKey: key, CreatedAt: time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (s *Service) GetForSession(ctx context.Context, sessionID, fileID string) (*domain.File, error) {
	var row domain.File
	err := s.db.WithContext(ctx).Where("id = ? AND session_id = ?", fileID, sessionID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) Open(ctx context.Context, row *domain.File) (io.ReadCloser, int64, error) {
	if s == nil || s.files == nil || row == nil {
		return nil, 0, fmt.Errorf("session file service unavailable")
	}
	return s.files.OpenObject(ctx, row.MinioKey)
}

func (s *Service) RunnerDownloadURL(ctx context.Context, row *domain.File) (string, error) {
	if s == nil || s.files == nil || row == nil || row.MinioKey == "" {
		return "", fmt.Errorf("session file service unavailable")
	}
	return s.files.GetInternalURL(ctx, row.MinioKey, runnerDownloadURLExpiry)
}

func sessionObjectKey(sessionID, fileID, filename string) string {
	ext := path.Ext(filename)
	return fmt.Sprintf("sessions/%s/files/%s%s", sessionID, fileID, ext)
}

func normalizeContentType(ct, filename string) string {
	ct = strings.TrimSpace(strings.Split(ct, ";")[0])
	if ct != "" && ct != "application/octet-stream" {
		return ct
	}
	switch strings.ToLower(path.Ext(filename)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	default:
		if ct == "" {
			return "application/octet-stream"
		}
		return ct
	}
}
