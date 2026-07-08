package tokenquota

import (
	"testing"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/tokenquota"
)

func i64(v int64) *int64    { return &v }
func str(v string) *string  { return &v }
func f64(v float64) *float64 { return &v }

func TestBuildReportAggregatesAndOverlaysQuotas(t *testing.T) {
	key := int64(7)
	rows := []usageRow{
		{UserID: 1, Model: "m1", InputTokens: 100, OutputTokens: 50},
		{UserID: 2, Model: "m1", InputTokens: 10, OutputTokens: 5, KeyID: &key},
	}
	quotas := []*domain.TokenQuota{
		{UserID: i64(1), LimitTokens: 100},  // user1 used 150 -> over
		{Model: str("m1"), LimitTokens: 1000}, // model used 165 -> ok
	}

	rep := buildReport(rows, quotas)

	if rep.TotalTokens != 165 {
		t.Fatalf("total tokens = %d, want 165", rep.TotalTokens)
	}
	if len(rep.ByUser) != 2 {
		t.Fatalf("by_user len = %d, want 2", len(rep.ByUser))
	}
	if len(rep.ByVirtualKey) != 1 || rep.ByVirtualKey[0].Tokens != 15 {
		t.Fatalf("by_virtual_key wrong: %+v", rep.ByVirtualKey)
	}

	var userQuota, modelQuota *ScopeUsage
	for i := range rep.Quotas {
		q := &rep.Quotas[i]
		if q.UserID != nil && *q.UserID == 1 {
			userQuota = q
		}
		if q.Model != nil && *q.Model == "m1" {
			modelQuota = q
		}
	}
	if userQuota == nil || userQuota.Tokens != 150 || !userQuota.Over {
		t.Fatalf("user1 quota expected 150 tokens over-limit, got %+v", userQuota)
	}
	if modelQuota == nil || modelQuota.Tokens != 165 || modelQuota.Over {
		t.Fatalf("model m1 quota expected 165 tokens under-limit, got %+v", modelQuota)
	}
}

func TestUsageRowCost(t *testing.T) {
	row := usageRow{
		InputTokens:      1_000_000,
		OutputTokens:     2_000_000,
		InputPerMillion:  f64(3),
		OutputPerMillion: f64(5),
	}
	if got := row.cost(); got != 13 {
		t.Fatalf("cost = %v, want 13", got)
	}
	noPrice := usageRow{InputTokens: 100}
	if got := noPrice.cost(); got != 0 {
		t.Fatalf("cost without price = %v, want 0", got)
	}
}
