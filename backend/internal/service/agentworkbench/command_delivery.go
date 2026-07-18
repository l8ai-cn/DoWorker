package agentworkbench

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	agentpod "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessiondomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	sessionmessagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionmessage"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
)

func (dispatcher *CommandDispatcher) deliver(
	ctx context.Context,
	session *sessiondomain.Session,
	command *agentworkbenchv2.CommandEnvelope,
) error {
	pod, err := dispatcher.pods.GetByKey(ctx, session.PodKey)
	if err != nil || pod == nil || pod.PodKey != session.PodKey {
		return ErrCommandUnavailable
	}
	switch value := command.Command.(type) {
	case *agentworkbenchv2.CommandEnvelope_SendPrompt:
		if value.SendPrompt == nil ||
			(value.SendPrompt.Text == "" && len(value.SendPrompt.Attachments) == 0) {
			return ErrInvalidCommand
		}
		return dispatcher.deliverPrompt(ctx, session, pod, command, value.SendPrompt)
	case *agentworkbenchv2.CommandEnvelope_Interrupt:
		return dispatcher.sendACP(ctx, pod.RunnerID, pod.PodKey, map[string]any{
			"type": "interrupt", "requestId": command.CommandId,
		})
	case *agentworkbenchv2.CommandEnvelope_ResolvePermission:
		return dispatcher.deliverPermission(ctx, pod.RunnerID, pod.PodKey, command, value.ResolvePermission)
	case *agentworkbenchv2.CommandEnvelope_ChangeConfiguration:
		return dispatcher.deliverConfiguration(ctx, pod.RunnerID, pod.PodKey, command, value.ChangeConfiguration)
	case *agentworkbenchv2.CommandEnvelope_ArtifactAction:
		return dispatcher.deliverArtifactAction(ctx, pod.RunnerID, pod.PodKey, command, value.ArtifactAction)
	case *agentworkbenchv2.CommandEnvelope_Extension:
		return dispatcher.deliverExtension(ctx, pod.RunnerID, pod.PodKey, command, value.Extension)
	default:
		return ErrCommandUnavailable
	}
}

func (dispatcher *CommandDispatcher) deliverPrompt(
	ctx context.Context,
	session *sessiondomain.Session,
	pod *agentpod.Pod,
	command *agentworkbenchv2.CommandEnvelope,
	prompt *agentworkbenchv2.SendPromptCommand,
) error {
	fileIDs, err := attachmentFileIDs(prompt.Attachments)
	if err != nil {
		return err
	}
	var attachmentPaths []string
	if len(fileIDs) > 0 {
		if dispatcher.attachments == nil || dispatcher.sandbox == nil {
			return ErrCommandUnavailable
		}
		attachmentPaths, err = dispatcher.attachments.Stage(
			ctx, dispatcher.sandbox, pod, session.ID, fileIDs,
		)
		if err != nil {
			return err
		}
	}
	promptText := materializedPrompt(prompt.Text, attachmentPaths)
	itemID := stableCommandID("item", session.ID, command.CommandId)
	responseID := stableCommandID("resp", session.ID, command.CommandId)
	item, err := sessionmessagesvc.UserItem(
		itemID,
		session.ID,
		responseID,
		promptItemContent(prompt.Text, attachmentPaths),
	)
	if err != nil {
		return err
	}
	return dispatcher.outbox.PersistAndQueue(ctx, sessionmessagesvc.PromptInput{
		OrganizationID: session.OrganizationID, RunnerID: pod.RunnerID,
		PodKey: session.PodKey, Item: item, Prompt: promptText,
	})
}

func (dispatcher *CommandDispatcher) sendACP(
	ctx context.Context,
	runnerID int64,
	podKey string,
	payload map[string]any,
) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return dispatcher.acp.SendAcpRelay(ctx, runnerID, podKey, string(encoded))
}

func stableCommandID(prefix, sessionID, commandID string) string {
	sum := sha256.Sum256([]byte(sessionID + "\x00" + commandID))
	return prefix + "_" + hex.EncodeToString(sum[:16])
}
