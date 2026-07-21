package agentworkbench

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	poddomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	sessiondomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	domainworkbench "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	sessionmessagesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessionmessage"
	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type commandRepository struct {
	state    *domainworkbench.SessionState
	receipt  *domainworkbench.CommandReceipt
	appended domainworkbench.AppendRequest
}

func (repo *commandRepository) Append(
	_ context.Context,
	request domainworkbench.AppendRequest,
) (domainworkbench.AppendResult, error) {
	repo.appended = request
	repo.state = &request.Projection
	if len(request.Receipts) > 0 {
		receipt := request.Receipts[len(request.Receipts)-1]
		repo.receipt = &receipt
	}
	return domainworkbench.AppendResult{Applied: true}, nil
}

func (repo *commandRepository) GetSnapshot(
	context.Context,
	string,
) (*domainworkbench.SessionState, error) {
	return repo.state, nil
}

func (*commandRepository) ListAfter(
	context.Context,
	string,
	string,
	uint64,
	int,
) ([]domainworkbench.Event, error) {
	return nil, nil
}

func (repo *commandRepository) PutCommandReceipt(
	_ context.Context,
	receipt domainworkbench.CommandReceipt,
) (*domainworkbench.CommandReceipt, error) {
	stored := receipt
	repo.receipt = &stored
	return repo.receipt, nil
}

func (repo *commandRepository) GetCommandReceipt(
	context.Context,
	string,
	string,
) (*domainworkbench.CommandReceipt, error) {
	return repo.receipt, nil
}

type commandPodLookup struct {
	pod *poddomain.Pod
}

func (lookup commandPodLookup) GetByKey(
	context.Context,
	string,
) (*poddomain.Pod, error) {
	return lookup.pod, nil
}

type commandPromptOutbox struct {
	input sessionmessagesvc.PromptInput
}

func (outbox *commandPromptOutbox) PersistAndQueue(
	_ context.Context,
	input sessionmessagesvc.PromptInput,
) error {
	outbox.input = input
	return nil
}

type commandACPSender struct {
	payload string
}

func (sender *commandACPSender) SendAcpRelay(
	_ context.Context,
	_ int64,
	_ string,
	payload string,
) error {
	sender.payload = payload
	return nil
}

func TestDeliverConfigurationMapsRunnerWireFields(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		wantPayload string
	}{
		{
			name:  "model",
			key:   "model",
			value: "gpt-5.5",
			wantPayload: `{
				"type":"set_model",
				"model":"gpt-5.5",
				"requestId":"command-config"
			}`,
		},
		{
			name:  "permission mode",
			key:   "permission_mode",
			value: "bypass",
			wantPayload: `{
				"type":"set_permission_mode",
				"mode":"bypass",
				"requestId":"command-config"
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sender := &commandACPSender{}
			dispatcher := &CommandDispatcher{acp: sender}
			command := &agentworkbenchv2.CommandEnvelope{CommandId: "command-config"}
			data, err := json.Marshal(test.value)
			require.NoError(t, err)
			value := &agentworkbenchv2.ChangeConfigurationCommand{
				Values: []*agentworkbenchv2.ConfigurationValue{{
					Key: test.key,
					Value: &agentworkbenchv2.StructuredPayload{
						MediaType: "application/json",
						Data:      data,
					},
				}},
			}

			err = dispatcher.deliverConfiguration(
				context.Background(),
				17,
				"pod-1",
				command,
				value,
			)

			require.NoError(t, err)
			require.JSONEq(t, test.wantPayload, sender.payload)
		})
	}
}

func TestCommandDispatcherQueuesPromptAndAppendsAcceptedReceipt(t *testing.T) {
	repo := &commandRepository{state: storedInitialSnapshot(t)}
	outbox := &commandPromptOutbox{}
	publisher := &ingressPublisher{}
	dispatcher, err := NewCommandDispatcher(
		repo,
		commandPodLookup{pod: &poddomain.Pod{
			PodKey: "pod-1", RunnerID: 17, OrganizationID: 7,
		}},
		outbox,
		&commandACPSender{},
		publisher,
		func() time.Time { return time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC) },
	)
	require.NoError(t, err)
	command := promptCommand(t, "创建一个视频预览")

	receipt, err := dispatcher.Execute(
		context.Background(),
		&sessiondomain.Session{ID: "conv_1", PodKey: "pod-1", OrganizationID: 7},
		command,
	)
	require.NoError(t, err)
	require.Equal(
		t,
		agentworkbenchv2.CommandReceiptState_COMMAND_RECEIPT_STATE_ACCEPTED,
		receipt.State,
	)
	require.Equal(t, "创建一个视频预览", outbox.input.Prompt)
	require.Equal(t, "conv_1", outbox.input.Item.SessionID)
	require.Len(t, repo.appended.Events, 2)
	require.Len(t, repo.appended.Receipts, 1)
	require.Equal(t, uint64(1), publisher.delta.Revision)
}

func TestCommandDispatcherReturnsPersistedAcceptedReceiptWithoutRedispatch(t *testing.T) {
	accepted := acceptedReceipt(promptCommand(t, "hello"))
	encoded, err := marshalDeterministic(accepted)
	require.NoError(t, err)
	repo := &commandRepository{
		state: storedInitialSnapshot(t),
		receipt: &domainworkbench.CommandReceipt{
			SessionID: "conv_1", CommandID: "command-1",
			PayloadDigest: accepted.PayloadDigest,
			State:         domainworkbench.ReceiptStateAccepted, Receipt: encoded,
		},
	}
	outbox := &commandPromptOutbox{}
	dispatcher, err := NewCommandDispatcher(
		repo,
		commandPodLookup{pod: &poddomain.Pod{PodKey: "pod-1", RunnerID: 17}},
		outbox,
		&commandACPSender{},
		&ingressPublisher{},
		time.Now,
	)
	require.NoError(t, err)

	receipt, err := dispatcher.Execute(
		context.Background(),
		&sessiondomain.Session{ID: "conv_1", PodKey: "pod-1"},
		promptCommand(t, "hello"),
	)
	require.NoError(t, err)
	require.Equal(t, accepted.State, receipt.State)
	require.Empty(t, outbox.input.PodKey)
	require.Empty(t, repo.appended.SessionID)
}

func TestCommandDispatcherRejectsStaleRevisionBeforeDelivery(t *testing.T) {
	repo := &commandRepository{state: storedInitialSnapshot(t)}
	outbox := &commandPromptOutbox{}
	dispatcher, err := NewCommandDispatcher(
		repo,
		commandPodLookup{pod: &poddomain.Pod{PodKey: "pod-1", RunnerID: 17}},
		outbox,
		&commandACPSender{},
		&ingressPublisher{},
		time.Now,
	)
	require.NoError(t, err)
	command := promptCommand(t, "hello")
	staleRevision := uint64(7)
	command.ExpectedRevision = &staleRevision
	command.PayloadDigest, err = CommandPayloadDigest(command)
	require.NoError(t, err)

	_, err = dispatcher.Execute(
		context.Background(),
		&sessiondomain.Session{ID: "conv_1", PodKey: "pod-1"},
		command,
	)

	require.ErrorIs(t, err, domainworkbench.ErrRevisionConflict)
	require.Empty(t, outbox.input.PodKey)
	require.Nil(t, repo.receipt)
}

func TestCommandDispatcherPersistsDeliveryFailure(t *testing.T) {
	repo := &commandRepository{state: storedInitialSnapshot(t)}
	dispatcher, err := NewCommandDispatcher(
		repo,
		commandPodLookup{},
		&commandPromptOutbox{},
		&commandACPSender{},
		&ingressPublisher{},
		time.Now,
	)
	require.NoError(t, err)

	receipt, err := dispatcher.Execute(
		context.Background(),
		&sessiondomain.Session{ID: "conv_1", PodKey: "pod-1"},
		promptCommand(t, "hello"),
	)

	require.NoError(t, err)
	require.Equal(
		t,
		agentworkbenchv2.CommandReceiptState_COMMAND_RECEIPT_STATE_FAILED,
		receipt.State,
	)
	require.True(t, receipt.GetError().GetRetryable())
}

func storedInitialSnapshot(t *testing.T) *domainworkbench.SessionState {
	t.Helper()
	snapshot := initialSnapshot("conv_1", "stream-1")
	snapshotDigest, err := deterministicDigest(snapshot)
	require.NoError(t, err)
	snapshot.Digest = stringPointer(snapshotDigest)
	encoded, err := proto.Marshal(snapshot)
	require.NoError(t, err)
	return &domainworkbench.SessionState{
		SessionID: "conv_1", StreamEpoch: "stream-1",
		Projection: encoded, Digest: snapshotDigest,
	}
}

func promptCommand(
	t *testing.T,
	text string,
) *agentworkbenchv2.CommandEnvelope {
	t.Helper()
	command := &agentworkbenchv2.CommandEnvelope{
		SessionId: "conv_1", StreamEpoch: "stream-1", CommandId: "command-1",
		IssuedAt: "2026-07-16T10:00:00Z",
		Command: &agentworkbenchv2.CommandEnvelope_SendPrompt{
			SendPrompt: &agentworkbenchv2.SendPromptCommand{Text: text},
		},
	}
	digest, err := CommandPayloadDigest(command)
	require.NoError(t, err)
	command.PayloadDigest = digest
	return command
}
