package workbench

import (
	"fmt"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
)

func (m *Mapper) ContentChunk(
	sessionID string,
	chunk acp.ContentChunk,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setExternalSessionLocked(sessionID)
	role, ok := messageRole(chunk.Role)
	if !ok {
		return m.batchLocked(
			chunk,
			m.unsupportedMutationLocked("message.role", chunk.Role),
		)
	}
	state := m.messages[chunk.Role]
	operation := agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE
	if state == nil {
		state = &timelineState{itemID: m.nextItemIDLocked("message-" + chunk.Role)}
		m.messages[chunk.Role] = state
		operation = agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_APPEND
	}
	state.text += chunk.Text
	content := &agentworkbenchv2.TimelineItemContent{
		Content: &agentworkbenchv2.TimelineItemContent_Message{
			Message: &agentworkbenchv2.MessageTimelineItem{
				Role:    role,
				Status:  agentworkbenchv2.TimelineItemStatus_TIMELINE_ITEM_STATUS_STREAMING,
				Content: []*agentworkbenchv2.ContentBlock{markdownBlock(state.itemID, state.text)},
			},
		},
	}
	return m.batchLocked(chunk, timelineMutation(operation, state.itemID, content))
}

func (m *Mapper) Thinking(
	sessionID string,
	update acp.ThinkingUpdate,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setExternalSessionLocked(sessionID)
	operation := agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE
	if m.thinking == nil {
		m.thinking = &timelineState{itemID: m.nextItemIDLocked("reasoning")}
		operation = agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_APPEND
	}
	m.thinking.text += update.Text
	content := &agentworkbenchv2.TimelineItemContent{
		Content: &agentworkbenchv2.TimelineItemContent_Reasoning{
			Reasoning: &agentworkbenchv2.ReasoningTimelineItem{
				Status: agentworkbenchv2.TimelineItemStatus_TIMELINE_ITEM_STATUS_RUNNING,
				Content: []*agentworkbenchv2.ContentBlock{
					textBlock(m.thinking.itemID, "reasoning", m.thinking.text),
				},
			},
		},
	}
	return m.batchLocked(update, timelineMutation(operation, m.thinking.itemID, content))
}

func (m *Mapper) Plan(
	sessionID string,
	update acp.PlanUpdate,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setExternalSessionLocked(sessionID)
	operation := agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE
	if !m.planSeen {
		m.planSeen = true
		operation = agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_APPEND
	}
	steps := make([]*agentworkbenchv2.PlanStep, 0, len(update.Steps))
	for index, step := range update.Steps {
		status, ok := planStatus(step.Status)
		if !ok {
			return m.batchLocked(
				update,
				m.unsupportedMutationLocked("plan.step.status", step.Status),
			)
		}
		steps = append(steps, &agentworkbenchv2.PlanStep{
			StepId: fmt.Sprintf("plan-step-%d", index+1),
			Title:  step.Title,
			Status: status,
		})
	}
	content := &agentworkbenchv2.TimelineItemContent{
		Content: &agentworkbenchv2.TimelineItemContent_Plan{
			Plan: &agentworkbenchv2.PlanTimelineItem{Steps: steps},
		},
	}
	return m.batchLocked(update, timelineMutation(operation, "plan", content))
}

func timelineMutation(
	operation agentworkbenchv2.RunnerTimelineOperation,
	itemID string,
	content *agentworkbenchv2.TimelineItemContent,
) *agentworkbenchv2.RunnerWorkbenchMutation {
	return &agentworkbenchv2.RunnerWorkbenchMutation{
		Mutation: &agentworkbenchv2.RunnerWorkbenchMutation_Timeline{
			Timeline: &agentworkbenchv2.RunnerTimelineMutation{
				Operation: operation,
				ItemId:    itemID,
				Content:   content,
			},
		},
	}
}

func markdownBlock(id, markdown string) *agentworkbenchv2.ContentBlock {
	return &agentworkbenchv2.ContentBlock{
		ContentId: id + ":markdown",
		Identity:  contentIdentity("content.markdown"),
		Content: &agentworkbenchv2.ContentBlock_Markdown{
			Markdown: &agentworkbenchv2.MarkdownContent{Markdown: markdown},
		},
	}
}

func textBlock(id, semanticKey, text string) *agentworkbenchv2.ContentBlock {
	return &agentworkbenchv2.ContentBlock{
		ContentId: id + ":" + semanticKey,
		Identity:  contentIdentity("content." + semanticKey),
		Content: &agentworkbenchv2.ContentBlock_Text{
			Text: &agentworkbenchv2.TextContent{Text: text},
		},
	}
}

func contentIdentity(semanticKey string) *agentworkbenchv2.ContentIdentity {
	return &agentworkbenchv2.ContentIdentity{
		Namespace:     "agentcloud.runner",
		SemanticKey:   semanticKey,
		SchemaVersion: "1",
	}
}
