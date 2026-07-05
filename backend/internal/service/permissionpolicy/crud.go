package permissionpolicy

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("permission policy not found")

type OrgRow struct {
	ID             int64     `gorm:"column:id;primaryKey"`
	OrganizationID int64     `gorm:"column:organization_id"`
	Scope          string    `gorm:"column:scope"`
	SessionID      *string   `gorm:"column:session_id"`
	AgentSlug      *string   `gorm:"column:agent_slug"`
	PolicyHandler  string    `gorm:"column:policy_handler"`
	ToolPattern    string    `gorm:"column:tool_pattern"`
	PathPattern    *string   `gorm:"column:path_pattern"`
	Verdict        string    `gorm:"column:verdict"`
	Priority       int       `gorm:"column:priority"`
	MaxUSD         *float64  `gorm:"column:max_usd"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (OrgRow) TableName() string { return "permission_policies" }

type CreateInput struct {
	PolicyHandler string
	ToolPattern   string
	PathPattern   *string
	Verdict       string
	Priority      int
	AgentSlug     *string
	MaxUSD        *float64
}

func (s *Service) ListOrg(ctx context.Context, orgID int64) ([]OrgRow, error) {
	var rows []OrgRow
	err := s.db.WithContext(ctx).
		Where("organization_id = ? AND scope = 'org'", orgID).
		Order("priority DESC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (s *Service) CreateOrg(ctx context.Context, orgID int64, in CreateInput) (*OrgRow, error) {
	handler := in.PolicyHandler
	if handler == "" {
		handler = HandlerACPToolRule
	}
	row := OrgRow{
		OrganizationID: orgID,
		Scope:          "org",
		PolicyHandler:  handler,
		AgentSlug:      in.AgentSlug,
		ToolPattern:    in.ToolPattern,
		PathPattern:    in.PathPattern,
		Verdict:        in.Verdict,
		Priority:       in.Priority,
		MaxUSD:         in.MaxUSD,
	}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) DeleteOrg(ctx context.Context, orgID, id int64) error {
	res := s.db.WithContext(ctx).
		Where("id = ? AND organization_id = ? AND scope = 'org'", id, orgID).
		Delete(&OrgRow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func ParsePolicyID(raw string) (int64, error) {
	const prefix = "pol_"
	if len(raw) <= len(prefix) || raw[:len(prefix)] != prefix {
		return 0, fmt.Errorf("invalid policy id")
	}
	var id int64
	if _, err := fmt.Sscanf(raw[len(prefix):], "%d", &id); err != nil {
		return 0, err
	}
	return id, nil
}
