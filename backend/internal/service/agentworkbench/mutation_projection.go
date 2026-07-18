package agentworkbench

import (
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

func projectMutation(
	sessionID string,
	streamEpoch string,
	revision uint64,
	sequence uint64,
	mutation *agentworkbenchv2.RunnerWorkbenchMutation,
) (*agentworkbenchv2.AgentEvent, error) {
	if mutation == nil || mutation.Source == nil {
		return nil, ErrInvalidBatch
	}
	envelope := &agentworkbenchv2.EventEnvelope{
		SessionId:          sessionID,
		StreamEpoch:        streamEpoch,
		Revision:           revision,
		Sequence:           sequence,
		CausationCommandId: mutation.Source.CausationCommandId,
		CreatedAt:          mutation.Source.OccurredAt,
	}
	event := &agentworkbenchv2.AgentEvent{Envelope: envelope}
	switch value := mutation.Mutation.(type) {
	case *agentworkbenchv2.RunnerWorkbenchMutation_Timeline:
		if err := projectTimelineMutation(envelope, event, value.Timeline); err != nil {
			return nil, err
		}
	case *agentworkbenchv2.RunnerWorkbenchMutation_Artifact:
		if value.Artifact == nil || value.Artifact.ArtifactId == "" {
			return nil, ErrInvalidBatch
		}
		envelope.ItemId = "artifact:" + value.Artifact.ArtifactId
		event.Event = &agentworkbenchv2.AgentEvent_ArtifactChanged{
			ArtifactChanged: &agentworkbenchv2.ArtifactChanged{
				Artifact: proto.Clone(value.Artifact).(*agentworkbenchv2.ArtifactDescriptor),
			},
		}
	case *agentworkbenchv2.RunnerWorkbenchMutation_PermissionRequest:
		if value.PermissionRequest == nil || value.PermissionRequest.PermissionRequestId == "" {
			return nil, ErrInvalidBatch
		}
		envelope.ItemId = "permission:" + value.PermissionRequest.PermissionRequestId
		event.Event = &agentworkbenchv2.AgentEvent_PermissionRequested{
			PermissionRequested: &agentworkbenchv2.PermissionRequested{
				Request: proto.Clone(value.PermissionRequest).(*agentworkbenchv2.PermissionRequest),
			},
		}
	case *agentworkbenchv2.RunnerWorkbenchMutation_PermissionResolution:
		resolution := value.PermissionResolution.GetResolution()
		if resolution == nil || resolution.PermissionRequestId == "" {
			return nil, ErrInvalidBatch
		}
		envelope.ItemId = "permission:" + resolution.PermissionRequestId
		event.Event = &agentworkbenchv2.AgentEvent_PermissionResolved{
			PermissionResolved: &agentworkbenchv2.PermissionResolved{
				Resolution: proto.Clone(resolution).(*agentworkbenchv2.PermissionResolution),
			},
		}
	case *agentworkbenchv2.RunnerWorkbenchMutation_Resource:
		if value.Resource == nil || value.Resource.ResourceId == "" {
			return nil, ErrInvalidBatch
		}
		envelope.ItemId = "resource:" + value.Resource.ResourceId
		event.Event = &agentworkbenchv2.AgentEvent_ResourceChanged{
			ResourceChanged: &agentworkbenchv2.ResourceChanged{
				Resource: proto.Clone(value.Resource).(*agentworkbenchv2.SessionResource),
			},
		}
	case *agentworkbenchv2.RunnerWorkbenchMutation_Capabilities:
		if value.Capabilities == nil {
			return nil, ErrInvalidBatch
		}
		envelope.ItemId = mutation.Source.StableEventId
		event.Event = &agentworkbenchv2.AgentEvent_CapabilitiesChanged{
			CapabilitiesChanged: &agentworkbenchv2.CapabilitiesChanged{
				Capabilities: proto.Clone(value.Capabilities).(*agentworkbenchv2.SupportCapabilities),
			},
		}
	case *agentworkbenchv2.RunnerWorkbenchMutation_Status:
		if value.Status == nil ||
			value.Status.Status == agentworkbenchv2.SessionStatus_SESSION_STATUS_UNSPECIFIED {
			return nil, ErrInvalidBatch
		}
		envelope.ItemId = mutation.Source.StableEventId
		event.Event = &agentworkbenchv2.AgentEvent_SessionStatusChanged{
			SessionStatusChanged: &agentworkbenchv2.SessionStatusChanged{
				Status: value.Status.Status,
				Error:  cloneAgentError(value.Status.Error),
			},
		}
	case *agentworkbenchv2.RunnerWorkbenchMutation_Configuration:
		if value.Configuration == nil ||
			(value.Configuration.Model == nil &&
				value.Configuration.PermissionMode == nil) {
			return nil, ErrInvalidBatch
		}
		envelope.ItemId = mutation.Source.StableEventId
		event.Event = &agentworkbenchv2.AgentEvent_ConfigurationChanged{
			ConfigurationChanged: &agentworkbenchv2.ConfigurationChanged{
				Configuration: proto.Clone(value.Configuration).(*agentworkbenchv2.SessionConfiguration),
			},
		}
	case *agentworkbenchv2.RunnerWorkbenchMutation_Unsupported:
		if value.Unsupported == nil || value.Unsupported.Value == nil {
			return nil, ErrInvalidBatch
		}
		envelope.ItemId = mutation.Source.StableEventId
		event.Event = &agentworkbenchv2.AgentEvent_Unsupported{
			Unsupported: proto.Clone(value.Unsupported.Value).(*agentworkbenchv2.UnsupportedValue),
		}
	default:
		return nil, ErrInvalidBatch
	}
	return event, nil
}

func projectTimelineMutation(
	envelope *agentworkbenchv2.EventEnvelope,
	event *agentworkbenchv2.AgentEvent,
	mutation *agentworkbenchv2.RunnerTimelineMutation,
) error {
	if mutation == nil || mutation.ItemId == "" || mutation.Content == nil {
		return ErrInvalidBatch
	}
	envelope.ItemId = mutation.ItemId
	envelope.ParentId = mutation.ParentId
	envelope.TurnId = mutation.TurnId
	content := proto.Clone(mutation.Content).(*agentworkbenchv2.TimelineItemContent)
	switch mutation.Operation {
	case agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_APPEND:
		event.Event = &agentworkbenchv2.AgentEvent_TimelineItemAppended{
			TimelineItemAppended: &agentworkbenchv2.TimelineItemAppended{Content: content},
		}
	case agentworkbenchv2.RunnerTimelineOperation_RUNNER_TIMELINE_OPERATION_UPDATE:
		event.Event = &agentworkbenchv2.AgentEvent_TimelineItemUpdated{
			TimelineItemUpdated: &agentworkbenchv2.TimelineItemUpdated{Content: content},
		}
	default:
		return ErrInvalidBatch
	}
	return nil
}

func cloneAgentError(value *agentworkbenchv2.AgentError) *agentworkbenchv2.AgentError {
	if value == nil {
		return nil
	}
	return proto.Clone(value).(*agentworkbenchv2.AgentError)
}
