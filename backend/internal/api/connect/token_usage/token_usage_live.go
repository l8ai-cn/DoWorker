package tokenusageconnect

import (
	sessionusagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionusage"
	"github.com/anthropics/agentsmesh/backend/internal/domain/tokenusage"
)

func mergeLivePodSessionUsage(live *sessionusagesvc.Aggregate, summary **tokenusage.UsageSummary, byModel *[]tokenusage.ModelUsage) {
	if live == nil || len(live.UsageByModel) == 0 {
		return
	}
	if *summary == nil {
		*summary = &tokenusage.UsageSummary{}
	}
	s := *summary
	for _, mu := range live.UsageByModel {
		s.InputTokens += mu.InputTokens
		s.OutputTokens += mu.OutputTokens
		s.CacheReadTokens += mu.CacheReadTokens
		s.CacheCreationTokens += mu.CacheCreationTokens
	}
	s.TotalTokens = s.InputTokens + s.OutputTokens + s.CacheReadTokens + s.CacheCreationTokens
	*summary = s

	byIdx := make(map[string]int, len(*byModel))
	for i, row := range *byModel {
		byIdx[row.Model] = i
	}
	for model, mu := range live.UsageByModel {
		total := mu.InputTokens + mu.OutputTokens + mu.CacheReadTokens + mu.CacheCreationTokens
		if i, ok := byIdx[model]; ok {
			row := &(*byModel)[i]
			row.InputTokens += mu.InputTokens
			row.OutputTokens += mu.OutputTokens
			row.CacheReadTokens += mu.CacheReadTokens
			row.CacheCreationTokens += mu.CacheCreationTokens
			row.TotalTokens += total
			continue
		}
		*byModel = append(*byModel, tokenusage.ModelUsage{
			Model:               model,
			InputTokens:         mu.InputTokens,
			OutputTokens:        mu.OutputTokens,
			CacheReadTokens:     mu.CacheReadTokens,
			CacheCreationTokens: mu.CacheCreationTokens,
			TotalTokens:         total,
		})
	}
}
