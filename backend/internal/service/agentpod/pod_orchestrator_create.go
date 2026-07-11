package agentpod

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	otelinit "github.com/anthropics/agentsmesh/backend/internal/infra/otel"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent/automation"
)

func (o *PodOrchestrator) CreatePod(ctx context.Context, req *OrchestrateCreatePodRequest) (*OrchestrateCreatePodResult, error) {
	createStart := time.Now()
	defer func() {
		otelinit.PodCreateDuration.Record(ctx, float64(time.Since(createStart).Milliseconds()))
	}()

	var sourcePod *podDomain.Pod
	var sessionID string
	isResumeMode := req.SourcePodKey != ""

	if isResumeMode {
		if req.WorkerSpecDraft != nil {
			return nil, ErrConflictingWorkerCreateInput
		}
		var err error
		sourcePod, sessionID, err = o.handleResumeMode(ctx, req)
		if err != nil {
			return nil, err
		}
	} else {
		if err := o.prepareStructuredWorkerCreate(ctx, req); err != nil {
			return nil, err
		}
		if req.AgentSlug == "" {
			return nil, ErrMissingAgentSlug
		}
	}
	if !isResumeMode && req.RunnerID != 0 {
		if err := o.resolveRunnerForFreshCreate(ctx, req); err != nil {
			return nil, err
		}
	}
	var agentDef *agentDomain.Agent
	if req.AgentSlug != "" && o.agentResolver != nil {
		var err error
		agentDef, err = o.agentResolver.GetAgent(ctx, req.AgentSlug)
		if err != nil {
			return nil, ErrMissingAgentSlug
		}
	}
	if !isResumeMode {
		if err := o.validatePreparedWorkerType(ctx, req); err != nil {
			return nil, err
		}
		if err := o.preResolveFreshRepository(ctx, req, agentDef); err != nil {
			return nil, err
		}
		if req.RunnerID == 0 {
			if err := o.resolveRunnerForFreshCreate(ctx, req); err != nil {
				return nil, err
			}
		}
		sessionID = uuid.New().String()
	}

	if err := o.applyWorkerModel(ctx, req, agentDef); err != nil {
		return nil, err
	}

	// Automation level → per-agent native permission/MODE. Injected as the
	// highest-priority AgentFile layer lines (appended last ⇒ override the
	// user layer and base defaults) so the existing resolve/eval pipeline
	// produces the right launch args without touching the runner. Resume is
	// exempt: it must replay the source pod's originally-resolved config
	// verbatim, so we only inherit the source's level for the DB column.
	if isResumeMode {
		if req.AutomationLevel == "" && sourcePod != nil {
			req.AutomationLevel = sourcePod.AutomationLevel
		}
		req.AutomationLevel = podDomain.NormalizeAutomationLevel(req.AutomationLevel)
	} else {
		req.AutomationLevel = podDomain.NormalizeAutomationLevel(req.AutomationLevel)
		// An explicitly requested MODE (pty/acp) is authoritative: the automation
		// adapter must not force a different mode over it, otherwise CLI/PTY
		// workers are unreachable under the default autonomous level. The
		// adapter's CONFIG overrides still apply, so an autonomous PTY worker
		// stays non-interactive via the agent's native CLI flags.
		canForceMode := agentDef != nil &&
			agentDef.SupportsMode(podDomain.InteractionModeACP) &&
			!agentfileLayerHasModeDecl(req.AgentfileLayer)
		if lines := automation.LayerLinesFor(req.AgentSlug, req.AutomationLevel, canForceMode); lines != "" {
			appendAgentfileLayer(&req.AgentfileLayer, lines)
		}
	}

	resolved := &agentfileResolved{}

	resumeAgentSession := req.ResumeAgentSession == nil || *req.ResumeAgentSession
	externalID := req.ResumeExternalSessionID
	if externalID == "" && isResumeMode && sourcePod != nil && sourcePod.ExternalSessionID != nil {
		externalID = *sourcePod.ExternalSessionID
	}
	systemOverrides := newSystemOverrides(sessionID, isResumeMode, resumeAgentSession, externalID)

	if agentDef != nil && agentDef.AgentfileSource != nil {
		var userPrefs map[string]interface{}
		if o.userConfigQuery != nil {
			userPrefs = o.userConfigQuery.GetUserConfigPrefs(ctx, req.UserID, req.AgentSlug)
		}
		if isResumeMode {
			userPrefs = mergeSourcePodConfigPrefs(userPrefs, sourcePod, agentDef)
		}

		layerSrc := ""
		if req.AgentfileLayer != nil {
			layerSrc = *req.AgentfileLayer
		}

		result, err := extractFromAgentfileLayer(
			*agentDef.AgentfileSource, layerSrc,
			userPrefs, systemOverrides,
		)
		if err != nil {
			return nil, err
		}
		resolved.MergedAgentfileSource = result.MergedAgentfileSource
		resolved.ConfigValues = result.ConfigValues
		if result.Mode != "" {
			resolved.InteractionMode = result.Mode
		}
		if result.Branch != "" {
			resolved.BranchName = result.Branch
		}
		if result.RepoSlug != "" {
			if err := o.resolveAgentfileRepository(ctx, req, resolved, result.RepoSlug); err != nil {
				return nil, err
			}
		}
		if result.Prompt != "" {
			resolved.Prompt = result.Prompt
		}
		resolved.Knowledge = result.Knowledge
	}

	// Effective Model / PermissionMode come from resolved.ConfigValues — the
	// single source of truth post-resolve. mergeSourcePodConfigPrefs already
	// projects legacy Claude columns into userPrefs so they re-emerge here.
	effectiveInteractionMode := firstNonEmpty(resolved.InteractionMode, podDomain.InteractionModePTY)
	effectiveModel := resolved.ConfigValues.GetString(agentDomain.ConfigKeyModel)
	effectivePermissionMode := resolved.ConfigValues.GetString(agentDomain.ConfigKeyPermissionMode)
	effectiveBranch := firstNonEmptyPtr(resolved.BranchName, req.BranchName) // req.BranchName only from resume
	effectiveRepoID := firstNonNilInt64(resolved.RepositoryID, req.RepositoryID)

	if agentDef != nil && !agentDef.SupportsMode(effectiveInteractionMode) {
		return nil, ErrUnsupportedInteractionMode
	}
	if err := o.resolveEffectiveRepository(ctx, req, resolved, effectiveRepoID); err != nil {
		return nil, err
	}

	if o.billingService != nil {
		if err := o.billingService.CheckQuota(ctx, req.OrganizationID, "concurrent_pods", 1); err != nil {
			slog.WarnContext(ctx, "pod quota check failed", "org_id", req.OrganizationID, "error", err)
			return nil, err
		}
	}

	o.resolveTicketID(ctx, req)

	pod, err := o.podService.CreatePod(ctx, newPodServiceCreateRequest(
		req,
		resolved,
		effectiveRepoID,
		effectiveBranch,
		sessionID,
		effectiveInteractionMode,
		effectiveModel,
		effectivePermissionMode,
		o.initialPodStatus(req),
	))
	if err != nil {
		return nil, err
	}

	podCmd, err := o.buildPodCommand(ctx, req, pod, sourcePod, isResumeMode, resolved)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build pod command", "pod_key", pod.PodKey, "error", err)
		return nil, errors.Join(ErrConfigBuildFailed, err)
	}

	return o.dispatchCreatedPod(ctx, req, pod, podCmd, sessionID, isResumeMode)
}
