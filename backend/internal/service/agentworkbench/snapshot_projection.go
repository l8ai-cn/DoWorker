package agentworkbench

import (
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

func applyProjectedEvent(
	snapshot *agentworkbenchv2.SessionSnapshot,
	event *agentworkbenchv2.AgentEvent,
) error {
	envelope := event.GetEnvelope()
	switch value := event.Event.(type) {
	case *agentworkbenchv2.AgentEvent_TimelineItemAppended:
		if timelineIndex(snapshot, envelope.ItemId) >= 0 {
			return ErrInvalidBatch
		}
		snapshot.History = append(snapshot.History, timelineItem(envelope, value.TimelineItemAppended.Content))
	case *agentworkbenchv2.AgentEvent_TimelineItemUpdated:
		index := timelineIndex(snapshot, envelope.ItemId)
		if index < 0 {
			return ErrInvalidBatch
		}
		snapshot.History[index] = timelineItem(envelope, value.TimelineItemUpdated.Content)
	case *agentworkbenchv2.AgentEvent_CommandReceiptChanged:
		receipt := value.CommandReceiptChanged.GetReceipt()
		if receipt == nil || receipt.CommandId == "" {
			return ErrInvalidBatch
		}
		snapshot.CommandReceipts = upsertReceipt(snapshot.CommandReceipts, receipt)
	case *agentworkbenchv2.AgentEvent_PermissionRequested:
		request := value.PermissionRequested.GetRequest()
		snapshot.PermissionRequests = upsertPermission(snapshot.PermissionRequests, request)
	case *agentworkbenchv2.AgentEvent_PermissionResolved:
		resolution := value.PermissionResolved.GetResolution()
		request := findPermission(snapshot, resolution.GetPermissionRequestId())
		if request == nil {
			return ErrInvalidBatch
		}
		request.State = agentworkbenchv2.PermissionRequestState_PERMISSION_REQUEST_STATE_RESOLVED
		request.Resolution = proto.Clone(resolution).(*agentworkbenchv2.PermissionResolution)
	case *agentworkbenchv2.AgentEvent_ResourceChanged:
		snapshot.Resources = upsertResource(
			snapshot.Resources,
			value.ResourceChanged.GetResource(),
		)
	case *agentworkbenchv2.AgentEvent_ArtifactChanged:
		if err := validateArtifactTransition(
			snapshot.Artifacts,
			value.ArtifactChanged.GetArtifact(),
		); err != nil {
			return err
		}
		snapshot.Artifacts = upsertArtifact(
			snapshot.Artifacts,
			value.ArtifactChanged.GetArtifact(),
		)
	case *agentworkbenchv2.AgentEvent_CapabilitiesChanged:
		snapshot.Capabilities = proto.Clone(
			value.CapabilitiesChanged.GetCapabilities(),
		).(*agentworkbenchv2.SupportCapabilities)
	case *agentworkbenchv2.AgentEvent_SessionStatusChanged:
		snapshot.Status = value.SessionStatusChanged.Status
		snapshot.Error = cloneAgentError(value.SessionStatusChanged.Error)
	case *agentworkbenchv2.AgentEvent_ConfigurationChanged:
		snapshot.Configuration = proto.Clone(
			value.ConfigurationChanged.GetConfiguration(),
		).(*agentworkbenchv2.SessionConfiguration)
	case *agentworkbenchv2.AgentEvent_Unsupported:
		if timelineIndex(snapshot, envelope.ItemId) >= 0 {
			return ErrInvalidBatch
		}
		snapshot.History = append(snapshot.History, timelineItem(
			envelope,
			&agentworkbenchv2.TimelineItemContent{
				Content: &agentworkbenchv2.TimelineItemContent_Unsupported{
					Unsupported: proto.Clone(value.Unsupported).(*agentworkbenchv2.UnsupportedValue),
				},
			},
		))
	default:
		return ErrInvalidBatch
	}
	return nil
}

func timelineItem(
	envelope *agentworkbenchv2.EventEnvelope,
	content *agentworkbenchv2.TimelineItemContent,
) *agentworkbenchv2.TimelineItem {
	return &agentworkbenchv2.TimelineItem{
		Envelope: proto.Clone(envelope).(*agentworkbenchv2.EventEnvelope),
		Content:  proto.Clone(content).(*agentworkbenchv2.TimelineItemContent),
	}
}

func timelineIndex(snapshot *agentworkbenchv2.SessionSnapshot, itemID string) int {
	for index, item := range snapshot.History {
		if item.GetEnvelope().GetItemId() == itemID {
			return index
		}
	}
	return -1
}

func findPermission(
	snapshot *agentworkbenchv2.SessionSnapshot,
	requestID string,
) *agentworkbenchv2.PermissionRequest {
	for _, request := range snapshot.PermissionRequests {
		if request.PermissionRequestId == requestID {
			return request
		}
	}
	return nil
}

func upsertPermission(
	items []*agentworkbenchv2.PermissionRequest,
	value *agentworkbenchv2.PermissionRequest,
) []*agentworkbenchv2.PermissionRequest {
	next := proto.Clone(value).(*agentworkbenchv2.PermissionRequest)
	for index, item := range items {
		if item.PermissionRequestId == value.PermissionRequestId {
			items[index] = next
			return items
		}
	}
	return append(items, next)
}

func upsertResource(
	items []*agentworkbenchv2.SessionResource,
	value *agentworkbenchv2.SessionResource,
) []*agentworkbenchv2.SessionResource {
	next := proto.Clone(value).(*agentworkbenchv2.SessionResource)
	for index, item := range items {
		if item.ResourceId == value.ResourceId {
			items[index] = next
			return items
		}
	}
	return append(items, next)
}

func upsertArtifact(
	items []*agentworkbenchv2.ArtifactDescriptor,
	value *agentworkbenchv2.ArtifactDescriptor,
) []*agentworkbenchv2.ArtifactDescriptor {
	next := proto.Clone(value).(*agentworkbenchv2.ArtifactDescriptor)
	for index, item := range items {
		if item.ArtifactId == value.ArtifactId {
			next.Revisions = mergeArtifactRevisions(item.Revisions, next.Revisions)
			items[index] = next
			return items
		}
	}
	return append(items, next)
}

func upsertReceipt(
	items []*agentworkbenchv2.CommandReceipt,
	value *agentworkbenchv2.CommandReceipt,
) []*agentworkbenchv2.CommandReceipt {
	next := proto.Clone(value).(*agentworkbenchv2.CommandReceipt)
	for index, item := range items {
		if item.CommandId == value.CommandId {
			items[index] = next
			return items
		}
	}
	return append(items, next)
}
