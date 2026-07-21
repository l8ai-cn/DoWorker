package agentpod

import (
	"context"
	"errors"
	"fmt"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
)

func (o *PodOrchestrator) inheritResumeState(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	source *podDomain.Pod,
) error {
	if err := inheritResumeInteractionMode(req, source); err != nil {
		return err
	}
	if req.ResumeExternalSessionID != "" {
		return nil
	}
	externalID, err := o.findResumeExternalSessionID(ctx, source)
	if err != nil {
		return err
	}
	req.ResumeExternalSessionID = externalID
	return nil
}

func inheritResumeInteractionMode(
	req *OrchestrateCreatePodRequest,
	source *podDomain.Pod,
) error {
	if source == nil {
		return nil
	}
	switch source.InteractionMode {
	case podDomain.InteractionModePTY, podDomain.InteractionModeACP:
	default:
		return ErrUnsupportedInteractionMode
	}
	if agentfileLayerMode(req.AgentfileLayer) == source.InteractionMode {
		return nil
	}
	appendAgentfileLayer(
		&req.AgentfileLayer,
		fmt.Sprintf("MODE %s", source.InteractionMode),
	)
	return nil
}

func (o *PodOrchestrator) findResumeExternalSessionID(
	ctx context.Context,
	source *podDomain.Pod,
) (string, error) {
	seen := map[string]struct{}{}
	current := source
	for current != nil {
		if current.ExternalSessionID != nil && *current.ExternalSessionID != "" {
			return *current.ExternalSessionID, nil
		}
		if current.SourcePodKey == nil || *current.SourcePodKey == "" {
			return "", nil
		}
		if _, exists := seen[current.PodKey]; exists {
			return "", ErrResumeLineageInvalid
		}
		seen[current.PodKey] = struct{}{}
		parent, err := o.podService.GetPod(ctx, *current.SourcePodKey)
		if err != nil {
			return "", errors.Join(ErrResumeLineageInvalid, err)
		}
		if parent.OrganizationID != source.OrganizationID {
			return "", ErrResumeLineageInvalid
		}
		current = parent
	}
	return "", nil
}
