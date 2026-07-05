package permissionpolicy

import (
	"context"
	"database/sql"
)

const (
	HandlerACPToolRule       = "acp_tool_rule"
	HandlerSessionCostBudget = "session_cost_budget"
)

func (s *Service) OrgCostBudgetUSD(ctx context.Context, orgID int64) (float64, bool, error) {
	var max sql.NullFloat64
	err := s.db.WithContext(ctx).Model(&OrgRow{}).
		Select("max_usd").
		Where("organization_id = ? AND scope = 'org' AND policy_handler = ?", orgID, HandlerSessionCostBudget).
		Order("priority DESC, id ASC").
		Limit(1).
		Scan(&max).Error
	if err != nil || !max.Valid || max.Float64 <= 0 {
		return 0, false, err
	}
	return max.Float64, true, nil
}
