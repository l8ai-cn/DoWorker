package workflowconnect

import (
	"time"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/pkg/protoconv"
	workflowv1 "github.com/anthropics/agentsmesh/proto/gen/go/workflow/v1"
)

func optStrPtr(p *string) *string {
	if p == nil || *p == "" {
		return nil
	}
	v := *p
	return &v
}

func optTimePtr(t *time.Time) *string {
	return protoconv.RFC3339Ptr(t)
}

func rawJSONString(b []byte) string {
	if len(b) == 0 {
		return "{}"
	}
	return string(b)
}

func toProtoWorkflow(l *workflowDomain.Workflow) *workflowv1.Workflow {
	if l == nil {
		return nil
	}
	out := &workflowv1.Workflow{
		Id:                  l.ID,
		Slug:                l.Slug,
		Name:                l.Name,
		Description:         optStrPtr(l.Description),
		AgentSlug:           l.AgentSlug,
		PermissionMode:      l.PermissionMode,
		PromptTemplate:      l.PromptTemplate,
		ConfigOverridesJson: rawJSONString(l.ConfigOverrides),
		PromptVariablesJson: rawJSONString(l.PromptVariables),
		ExecutionMode:       l.ExecutionMode,
		CronExpression:      optStrPtr(l.CronExpression),
		AutopilotConfigJson: rawJSONString(l.AutopilotConfig),
		CallbackUrl:         optStrPtr(l.CallbackURL),
		RepositoryId:        l.RepositoryID,
		RunnerId:            l.RunnerID,
		BranchName:          optStrPtr(l.BranchName),
		TicketId:            l.TicketID,
		ModelResourceId:     l.ModelResourceID,
		Status:              l.Status,
		SandboxStrategy:     l.SandboxStrategy,
		SessionPersistence:  l.SessionPersistence,
		ConcurrencyPolicy:   l.ConcurrencyPolicy,
		MaxConcurrentRuns:   int32(l.MaxConcurrentRuns),
		MaxRetainedRuns:     int32(l.MaxRetainedRuns),
		TimeoutMinutes:      int32(l.TimeoutMinutes),
		IdleTimeoutSec:      int32(l.IdleTimeoutSec),
		TotalRuns:           int64(l.TotalRuns),
		SuccessfulRuns:      int64(l.SuccessfulRuns),
		FailedRuns:          int64(l.FailedRuns),
		ActiveRunCount:      int64(l.ActiveRunCount),
		AvgDurationSec:      l.AvgDurationSec,
		LastRunAt:           optTimePtr(l.LastRunAt),
		CreatedAt:           protoconv.RFC3339(l.CreatedAt),
		UpdatedAt:           protoconv.RFC3339(l.UpdatedAt),
		UsedEnvBundles:      []string(l.UsedEnvBundles),
	}
	return out
}

func toProtoWorkflowRun(r *workflowDomain.WorkflowRun) *workflowv1.WorkflowRun {
	if r == nil {
		return nil
	}
	out := &workflowv1.WorkflowRun{
		Id:           r.ID,
		WorkflowId:   r.WorkflowID,
		RunNumber:    int64(r.RunNumber),
		Status:       r.Status,
		PodKey:       optStrPtr(r.PodKey),
		StartedAt:    optTimePtr(r.StartedAt),
		CompletedAt:  optTimePtr(r.FinishedAt),
		ErrorMessage: optStrPtr(r.ErrorMessage),
		CreatedAt:    protoconv.RFC3339(r.CreatedAt),
	}
	return out
}
