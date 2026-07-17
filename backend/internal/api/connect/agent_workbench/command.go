package agentworkbenchconnect

import (
	"context"

	"connectrpc.com/connect"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
)

func (s *Server) ExecuteCommand(
	ctx context.Context,
	request *connect.Request[agentworkbenchv2.ExecuteCommandRequest],
) (*connect.Response[agentworkbenchv2.CommandReceipt], error) {
	command := request.Msg.GetCommand()
	if command == nil ||
		command.GetSessionId() == "" ||
		command.GetCommandId() == "" ||
		command.GetPayloadDigest() == "" ||
		command.GetCommand() == nil {
		return nil, invalidArgument("command is incomplete")
	}
	ctx, session, err := s.authorizeSession(ctx, request.Msg, command.GetSessionId())
	if err != nil {
		return nil, err
	}
	if err := requireEmbedCommandCapability(embedAccess(ctx), command); err != nil {
		return nil, err
	}
	snapshot, err := s.loadSnapshot(ctx, command.GetSessionId())
	if err != nil {
		return nil, err
	}
	authorization, err := viewerAuthorizationFor(ctx, session)
	if err != nil {
		return nil, err
	}
	if err := authorization.decorateSnapshot(snapshot); err != nil {
		return nil, internalError(err)
	}
	if artifact := command.GetArtifactAction(); artifact != nil {
		if err := requireArtifactAuthorization(snapshot, artifact); err != nil {
			return nil, err
		}
	}
	if command.GetStreamEpoch() != snapshot.GetStreamEpoch() {
		return nil, failedPrecondition(
			"agent workbench command stream is stale; resync required",
		)
	}
	if command.ExpectedRevision != nil &&
		command.GetExpectedRevision() != snapshot.GetRevision() {
		return nil, failedPrecondition(
			"agent workbench command revision is stale; resync required",
		)
	}
	if s.executor == nil {
		return nil, unavailable("agent workbench command executor is unavailable")
	}
	receipt, err := s.executor.Execute(ctx, session, command)
	if err != nil {
		return nil, commandExecutionError(err)
	}
	if receipt == nil ||
		receipt.GetSessionId() != command.GetSessionId() ||
		receipt.GetCommandId() != command.GetCommandId() ||
		receipt.GetPayloadDigest() != command.GetPayloadDigest() ||
		receipt.GetState() ==
			agentworkbenchv2.CommandReceiptState_COMMAND_RECEIPT_STATE_UNSPECIFIED {
		return nil, dataLoss("agent workbench command receipt is inconsistent")
	}
	return connect.NewResponse(receipt), nil
}
