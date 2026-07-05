package permissionpolicy

import (
	"context"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"gorm.io/gorm"
)

type Row struct {
	ToolPattern string `gorm:"column:tool_pattern"`
	PathPattern string `gorm:"column:path_pattern"`
	Verdict     string `gorm:"column:verdict"`
	Priority    int    `gorm:"column:priority"`
}

func (Row) TableName() string { return "permission_policies" }

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) SnapshotForPodCreate(ctx context.Context, orgID int64, agentSlug string) ([]*runnerv1.PolicyRuleSnapshot, error) {
	var rows []Row
	q := s.db.WithContext(ctx).
		Where("organization_id = ? AND scope = 'org' AND policy_handler = ?", orgID, HandlerACPToolRule).
		Order("priority DESC")
	if agentSlug != "" {
		q = q.Where("agent_slug IS NULL OR agent_slug = ?", agentSlug)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*runnerv1.PolicyRuleSnapshot, 0, len(rows))
	for i := range rows {
		out = append(out, &runnerv1.PolicyRuleSnapshot{
			ToolPattern: rows[i].ToolPattern,
			PathPattern: rows[i].PathPattern,
			Verdict:     rows[i].Verdict,
			Priority:    int32(rows[i].Priority),
		})
	}
	return out, nil
}
