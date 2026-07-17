package extension

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/extension"
	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"github.com/anthropics/agentsmesh/backend/internal/infra/storage"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
)

var (
	ErrNotFound         = errors.New("resource not found")
	ErrForbidden        = errors.New("access denied")
	ErrInvalidScope     = errors.New("invalid scope")
	ErrInvalidInput     = errors.New("invalid input")
	ErrAlreadyInstalled = errors.New("already installed")
)

// validateScope checks that scope is "org" or "user".
func validateScope(scope string) error {
	if scope != extension.ScopeOrg && scope != extension.ScopeUser {
		return fmt.Errorf("%w: %s, must be 'org' or 'user'", ErrInvalidScope, scope)
	}
	return nil
}

const presignedURLExpiry = 15 * time.Minute

// SkillCatalog is the read seam onto the unified skills catalog (owned by
// the skill service / skills table). The extension service consumes it for
// marketplace browsing and catalog installs.
type SkillCatalog interface {
	ListCatalog(ctx context.Context, orgID int64, query, category string) ([]skilldom.Skill, error)
	GetAnyByID(ctx context.Context, id int64) (*skilldom.Skill, error)
}

// Service provides extension management capabilities.
type Service struct {
	repo     extension.Repository
	storage  storage.Storage
	crypto   *crypto.Encryptor
	packager *SkillPackager
	catalog  SkillCatalog
}

func NewService(repo extension.Repository, storage storage.Storage, cryptoEncryptor *crypto.Encryptor) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
		crypto:  cryptoEncryptor,
	}
}

// SetSkillPackager sets the SkillPackager dependency.
// This uses a setter to avoid circular initialization issues.
func (s *Service) SetSkillPackager(p *SkillPackager) {
	s.packager = p
}

// SetSkillCatalog wires the unified skill catalog read seam.
func (s *Service) SetSkillCatalog(c SkillCatalog) {
	s.catalog = c
}

// SkillPackager exposes the configured packager so peer services (e.g. the
// git-backed skill service) can reuse package preparation and storage.
// Returns nil when no packager is configured.
func (s *Service) SkillPackager() *SkillPackager {
	return s.packager
}

// DecryptCredential decrypts a single credential string.
func (s *Service) DecryptCredential(encrypted string) (string, error) {
	return s.decryptCredential(encrypted)
}
