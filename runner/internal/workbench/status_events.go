package workbench

import (
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

func (m *Mapper) State(
	state string,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	m.mu.Lock()
	defer m.mu.Unlock()
	status, ok := sessionStatus(state)
	if !ok {
		return m.batchLocked(
			map[string]string{"state": state},
			m.unsupportedMutationLocked("session.state", state),
		)
	}
	mutations := m.completeActiveTimelineLocked()
	mutations = append(mutations, &agentworkbenchv2.RunnerWorkbenchMutation{
		Mutation: &agentworkbenchv2.RunnerWorkbenchMutation_Status{
			Status: &agentworkbenchv2.RunnerStatusMutation{Status: status},
		},
	})
	return m.batchLocked(map[string]string{"state": state}, mutations...)
}

func (m *Mapper) Permission(
	request acp.PermissionRequest,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setExternalSessionLocked(request.SessionID)
	permission := &agentworkbenchv2.PermissionRequest{
		PermissionRequestId: request.RequestID,
		State:               agentworkbenchv2.PermissionRequestState_PERMISSION_REQUEST_STATE_PENDING,
		Request: &agentworkbenchv2.PermissionRequest_Approval{
			Approval: &agentworkbenchv2.PermissionApproval{
				Title:       request.ToolName,
				Description: stringPointer(request.Description),
			},
		},
	}
	return m.batchLocked(request, &agentworkbenchv2.RunnerWorkbenchMutation{
		Mutation: &agentworkbenchv2.RunnerWorkbenchMutation_PermissionRequest{
			PermissionRequest: permission,
		},
	})
}

func (m *Mapper) Log(
	level, message string,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	m.mu.Lock()
	defer m.mu.Unlock()
	itemID := m.nextItemIDLocked("log")
	block := &agentworkbenchv2.ContentBlock{
		ContentId: itemID + ":log",
		Identity:  contentIdentity("content.log"),
		Content: &agentworkbenchv2.ContentBlock_Log{
			Log: &agentworkbenchv2.LogContent{Level: level, Message: message},
		},
	}
	content := &agentworkbenchv2.TimelineItemContent{
		Content: &agentworkbenchv2.TimelineItemContent_System{
			System: &agentworkbenchv2.SystemTimelineItem{
				Content: []*agentworkbenchv2.ContentBlock{block},
			},
		},
	}
	source := map[string]string{"level": level, "message": message}
	return m.batchLocked(
		source,
		timelineMutation(
			agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_APPEND,
			itemID,
			content,
		),
	)
}

func (m *Mapper) completeActiveTimelineLocked() []*agentworkbenchv2.RunnerWorkbenchMutation {
	mutations := make([]*agentworkbenchv2.RunnerWorkbenchMutation, 0, len(m.messages)+1)
	for role, state := range m.messages {
		messageRole, ok := messageRole(role)
		if !ok {
			mutations = append(
				mutations,
				m.unsupportedMutationLocked("message.role", role),
			)
			continue
		}
		content := &agentworkbenchv2.TimelineItemContent{
			Content: &agentworkbenchv2.TimelineItemContent_Message{
				Message: &agentworkbenchv2.MessageTimelineItem{
					Role:    messageRole,
					Status:  agentworkbenchv2.TimelineItemStatus_TIMELINE_ITEM_STATUS_COMPLETED,
					Content: []*agentworkbenchv2.ContentBlock{markdownBlock(state.itemID, state.text)},
				},
			},
		}
		mutations = append(mutations, timelineMutation(
			agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE,
			state.itemID,
			content,
		))
	}
	if m.thinking != nil {
		content := &agentworkbenchv2.TimelineItemContent{
			Content: &agentworkbenchv2.TimelineItemContent_Reasoning{
				Reasoning: &agentworkbenchv2.ReasoningTimelineItem{
					Status: agentworkbenchv2.TimelineItemStatus_TIMELINE_ITEM_STATUS_COMPLETED,
					Content: []*agentworkbenchv2.ContentBlock{
						textBlock(m.thinking.itemID, "reasoning", m.thinking.text),
					},
				},
			},
		}
		mutations = append(mutations, timelineMutation(
			agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE,
			m.thinking.itemID,
			content,
		))
	}
	m.messages = make(map[string]*timelineState)
	m.thinking = nil
	return mutations
}

func (m *Mapper) unsupportedMutationLocked(
	semanticKey, value string,
) *agentworkbenchv2.RunnerWorkbenchMutation {
	identity := &agentworkbenchv2.ContentIdentity{
		Namespace:     "agentsmesh." + m.sourceProtocol,
		SemanticKey:   semanticKey,
		SchemaVersion: "1",
	}
	unsupported := &agentworkbenchv2.UnsupportedValue{
		Identity: identity,
		Reason:   agentworkbenchv2.UnsupportedReason_UNSUPPORTED_REASON_UNKNOWN,
		Payload:  rawPayload("text/plain", value),
	}
	return &agentworkbenchv2.RunnerWorkbenchMutation{
		Mutation: &agentworkbenchv2.RunnerWorkbenchMutation_Unsupported{
			Unsupported: &agentworkbenchv2.RunnerUnsupportedMutation{Value: unsupported},
		},
	}
}

func (m *Mapper) setExternalSessionLocked(sessionID string) {
	if sessionID != "" {
		m.externalSession = sessionID
	}
}

func messageRole(role string) (agentworkbenchv2.MessageRole, bool) {
	switch role {
	case "user":
		return agentworkbenchv2.MessageRole_MESSAGE_ROLE_USER, true
	case "assistant":
		return agentworkbenchv2.MessageRole_MESSAGE_ROLE_ASSISTANT, true
	case "system":
		return agentworkbenchv2.MessageRole_MESSAGE_ROLE_SYSTEM, true
	default:
		return agentworkbenchv2.MessageRole_MESSAGE_ROLE_UNSPECIFIED, false
	}
}

func planStatus(status string) (agentworkbenchv2.PlanStepStatus, bool) {
	switch status {
	case "", "pending":
		return agentworkbenchv2.PlanStepStatus_PLAN_STEP_STATUS_PENDING, true
	case "in_progress", "running":
		return agentworkbenchv2.PlanStepStatus_PLAN_STEP_STATUS_RUNNING, true
	case "completed":
		return agentworkbenchv2.PlanStepStatus_PLAN_STEP_STATUS_COMPLETED, true
	case "failed":
		return agentworkbenchv2.PlanStepStatus_PLAN_STEP_STATUS_FAILED, true
	default:
		return agentworkbenchv2.PlanStepStatus_PLAN_STEP_STATUS_UNSPECIFIED, false
	}
}

func sessionStatus(state string) (agentworkbenchv2.SessionStatus, bool) {
	switch state {
	case acp.StateUninitialized, acp.StateInitializing:
		return agentworkbenchv2.SessionStatus_SESSION_STATUS_LAUNCHING, true
	case acp.StateProcessing:
		return agentworkbenchv2.SessionStatus_SESSION_STATUS_RUNNING, true
	case acp.StateWaitingPermission:
		return agentworkbenchv2.SessionStatus_SESSION_STATUS_WAITING, true
	case acp.StateIdle:
		return agentworkbenchv2.SessionStatus_SESSION_STATUS_IDLE, true
	case acp.StateStopped:
		return agentworkbenchv2.SessionStatus_SESSION_STATUS_COMPLETED, true
	default:
		return agentworkbenchv2.SessionStatus_SESSION_STATUS_UNSPECIFIED, false
	}
}
