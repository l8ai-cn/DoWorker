package sessionapi

import (
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
)

func sessionSnapshotSource(
	row *domain.Session,
	pod *podDomain.Pod,
) (*int64, error) {
	if row == nil || pod == nil || pod.PodKey == "" {
		return nil, invalidSessionWorkerPlan("worker_spec_snapshot_id", "source pod is required")
	}
	if pod.PodKey != row.PodKey ||
		pod.OrganizationID != row.OrganizationID ||
		pod.CreatedByID != row.UserID {
		return nil, invalidSessionWorkerPlan("worker_spec_snapshot_id", "source pod does not belong to session")
	}
	if pod.WorkerSpecSnapshotID == nil || *pod.WorkerSpecSnapshotID <= 0 {
		return nil, invalidSessionWorkerPlan("worker_spec_snapshot_id", "source pod snapshot is required")
	}
	return pod.WorkerSpecSnapshotID, nil
}
