package workflow

import (
	"context"
	"log/slog"
	"strings"
	"time"

	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
)

func (s *WorkflowService) Create(ctx context.Context, req *CreateWorkflowRequest) (*workflowDomain.Workflow, error) {
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}
	if !isValidSlug(slug) {
		return nil, ErrInvalidSlug
	}

	if req.PermissionMode == "" {
		req.PermissionMode = "bypassPermissions"
	}
	if req.ExecutionMode == "" {
		req.ExecutionMode = workflowDomain.ExecutionModeAutopilot
	}
	if req.SandboxStrategy == "" {
		req.SandboxStrategy = workflowDomain.SandboxStrategyPersistent
	}
	if req.ConcurrencyPolicy == "" {
		req.ConcurrencyPolicy = workflowDomain.ConcurrencyPolicySkip
	}
	if req.MaxConcurrentRuns == 0 {
		req.MaxConcurrentRuns = 1
	}
	if req.TimeoutMinutes == 0 {
		req.TimeoutMinutes = 60
	}
	if req.AutopilotConfig == nil {
		req.AutopilotConfig = []byte("{}")
	}
	if req.ConfigOverrides == nil {
		req.ConfigOverrides = []byte("{}")
	}
	if req.PromptVariables == nil {
		req.PromptVariables = []byte("{}")
	}

	if err := validateEnumFields(req.ExecutionMode, req.SandboxStrategy, req.ConcurrencyPolicy); err != nil {
		return nil, err
	}

	if req.CallbackURL != nil {
		if err := validateCallbackURL(*req.CallbackURL); err != nil {
			return nil, err
		}
	}

	var nextRunAt *time.Time
	if req.CronExpression != nil && *req.CronExpression != "" {
		schedule, err := cronParser.Parse(*req.CronExpression)
		if err != nil {
			return nil, ErrInvalidCron
		}
		next := schedule.Next(time.Now())
		nextRunAt = &next
	}

	workflow := &workflowDomain.Workflow{
		OrganizationID:     req.OrganizationID,
		Name:               req.Name,
		Slug:               slug,
		Description:        req.Description,
		AgentSlug:          req.AgentSlug,
		PermissionMode:     req.PermissionMode,
		PromptTemplate:     req.PromptTemplate,
		PromptVariables:    req.PromptVariables,
		RepositoryID:       req.RepositoryID,
		RunnerID:           req.RunnerID,
		BranchName:         req.BranchName,
		TicketID:           req.TicketID,
		ModelResourceID:    req.ModelResourceID,
		UsedEnvBundles:     req.UsedEnvBundles,
		ConfigOverrides:    req.ConfigOverrides,
		ExecutionMode:      req.ExecutionMode,
		CronExpression:     req.CronExpression,
		AutopilotConfig:    req.AutopilotConfig,
		CallbackURL:        req.CallbackURL,
		Status:             workflowDomain.StatusEnabled,
		SandboxStrategy:    req.SandboxStrategy,
		SessionPersistence: req.SessionPersistence,
		ConcurrencyPolicy:  req.ConcurrencyPolicy,
		MaxConcurrentRuns:  req.MaxConcurrentRuns,
		MaxRetainedRuns:    req.MaxRetainedRuns,
		TimeoutMinutes:     req.TimeoutMinutes,
		IdleTimeoutSec:     req.IdleTimeoutSec,
		CreatedByID:        req.CreatedByID,
		NextRunAt:          nextRunAt,
	}

	if err := s.repo.Create(ctx, workflow); err != nil {
		if strings.Contains(err.Error(), "idx_workflows_org_slug") {
			return nil, ErrDuplicateSlug
		}
		slog.ErrorContext(ctx, "failed to create workflow", "slug", slug, "org_id", req.OrganizationID, "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "workflow created", "workflow_id", workflow.ID, "slug", slug, "org_id", req.OrganizationID)
	return workflow, nil
}
