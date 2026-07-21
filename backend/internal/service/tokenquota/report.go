package tokenquota

import (
	"context"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/tokenquota"
)

// ScopeUsage is consumption for one aggregation bucket, optionally paired with
// a matching quota limit and an over-limit flag.
type ScopeUsage struct {
	UserID     *int64  `json:"user_id,omitempty"`
	Model      *string `json:"model,omitempty"`
	VirtualKey *int64  `json:"virtual_api_key_id,omitempty"`
	Tokens     int64   `json:"tokens"`
	CostUSD    float64 `json:"cost_usd"`
	Limit      *int64  `json:"limit_tokens,omitempty"`
	Over       bool    `json:"over_limit"`
}

// Report is the org-wide usage-vs-quota snapshot.
type Report struct {
	TotalTokens  int64        `json:"total_tokens"`
	TotalCostUSD float64      `json:"total_cost_usd"`
	ByUser       []ScopeUsage `json:"by_user"`
	ByModel      []ScopeUsage `json:"by_model"`
	ByVirtualKey []ScopeUsage `json:"by_virtual_key"`
	Quotas       []ScopeUsage `json:"quotas"`
}

type usageRow struct {
	UserID              int64
	KeyID               *int64
	Model               string
	InputTokens         int64
	OutputTokens        int64
	CacheReadTokens     int64
	CacheCreationTokens int64
	InputPerMillion     *float64
	OutputPerMillion    *float64
}

func (r usageRow) tokens() int64 {
	return r.InputTokens + r.OutputTokens + r.CacheReadTokens + r.CacheCreationTokens
}

func (r usageRow) cost() float64 {
	if r.InputPerMillion == nil || r.OutputPerMillion == nil {
		return 0
	}
	inPrice := *r.InputPerMillion
	outPrice := *r.OutputPerMillion
	return float64(r.InputTokens)/1e6*inPrice + float64(r.OutputTokens)/1e6*outPrice
}

// Report aggregates live session usage for the org and overlays configured
// quotas. Consumption is computed on read (no counters), so it always
// reflects the authoritative pod_session_usage table.
func (s *Service) Report(ctx context.Context, orgID int64) (*Report, error) {
	rows, err := s.usageRows(ctx, orgID)
	if err != nil {
		return nil, err
	}
	quotas, err := s.repo.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return buildReport(rows, quotas), nil
}

func (s *Service) usageRows(ctx context.Context, orgID int64) ([]usageRow, error) {
	var rows []usageRow
	err := s.db.WithContext(ctx).
		Table("pod_session_usage psu").
		Select(`p.created_by_id AS user_id,
			p.virtual_api_key_id AS key_id,
			psu.model AS model,
			psu.input_tokens, psu.output_tokens,
			psu.cache_read_tokens, psu.cache_creation_tokens,
			mp.input_per_million, mp.output_per_million`).
		Joins("JOIN pods p ON p.pod_key = psu.pod_key").
		Joins("LEFT JOIN model_prices mp ON mp.model = psu.model").
		Where("p.organization_id = ?", orgID).
		Scan(&rows).Error
	return rows, err
}

func buildReport(rows []usageRow, quotas []*domain.TokenQuota) *Report {
	rep := &Report{}
	byUser := map[int64]*ScopeUsage{}
	byModel := map[string]*ScopeUsage{}
	byKey := map[int64]*ScopeUsage{}

	for _, row := range rows {
		tok, cost := row.tokens(), row.cost()
		rep.TotalTokens += tok
		rep.TotalCostUSD += cost
		accUser(byUser, row.UserID, tok, cost)
		accModel(byModel, row.Model, tok, cost)
		if row.KeyID != nil {
			accKey(byKey, *row.KeyID, tok, cost)
		}
	}

	rep.ByUser = flatten(byUser)
	rep.ByModel = flattenModel(byModel)
	rep.ByVirtualKey = flatten(byKey)
	rep.Quotas = overlayQuotas(rows, quotas)
	return rep
}
