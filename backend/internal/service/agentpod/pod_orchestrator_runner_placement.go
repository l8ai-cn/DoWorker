package agentpod

import (
	"context"
	"log/slog"
)

func (o *PodOrchestrator) resolveRunnerForFreshCreate(ctx context.Context, req *OrchestrateCreatePodRequest) error {
	if req.RunnerID != 0 {
		return o.resolveExplicitRunner(ctx, req)
	}
	if o.runnerSelector == nil || o.agentResolver == nil {
		return ErrMissingRunnerID
	}

	hints := o.buildAffinityHints(req)
	repoHistory := o.fetchRepoHistory(ctx, req.OrganizationID, hints)
	selectedRunner, err := o.runnerSelector.SelectRunnerWithAffinity(
		ctx, req.OrganizationID, req.UserID, req.AgentSlug, hints, repoHistory,
	)
	if err != nil {
		slog.WarnContext(ctx, "runner auto-selection failed", "org_id", req.OrganizationID, "agent_slug", req.AgentSlug, "error", err)
		return ErrNoAvailableRunner
	}
	req.RunnerID = selectedRunner.ID
	req.clusterID = selectedRunner.ClusterID
	slog.InfoContext(ctx, "runner auto-selected", "runner_id", selectedRunner.ID, "org_id", req.OrganizationID, "agent_slug", req.AgentSlug)
	return nil
}

func (o *PodOrchestrator) resolveExplicitRunner(ctx context.Context, req *OrchestrateCreatePodRequest) error {
	if o.runnerSelector == nil {
		return ErrNoAvailableRunner
	}
	selectedRunner, err := o.runnerSelector.ResolveRunnerForCreate(
		ctx, req.RunnerID, req.OrganizationID, req.UserID, req.AgentSlug, req.QueueIfUnavailable,
	)
	if err != nil {
		slog.WarnContext(ctx, "explicit runner eligibility failed", "runner_id", req.RunnerID, "org_id", req.OrganizationID, "agent_slug", req.AgentSlug, "error", err)
		return ErrNoAvailableRunner
	}
	req.clusterID = selectedRunner.ClusterID
	return nil
}
