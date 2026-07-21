package tokenquota

import (
	"context"
	"strings"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/tokenquota"
	"gorm.io/gorm"
)

// Service manages token quotas and computes report-only usage-vs-quota
// aggregates from pod_session_usage joined with pods.
type Service struct {
	repo domain.Repository
	db   *gorm.DB
}

func NewService(repo domain.Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

type UpsertInput struct {
	OrgID       int64
	UserID      *int64
	Model       *string
	LimitTokens int64
	Period      string
}

func (s *Service) Upsert(ctx context.Context, in UpsertInput) error {
	period := strings.TrimSpace(in.Period)
	if period == "" {
		period = domain.PeriodTotal
	}
	q := &domain.TokenQuota{
		OrganizationID: in.OrgID,
		UserID:         in.UserID,
		Model:          normalizeModel(in.Model),
		LimitTokens:    in.LimitTokens,
		Period:         period,
	}
	return s.repo.Upsert(ctx, q)
}

func (s *Service) List(ctx context.Context, orgID int64) ([]*domain.TokenQuota, error) {
	return s.repo.ListByOrg(ctx, orgID)
}

func (s *Service) Delete(ctx context.Context, id, orgID int64) error {
	return s.repo.Delete(ctx, id, orgID)
}

func normalizeModel(m *string) *string {
	if m == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*m)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
