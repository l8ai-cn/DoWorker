package tokenquota

import (
	"sort"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/tokenquota"
)

func accUser(m map[int64]*ScopeUsage, userID, tok int64, cost float64) {
	u := m[userID]
	if u == nil {
		id := userID
		u = &ScopeUsage{UserID: &id}
		m[userID] = u
	}
	u.Tokens += tok
	u.CostUSD += cost
}

func accModel(m map[string]*ScopeUsage, model string, tok int64, cost float64) {
	u := m[model]
	if u == nil {
		name := model
		u = &ScopeUsage{Model: &name}
		m[model] = u
	}
	u.Tokens += tok
	u.CostUSD += cost
}

func accKey(m map[int64]*ScopeUsage, keyID, tok int64, cost float64) {
	u := m[keyID]
	if u == nil {
		id := keyID
		u = &ScopeUsage{VirtualKey: &id}
		m[keyID] = u
	}
	u.Tokens += tok
	u.CostUSD += cost
}

func flatten(m map[int64]*ScopeUsage) []ScopeUsage {
	out := make([]ScopeUsage, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Tokens > out[j].Tokens })
	return out
}

func flattenModel(m map[string]*ScopeUsage) []ScopeUsage {
	out := make([]ScopeUsage, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Tokens > out[j].Tokens })
	return out
}

// overlayQuotas computes consumption for each configured quota scope and flags
// the ones over their limit.
func overlayQuotas(rows []usageRow, quotas []*domain.TokenQuota) []ScopeUsage {
	out := make([]ScopeUsage, 0, len(quotas))
	for _, q := range quotas {
		var tokens int64
		var cost float64
		for _, row := range rows {
			if !scopeMatches(q, row) {
				continue
			}
			tokens += row.tokens()
			cost += row.cost()
		}
		limit := q.LimitTokens
		out = append(out, ScopeUsage{
			UserID:  q.UserID,
			Model:   q.Model,
			Tokens:  tokens,
			CostUSD: cost,
			Limit:   &limit,
			Over:    tokens > q.LimitTokens,
		})
	}
	return out
}

func scopeMatches(q *domain.TokenQuota, row usageRow) bool {
	if q.UserID != nil && *q.UserID != row.UserID {
		return false
	}
	if q.Model != nil && *q.Model != row.Model {
		return false
	}
	return true
}
