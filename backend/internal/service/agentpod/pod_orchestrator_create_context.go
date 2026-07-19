package agentpod

import (
	"context"
	"strings"

	"github.com/google/uuid"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

type podCreationContext struct {
	sourcePod       *podDomain.Pod
	workerLaunchPod *podDomain.Pod
	agentDef        *agentDomain.Agent
	sessionID       string
	isResumeMode    bool
}

func (o *PodOrchestrator) preparePodCreationContext(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
) (podCreationContext, error) {
	source, err := resolveExecutionSource(req)
	if err != nil {
		return podCreationContext{}, err
	}
	state := podCreationContext{
		isResumeMode: source == ExecutionSourceLineage,
	}
	if state.isResumeMode {
		state.sourcePod, state.sessionID, err = o.handleResumeMode(ctx, req)
		if err != nil {
			return podCreationContext{}, err
		}
	} else {
		if err := o.prepareFreshWorkerCreate(ctx, req, source); err != nil {
			return podCreationContext{}, err
		}
		state.workerLaunchPod, err = o.bindExistingWorkerLaunchPod(ctx, req)
		if err != nil {
			return podCreationContext{}, err
		}
	}
	if !state.isResumeMode && state.workerLaunchPod == nil &&
		req.RunnerID != 0 {
		if err := o.resolveRunnerForFreshCreate(ctx, req); err != nil {
			return podCreationContext{}, err
		}
	}
	state.agentDef, err = o.runtimeAgentDefinition(ctx, req)
	if err != nil {
		return podCreationContext{}, err
	}
	if state.agentDef == nil ||
		strings.TrimSpace(state.agentDef.AdapterID) == "" {
		return podCreationContext{}, ErrMissingAgentAdapter
	}
	if !state.isResumeMode {
		if err := o.prepareFreshWorkerRuntime(
			ctx,
			req,
			state.agentDef,
			state.workerLaunchPod,
		); err != nil {
			return podCreationContext{}, err
		}
		if state.workerLaunchPod == nil {
			state.sessionID = uuid.NewString()
		} else if state.workerLaunchPod.SessionID == nil ||
			*state.workerLaunchPod.SessionID == "" {
			return podCreationContext{}, ErrWorkerLaunchPodMismatch
		} else {
			state.sessionID = *state.workerLaunchPod.SessionID
		}
	}
	return state, nil
}

func (o *PodOrchestrator) prepareFreshWorkerCreate(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	source ExecutionSource,
) error {
	if source == ExecutionSourceSnapshot {
		if err := o.prepareSnapshotWorkerCreate(ctx, req); err != nil {
			return err
		}
	} else if err := o.prepareStructuredWorkerCreate(ctx, req); err != nil {
		return err
	}
	if req.AgentSlug == "" {
		return ErrMissingAgentSlug
	}
	return nil
}

func (o *PodOrchestrator) prepareFreshWorkerRuntime(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	agentDef *agentDomain.Agent,
	existing *podDomain.Pod,
) error {
	if err := o.validatePreparedWorkerType(ctx, req); err != nil {
		return err
	}
	if err := o.preResolveFreshRepository(ctx, req, agentDef); err != nil {
		return err
	}
	if existing == nil && req.RunnerID == 0 {
		return o.resolveRunnerForFreshCreate(ctx, req)
	}
	return nil
}
