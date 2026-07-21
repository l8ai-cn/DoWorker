package sessionapi

import (
	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	workercreation "github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
)

func buildForkSnapshotPodRequest(
	source *domain.Session,
	runnerID int64,
	snapshotID *int64,
) *agentpod.OrchestrateCreatePodRequest {
	return &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:       source.OrganizationID,
		UserID:               source.UserID,
		RunnerID:             runnerID,
		WorkerSpecSnapshotID: snapshotID,
		DeferRunnerDispatch:  true,
	}
}

func buildForkPlanPodRequest(
	source *domain.Session,
	runnerID int64,
	draft *workercreation.Draft,
) *agentpod.OrchestrateCreatePodRequest {
	return &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:      source.OrganizationID,
		UserID:              source.UserID,
		RunnerID:            runnerID,
		WorkerSpecDraft:     draft,
		DeferRunnerDispatch: true,
	}
}
