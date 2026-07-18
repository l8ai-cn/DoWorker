package sessionapi

import (
	"context"
	"errors"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

func (d *Deps) ensureMessagePod(
	ctx context.Context,
	row *sessionDomain.Session,
	pod *podDomain.Pod,
) (*podDomain.Pod, error) {
	if pod.IsActive() {
		return pod, nil
	}
	if !podDomain.IsPodStatusFinished(pod.Status) ||
		d.PodOrchestrator == nil ||
		d.Sessions == nil {
		return nil, errSwitchUnavailable
	}
	result, err := d.PodOrchestrator.CreatePod(ctx, &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:    row.OrganizationID,
		UserID:            row.UserID,
		SourcePodKey:      pod.PodKey,
		AgentSessionID:    row.ID,
		SessionMcpServers: sessionDomain.McpServersToInstalled(sessionDomain.ParseMcpServers(row.McpServers)),
	})
	if errors.Is(err, agentpod.ErrSourcePodAlreadyResumed) {
		return d.useActiveResumedPod(ctx, row, pod.PodKey)
	}
	if err != nil {
		return nil, err
	}
	if result == nil || result.Pod == nil {
		return nil, errSwitchUnavailable
	}
	return d.bindSessionToPod(ctx, row, result.Pod)
}

func (d *Deps) useActiveResumedPod(
	ctx context.Context,
	row *sessionDomain.Session,
	sourcePodKey string,
) (*podDomain.Pod, error) {
	if d.Pod == nil {
		return nil, errSwitchUnavailable
	}
	pod, err := d.Pod.GetActivePodBySourcePodKey(ctx, sourcePodKey)
	if err != nil || pod == nil {
		return nil, errSwitchUnavailable
	}
	return d.bindSessionToPod(ctx, row, pod)
}

func (d *Deps) bindSessionToPod(
	ctx context.Context,
	row *sessionDomain.Session,
	pod *podDomain.Pod,
) (*podDomain.Pod, error) {
	if err := d.Sessions.UpdatePodKey(ctx, row.ID, pod.PodKey); err != nil {
		return nil, err
	}
	row.PodKey = pod.PodKey
	if d.Updates != nil {
		d.Updates.NotifyChanged(row.ID)
	}
	if d.Stream != nil {
		d.Stream.PublishPodStatus(ctx, pod.PodKey, pod.Status, pod.AgentStatus)
	}
	return pod, nil
}
