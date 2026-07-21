package agentworkbench

import (
	"context"
	"fmt"

	domainworkbench "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

func (dispatcher *CommandDispatcher) storeReceivedReceipt(
	ctx context.Context,
	command *agentworkbenchv2.CommandEnvelope,
) (*agentworkbenchv2.CommandReceipt, error) {
	now := dispatcher.now().UTC().Format(timeFormat)
	receipt := &agentworkbenchv2.CommandReceipt{
		SessionId: command.SessionId, CommandId: command.CommandId,
		State:         agentworkbenchv2.CommandReceiptState_COMMAND_RECEIPT_STATE_RECEIVED,
		PayloadDigest: command.PayloadDigest,
		ReceivedAt:    stringPointer(now), UpdatedAt: stringPointer(now),
	}
	return dispatcher.putReceipt(ctx, receipt)
}

func (dispatcher *CommandDispatcher) storeFailedReceipt(
	ctx context.Context,
	command *agentworkbenchv2.CommandEnvelope,
	failure error,
) (*agentworkbenchv2.CommandReceipt, error) {
	now := dispatcher.now().UTC().Format(timeFormat)
	receipt := &agentworkbenchv2.CommandReceipt{
		SessionId: command.SessionId, CommandId: command.CommandId,
		State:         agentworkbenchv2.CommandReceiptState_COMMAND_RECEIPT_STATE_FAILED,
		PayloadDigest: command.PayloadDigest,
		Error: &agentworkbenchv2.AgentError{
			Code: "command_delivery_failed", Message: failure.Error(), Retryable: true,
		},
		UpdatedAt: stringPointer(now),
	}
	stored, err := dispatcher.putReceipt(ctx, receipt)
	if err != nil {
		return nil, err
	}
	return stored, nil
}

func (dispatcher *CommandDispatcher) putReceipt(
	ctx context.Context,
	receipt *agentworkbenchv2.CommandReceipt,
) (*agentworkbenchv2.CommandReceipt, error) {
	domain, err := domainReceipt(receipt)
	if err != nil {
		return nil, err
	}
	stored, err := dispatcher.repository.PutCommandReceipt(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("persist agent workbench command receipt: %w", err)
	}
	return decodeDomainReceipt(stored)
}

func domainReceipt(
	receipt *agentworkbenchv2.CommandReceipt,
) (domainworkbench.CommandReceipt, error) {
	encoded, err := marshalDeterministic(receipt)
	if err != nil {
		return domainworkbench.CommandReceipt{}, err
	}
	return domainworkbench.CommandReceipt{
		SessionID: receipt.SessionId, CommandID: receipt.CommandId,
		PayloadDigest: receipt.PayloadDigest,
		State:         domainworkbench.ReceiptState(receipt.State),
		Receipt:       encoded,
	}, nil
}

func decodeDomainReceipt(
	receipt *domainworkbench.CommandReceipt,
) (*agentworkbenchv2.CommandReceipt, error) {
	if receipt == nil {
		return nil, ErrInvalidCommand
	}
	decoded := &agentworkbenchv2.CommandReceipt{}
	if err := proto.Unmarshal(receipt.Receipt, decoded); err != nil {
		return nil, fmt.Errorf("decode agent workbench command receipt: %w", err)
	}
	if decoded.SessionId != receipt.SessionID ||
		decoded.CommandId != receipt.CommandID ||
		decoded.PayloadDigest != receipt.PayloadDigest ||
		int16(decoded.State) != int16(receipt.State) {
		return nil, ErrInvalidCommand
	}
	return decoded, nil
}

const timeFormat = "2006-01-02T15:04:05.999999999Z07:00"
