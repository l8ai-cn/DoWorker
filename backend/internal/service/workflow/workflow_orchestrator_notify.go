package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	eventsv1 "github.com/anthropics/agentsmesh/proto/gen/go/events/v1"
)

func (o *WorkflowOrchestrator) publishRunEvent(orgID int64, eventType eventbus.EventType, run *workflowDomain.WorkflowRun) {
	if o.eventBus == nil {
		return
	}

	data, err := eventbus.MarshalEventData(&eventsv1.WorkflowRunEventData{
		WorkflowId: run.WorkflowID,
		RunId:      run.ID,
		RunNumber:  int32(run.RunNumber),
		Status:     run.Status,
		PodKey:     run.PodKey,
	})
	if err != nil {
		o.logger.Warn("failed to marshal workflow run event", "error", err)
		return
	}

	_ = o.eventBus.Publish(context.Background(), &eventbus.Event{
		Type:           eventType,
		Category:       eventbus.CategoryEntity,
		OrganizationID: orgID,
		EntityType:     "workflow_run",
		EntityID:       fmt.Sprintf("%d", run.ID),
		Data:           data,
		Timestamp:      time.Now().UnixMilli(),
	})
}

func (o *WorkflowOrchestrator) sendWebhookCallback(callbackURL string, workflow *workflowDomain.Workflow, run *workflowDomain.WorkflowRun, status string) {
	payload, _ := json.Marshal(map[string]interface{}{
		"workflow_id":   workflow.ID,
		"workflow_slug": workflow.Slug,
		"loop_name":     workflow.Name,
		"run_id":        run.ID,
		"run_number":    run.RunNumber,
		"status":        status,
		"trigger":       run.TriggerType,
		"exit_summary":  run.ExitSummary,
		"started_at": func() string {
			if run.StartedAt != nil {
				return run.StartedAt.Format(time.RFC3339)
			}
			return ""
		}(),
		"finished_at": func() string {
			if run.FinishedAt != nil {
				return run.FinishedAt.Format(time.RFC3339)
			}
			return time.Now().Format(time.RFC3339)
		}(),
	})

	resp, err := o.httpClient.Post(callbackURL, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		o.logger.Warn("webhook callback failed",
			"workflow_id", workflow.ID, "run_id", run.ID, "url", callbackURL, "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		o.logger.Warn("webhook callback returned error",
			"workflow_id", workflow.ID, "run_id", run.ID, "url", callbackURL, "status", resp.StatusCode)
	}
}

func (o *WorkflowOrchestrator) postTicketComment(ctx context.Context, ticketID int64, userID int64, workflow *workflowDomain.Workflow, run *workflowDomain.WorkflowRun, status string) {
	statusEmoji := "✅"
	switch status {
	case workflowDomain.RunStatusFailed:
		statusEmoji = "❌"
	case workflowDomain.RunStatusTimeout:
		statusEmoji = "⏰"
	case workflowDomain.RunStatusCancelled:
		statusEmoji = "⊘"
	}

	durationStr := "-"
	if run.StartedAt != nil && run.FinishedAt != nil {
		durationStr = fmt.Sprintf("%.0fs", run.FinishedAt.Sub(*run.StartedAt).Seconds())
	}

	content := fmt.Sprintf(
		"%s **Workflow Run #%d** — %s\n\nLoop: **%s** (`%s`)\nDuration: %s\nTrigger: %s",
		statusEmoji, run.RunNumber, status, workflow.Name, workflow.Slug, durationStr, run.TriggerType,
	)

	if _, err := o.ticketService.CreateComment(ctx, ticketID, userID, content, nil, nil); err != nil {
		o.logger.Warn("failed to create ticket comment for workflow run",
			"workflow_id", workflow.ID, "run_id", run.ID, "ticket_id", ticketID, "error", err)
	}
}

func resolvePrompt(template string, defaults json.RawMessage, overrides json.RawMessage) string {
	vars := make(map[string]interface{})
	if len(defaults) > 0 {
		_ = json.Unmarshal(defaults, &vars)
	}
	if len(overrides) > 0 {
		var ov map[string]interface{}
		if err := json.Unmarshal(overrides, &ov); err == nil {
			for k, v := range ov {
				vars[k] = v
			}
		}
	}

	result := template
	for k, v := range vars {
		placeholder := "{{" + k + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", v))
	}
	return result
}
