package agentpod

import (
	"context"
	"log/slog"

	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
)

func (o *PodOrchestrator) buildAffinityHints(req *OrchestrateCreatePodRequest) *runnerDomain.AffinityHints {
	hints := &runnerDomain.AffinityHints{}
	if req.preResolvedRepository != nil {
		hints.RepositoryID = &req.preResolvedRepository.ID
	}
	return hints
}

// fetchRepoHistory queries pod history for repo affinity scoring.
// Returns nil if no repo hint or podRepo unavailable.
func (o *PodOrchestrator) fetchRepoHistory(ctx context.Context, orgID int64, hints *runnerDomain.AffinityHints) map[int64]int {
	if hints == nil || hints.RepositoryID == nil || o.podRepo == nil {
		return nil
	}
	histories, err := o.podRepo.ListRunnersByRepo(ctx, orgID, *hints.RepositoryID, 10)
	if err != nil {
		slog.Warn("repo history lookup failed, ignoring repo affinity", "error", err)
		return nil
	}
	m := make(map[int64]int, len(histories))
	for _, h := range histories {
		m[h.RunnerID] = h.PodCount
	}
	return m
}
