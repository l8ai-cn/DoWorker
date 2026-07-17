package agentworkbench

import (
	"context"
	"encoding/json"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
)

func (dispatcher *CommandDispatcher) deliverPermission(
	ctx context.Context,
	runnerID int64,
	podKey string,
	command *agentworkbenchv2.CommandEnvelope,
	value *agentworkbenchv2.ResolvePermissionCommand,
) error {
	if value == nil || value.PermissionRequestId == "" ||
		value.Decision == agentworkbenchv2.PermissionDecision_PERMISSION_DECISION_UNSPECIFIED {
		return ErrInvalidCommand
	}
	updatedInput := map[string]any{}
	if value.Response != nil {
		if value.Response.MediaType != "application/json" ||
			json.Unmarshal(value.Response.Data, &updatedInput) != nil {
			return ErrInvalidCommand
		}
	}
	return dispatcher.sendACP(ctx, runnerID, podKey, map[string]any{
		"type": "permission_response", "requestId": value.PermissionRequestId,
		"approved":     value.Decision == agentworkbenchv2.PermissionDecision_PERMISSION_DECISION_ACCEPT,
		"updatedInput": updatedInput, "commandId": command.CommandId,
	})
}

func (dispatcher *CommandDispatcher) deliverConfiguration(
	ctx context.Context,
	runnerID int64,
	podKey string,
	command *agentworkbenchv2.CommandEnvelope,
	value *agentworkbenchv2.ChangeConfigurationCommand,
) error {
	if value == nil || len(value.Values) != 1 || value.Values[0].Value == nil {
		return ErrInvalidCommand
	}
	var selected string
	if value.Values[0].Value.MediaType != "application/json" ||
		json.Unmarshal(value.Values[0].Value.Data, &selected) != nil ||
		selected == "" {
		return ErrInvalidCommand
	}
	commandType := ""
	payloadField := ""
	switch value.Values[0].Key {
	case "model":
		commandType = "set_model"
		payloadField = "model"
	case "permission_mode":
		commandType = "set_permission_mode"
		payloadField = "mode"
	default:
		return ErrCommandUnavailable
	}
	return dispatcher.sendACP(ctx, runnerID, podKey, map[string]any{
		"type": commandType, payloadField: selected,
		"requestId": command.CommandId,
	})
}

func (dispatcher *CommandDispatcher) deliverArtifactAction(
	ctx context.Context,
	runnerID int64,
	podKey string,
	command *agentworkbenchv2.CommandEnvelope,
	value *agentworkbenchv2.ArtifactActionCommand,
) error {
	if value == nil || value.ArtifactId == "" || value.ActionType == "" ||
		value.ActionSchemaVersion == "" || value.Payload == nil {
		return ErrInvalidCommand
	}
	payload, err := structuredJSON(value.Payload)
	if err != nil {
		return err
	}
	return dispatcher.sendACP(ctx, runnerID, podKey, map[string]any{
		"type": "control_request", "subtype": "artifact_action",
		"requestId": command.CommandId,
		"payload": map[string]any{
			"artifactId": value.ArtifactId, "representationId": value.RepresentationId,
			"baseRevision": value.BaseRevision, "clientActionId": value.ClientActionId,
			"actionType": value.ActionType, "schemaVersion": value.ActionSchemaVersion,
			"input": payload,
		},
	})
}

func (dispatcher *CommandDispatcher) deliverExtension(
	ctx context.Context,
	runnerID int64,
	podKey string,
	command *agentworkbenchv2.CommandEnvelope,
	value *agentworkbenchv2.ExtensionCommand,
) error {
	if value == nil || value.Namespace == "" || value.SemanticType == "" ||
		value.SchemaVersion == "" || value.Payload == nil {
		return ErrInvalidCommand
	}
	payload, err := structuredJSON(value.Payload)
	if err != nil {
		return err
	}
	return dispatcher.sendACP(ctx, runnerID, podKey, map[string]any{
		"type": "control_request", "subtype": value.SemanticType,
		"requestId": command.CommandId,
		"payload": map[string]any{
			"namespace": value.Namespace, "schemaVersion": value.SchemaVersion,
			"input": payload,
		},
	})
}

func structuredJSON(value *agentworkbenchv2.StructuredPayload) (any, error) {
	if value.MediaType != "application/json" {
		return nil, ErrInvalidCommand
	}
	var decoded any
	if err := json.Unmarshal(value.Data, &decoded); err != nil {
		return nil, ErrInvalidCommand
	}
	return decoded, nil
}
