package agentworkbench

import (
	"context"
	"errors"
	"fmt"
	"time"

	poddomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	sessiondomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	domainworkbench "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	sessionfile "github.com/l8ai-cn/agentcloud/backend/internal/service/sessionfile"
	sessionmessagesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessionmessage"
	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
)

var (
	ErrCommandConflict    = errors.New("agent workbench command conflict")
	ErrCommandUnavailable = errors.New("agent workbench command unavailable")
)

type CommandPodLookup interface {
	GetByKey(context.Context, string) (*poddomain.Pod, error)
}

type PromptOutbox interface {
	PersistAndQueue(context.Context, sessionmessagesvc.PromptInput) error
}

type ACPCommandSender interface {
	SendAcpRelay(context.Context, int64, string, string) error
}

type AttachmentDelivery interface {
	Stage(context.Context, sessionfile.AttachmentSandbox, *poddomain.Pod, string, []string) ([]string, error)
}

type CommandDispatcher struct {
	repository  domainworkbench.Repository
	pods        CommandPodLookup
	outbox      PromptOutbox
	acp         ACPCommandSender
	publisher   DeltaPublisher
	attachments AttachmentDelivery
	sandbox     sessionfile.AttachmentSandbox
	now         func() time.Time
}

type CommandDispatcherOption func(*CommandDispatcher)

func WithAttachmentDelivery(
	attachments AttachmentDelivery,
	sandbox sessionfile.AttachmentSandbox,
) CommandDispatcherOption {
	return func(dispatcher *CommandDispatcher) {
		dispatcher.attachments = attachments
		dispatcher.sandbox = sandbox
	}
}

func NewCommandDispatcher(
	repository domainworkbench.Repository,
	pods CommandPodLookup,
	outbox PromptOutbox,
	acp ACPCommandSender,
	publisher DeltaPublisher,
	now func() time.Time,
	options ...CommandDispatcherOption,
) (*CommandDispatcher, error) {
	if repository == nil || pods == nil || outbox == nil ||
		acp == nil || publisher == nil || now == nil {
		return nil, ErrIngressConfiguration
	}
	dispatcher := &CommandDispatcher{
		repository: repository, pods: pods, outbox: outbox,
		acp: acp, publisher: publisher, now: now,
	}
	for _, option := range options {
		option(dispatcher)
	}
	return dispatcher, nil
}

func (dispatcher *CommandDispatcher) Execute(
	ctx context.Context,
	session *sessiondomain.Session,
	command *agentworkbenchv2.CommandEnvelope,
) (*agentworkbenchv2.CommandReceipt, error) {
	if err := validateCommandEnvelope(session, command); err != nil {
		return nil, err
	}
	if err := dispatcher.validateCommandPosition(ctx, command); err != nil {
		return nil, err
	}
	existing, err := dispatcher.repository.GetCommandReceipt(
		ctx,
		session.ID,
		command.CommandId,
	)
	if err != nil {
		return nil, fmt.Errorf("load agent workbench command receipt: %w", err)
	}
	if existing != nil {
		receipt, err := decodeDomainReceipt(existing)
		if err != nil {
			return nil, err
		}
		if receipt.PayloadDigest != command.PayloadDigest {
			return nil, ErrCommandConflict
		}
		if receipt.State != agentworkbenchv2.CommandReceiptState_COMMAND_RECEIPT_STATE_RECEIVED {
			return receipt, nil
		}
	} else if _, err := dispatcher.storeReceivedReceipt(ctx, command); err != nil {
		return nil, err
	}
	if err := dispatcher.deliver(ctx, session, command); err != nil {
		return dispatcher.storeFailedReceipt(ctx, command, err)
	}
	return dispatcher.appendAccepted(ctx, command)
}

func (dispatcher *CommandDispatcher) validateCommandPosition(
	ctx context.Context,
	command *agentworkbenchv2.CommandEnvelope,
) error {
	stored, err := dispatcher.repository.GetSnapshot(ctx, command.SessionId)
	if err != nil {
		return fmt.Errorf("load agent workbench command snapshot: %w", err)
	}
	snapshot, err := decodeStoredSnapshot(command.SessionId, stored)
	if err != nil {
		return err
	}
	if command.StreamEpoch != snapshot.StreamEpoch {
		return domainworkbench.ErrStreamConflict
	}
	if command.ExpectedRevision != nil &&
		*command.ExpectedRevision != snapshot.Revision {
		return domainworkbench.ErrRevisionConflict
	}
	return nil
}

func validateCommandEnvelope(
	session *sessiondomain.Session,
	command *agentworkbenchv2.CommandEnvelope,
) error {
	if session == nil || session.ID == "" || session.PodKey == "" ||
		command == nil || command.SessionId != session.ID ||
		command.StreamEpoch == "" || command.CommandId == "" ||
		command.PayloadDigest == "" || command.IssuedAt == "" ||
		command.Command == nil {
		return ErrInvalidCommand
	}
	if _, err := time.Parse(time.RFC3339Nano, command.IssuedAt); err != nil {
		return ErrInvalidCommand
	}
	digest, err := CommandPayloadDigest(command)
	if err != nil || digest != command.PayloadDigest {
		return ErrCommandConflict
	}
	return nil
}

func CommandPayloadDigest(
	command *agentworkbenchv2.CommandEnvelope,
) (string, error) {
	if command == nil {
		return "", ErrInvalidCommand
	}
	cloned := cloneCommand(command)
	cloned.PayloadDigest = ""
	return deterministicDigest(cloned)
}
