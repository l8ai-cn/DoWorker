package permissionpolicy

import (
	"context"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (s *Service) ListSession(ctx context.Context, orgID int64, sessionID string) ([]OrgRow, error) {
	var rows []OrgRow
	err := s.db.WithContext(ctx).
		Where("organization_id = ? AND scope = 'session' AND session_id = ?", orgID, sessionID).
		Order("priority DESC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (s *Service) CreateSession(ctx context.Context, orgID int64, sessionID string, in CreateInput) (*OrgRow, error) {
	sid := sessionID
	row := OrgRow{
		OrganizationID: orgID, Scope: "session", SessionID: &sid,
		PolicyHandler: HandlerACPToolRule, AgentSlug: in.AgentSlug,
		ToolPattern: in.ToolPattern, PathPattern: in.PathPattern,
		Verdict: in.Verdict, Priority: in.Priority,
	}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) DeleteSession(ctx context.Context, orgID int64, sessionID string, id int64) error {
	res := s.db.WithContext(ctx).
		Where("id = ? AND organization_id = ? AND scope = 'session' AND session_id = ?", id, orgID, sessionID).
		Delete(&OrgRow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

type PatchInput struct {
	ToolPattern *string
	PathPattern *string
	Verdict     *string
	Priority    *int
}

func (s *Service) UpdateOrg(ctx context.Context, orgID, id int64, in PatchInput) (*OrgRow, error) {
	updates := map[string]any{"updated_at": time.Now()}
	if in.ToolPattern != nil {
		updates["tool_pattern"] = *in.ToolPattern
	}
	if in.PathPattern != nil {
		updates["path_pattern"] = *in.PathPattern
	}
	if in.Verdict != nil {
		updates["verdict"] = *in.Verdict
	}
	if in.Priority != nil {
		updates["priority"] = *in.Priority
	}
	res := s.db.WithContext(ctx).Model(&OrgRow{}).
		Where("id = ? AND organization_id = ? AND scope = 'org'", id, orgID).
		Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	var row OrgRow
	if err := s.db.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) SnapshotForSession(ctx context.Context, orgID int64, sessionID, agentSlug string) ([]*runnerv1.PolicyRuleSnapshot, error) {
	orgRules, err := s.SnapshotForPodCreate(ctx, orgID, agentSlug)
	if err != nil {
		return nil, err
	}
	var rows []Row
	err = s.db.WithContext(ctx).
		Where("organization_id = ? AND scope = 'session' AND session_id = ? AND policy_handler = ?",
			orgID, sessionID, HandlerACPToolRule).
		Order("priority DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := append([]*runnerv1.PolicyRuleSnapshot{}, orgRules...)
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
