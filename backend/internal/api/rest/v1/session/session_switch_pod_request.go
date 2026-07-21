package sessionapi

import (
	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	workercreation "github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
)

func buildSessionSnapshotRebuildPodRequest(
	row *domain.Session,
	runnerID int64,
	snapshotID *int64,
) *agentpod.OrchestrateCreatePodRequest {
	return &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:       row.OrganizationID,
		UserID:               row.UserID,
		RunnerID:             runnerID,
		AgentSessionID:       row.ID,
		WorkerSpecSnapshotID: snapshotID,
		SessionMcpServers:    domain.McpServersToInstalled(domain.ParseMcpServers(row.McpServers)),
		DeferRunnerDispatch:  true,
	}
}

func buildSessionPlanRebuildPodRequest(
	row *domain.Session,
	runnerID int64,
	draft *workercreation.Draft,
) *agentpod.OrchestrateCreatePodRequest {
	return &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:      row.OrganizationID,
		UserID:              row.UserID,
		RunnerID:            runnerID,
		AgentSessionID:      row.ID,
		WorkerSpecDraft:     draft,
		SessionMcpServers:   domain.McpServersToInstalled(domain.ParseMcpServers(row.McpServers)),
		DeferRunnerDispatch: true,
	}
}
