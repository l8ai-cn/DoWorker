package agentworkbenchconnect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	sessiondomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	workbenchdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	workbenchsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentworkbench"
	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type fakeCommandExecutor struct {
	command *agentworkbenchv2.CommandEnvelope
	session *sessiondomain.Session
	err     error
}

func (f *fakeCommandExecutor) Execute(
	_ context.Context,
	session *sessiondomain.Session,
	command *agentworkbenchv2.CommandEnvelope,
) (*agentworkbenchv2.CommandReceipt, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.session = session
	f.command = command
	return &agentworkbenchv2.CommandReceipt{
		SessionId:     command.GetSessionId(),
		CommandId:     command.GetCommandId(),
		PayloadDigest: command.GetPayloadDigest(),
		State:         agentworkbenchv2.CommandReceiptState_COMMAND_RECEIPT_STATE_ACCEPTED,
	}, nil
}

func TestExecuteCommandDelegatesToInjectedExecutor(t *testing.T) {
	repository := &fakeRepository{state: statePointer(testSnapshotState(t, 2, 2))}
	executor := &fakeCommandExecutor{}
	command := testCommand()

	response, err := testServer(repository, executor).ExecuteCommand(
		ownerContext(),
		connect.NewRequest(&agentworkbenchv2.ExecuteCommandRequest{
			OrgSlug: testOrgSlug,
			Command: command,
		}),
	)

	require.NoError(t, err)
	require.Same(t, command, executor.command)
	require.Equal(t, testSessionID, executor.session.ID)
	require.Equal(t, command.GetCommandId(), response.Msg.GetCommandId())
	require.Equal(
		t,
		agentworkbenchv2.CommandReceiptState_COMMAND_RECEIPT_STATE_ACCEPTED,
		response.Msg.GetState(),
	)
}

func TestExecuteCommandRejectsStaleRevisionBeforeDelegation(t *testing.T) {
	repository := &fakeRepository{state: statePointer(testSnapshotState(t, 2, 2))}
	executor := &fakeCommandExecutor{}
	command := testCommand()
	stale := uint64(1)
	command.ExpectedRevision = &stale

	_, err := testServer(repository, executor).ExecuteCommand(
		ownerContext(),
		connect.NewRequest(&agentworkbenchv2.ExecuteCommandRequest{
			OrgSlug: testOrgSlug,
			Command: command,
		}),
	)

	requireConnectCode(t, err, connect.CodeFailedPrecondition)
	require.Nil(t, executor.command)
}

func TestExecuteCommandEnforcesEmbedCapability(t *testing.T) {
	repository := &fakeRepository{state: statePointer(testSnapshotState(t, 2, 2))}
	tests := []struct {
		name         string
		capabilities []string
		code         connect.Code
		delegated    bool
	}{
		{"read only", []string{"read"}, connect.CodePermissionDenied, false},
		{"write", []string{"read", "write"}, connect.CodeUnknown, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := &fakeCommandExecutor{}
			_, err := testServer(repository, executor).ExecuteCommand(
				embedContext(test.capabilities...),
				connect.NewRequest(&agentworkbenchv2.ExecuteCommandRequest{
					OrgSlug: testOrgSlug,
					Command: testCommand(),
				}),
			)
			if test.delegated {
				require.NoError(t, err)
				require.NotNil(t, executor.command)
				return
			}
			requireConnectCode(t, err, test.code)
			require.Nil(t, executor.command)
		})
	}
}

func TestExecuteCommandRequiresTerminalAndControlEmbedCapabilities(t *testing.T) {
	repository := &fakeRepository{state: statePointer(testSnapshotState(t, 2, 2))}
	tests := []struct {
		name         string
		capabilities []string
		delegated    bool
	}{
		{"control only", []string{"read", "control"}, false},
		{"terminal only", []string{"read", "terminal"}, false},
		{"terminal control", []string{"read", "terminal", "control"}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := &fakeCommandExecutor{}
			_, err := testServer(repository, executor).ExecuteCommand(
				embedContext(test.capabilities...),
				connect.NewRequest(&agentworkbenchv2.ExecuteCommandRequest{
					OrgSlug: testOrgSlug,
					Command: terminalCommand(),
				}),
			)
			if test.delegated {
				require.NoError(t, err)
				require.NotNil(t, executor.command)
				return
			}
			requireConnectCode(t, err, connect.CodePermissionDenied)
			require.Nil(t, executor.command)
		})
	}
}

func TestExecuteArtifactActionRequiresExactArtifactGrant(t *testing.T) {
	state := testSnapshotState(t, 2, 2)
	state = withImageEditArtifact(t, state)
	repository := &fakeRepository{state: &state}
	tests := []struct {
		name       string
		ctx        context.Context
		actionType string
		delegated  bool
	}{
		{"owner exact action", ownerContext(), "image.edit", true},
		{"embed exact action", embedContext("read", "write"), "image.edit", true},
		{"embed without write", embedContext("read", "control"), "image.edit", false},
		{"unknown action", ownerContext(), "image.delete", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := &fakeCommandExecutor{}
			_, err := testServer(repository, executor).ExecuteCommand(
				test.ctx,
				connect.NewRequest(&agentworkbenchv2.ExecuteCommandRequest{
					OrgSlug: testOrgSlug,
					Command: artifactCommand(test.actionType),
				}),
			)
			if test.delegated {
				require.NoError(t, err)
				require.NotNil(t, executor.command)
				return
			}
			requireConnectCode(t, err, connect.CodePermissionDenied)
			require.Nil(t, executor.command)
		})
	}
}

func TestExecuteArtifactActionRejectsManifestActionMissingFromCapabilities(t *testing.T) {
	state := testSnapshotState(t, 2, 2)
	state = withImageEditArtifact(t, state)
	state = withArtifactOperations(t, state, "artifact.download")
	executor := &fakeCommandExecutor{}

	_, err := testServer(&fakeRepository{state: &state}, executor).ExecuteCommand(
		ownerContext(),
		connect.NewRequest(&agentworkbenchv2.ExecuteCommandRequest{
			OrgSlug: testOrgSlug,
			Command: artifactCommand("image.edit"),
		}),
	)

	requireConnectCode(t, err, connect.CodePermissionDenied)
	require.Nil(t, executor.command)
}

func TestExecuteCommandMapsDomainFailures(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		code connect.Code
	}{
		{"invalid", workbenchsvc.ErrInvalidCommand, connect.CodeInvalidArgument},
		{"conflict", workbenchsvc.ErrCommandConflict, connect.CodeAborted},
		{"unavailable", workbenchsvc.ErrCommandUnavailable, connect.CodeUnavailable},
		{"revision", workbenchdomain.ErrRevisionConflict, connect.CodeFailedPrecondition},
		{"unknown", errors.New("database failed"), connect.CodeInternal},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &fakeRepository{
				state: statePointer(testSnapshotState(t, 2, 2)),
			}
			executor := &fakeCommandExecutor{err: testCase.err}

			_, err := testServer(repository, executor).ExecuteCommand(
				ownerContext(),
				connect.NewRequest(&agentworkbenchv2.ExecuteCommandRequest{
					OrgSlug: testOrgSlug,
					Command: testCommand(),
				}),
			)

			requireConnectCode(t, err, testCase.code)
		})
	}
}

func TestExecuteCommandWithoutExecutorReturnsUnavailable(t *testing.T) {
	repository := &fakeRepository{state: statePointer(testSnapshotState(t, 2, 2))}

	_, err := testServer(repository, nil).ExecuteCommand(
		ownerContext(),
		connect.NewRequest(&agentworkbenchv2.ExecuteCommandRequest{
			OrgSlug: testOrgSlug,
			Command: testCommand(),
		}),
	)

	requireConnectCode(t, err, connect.CodeUnavailable)
}

func testCommand() *agentworkbenchv2.CommandEnvelope {
	return &agentworkbenchv2.CommandEnvelope{
		SessionId:     testSessionID,
		StreamEpoch:   testEpoch,
		CommandId:     "command-1",
		PayloadDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		IssuedAt:      "2026-07-16T10:00:00Z",
		Command: &agentworkbenchv2.CommandEnvelope_SendPrompt{
			SendPrompt: &agentworkbenchv2.SendPromptCommand{Text: "hello"},
		},
	}
}

func terminalCommand() *agentworkbenchv2.CommandEnvelope {
	command := testCommand()
	command.Command = &agentworkbenchv2.CommandEnvelope_TerminalOperation{
		TerminalOperation: &agentworkbenchv2.TerminalOperationCommand{
			ResourceId: "terminal-1",
			Operation: &agentworkbenchv2.TerminalOperationCommand_Input{
				Input: &agentworkbenchv2.TerminalInput{Data: []byte("pwd\n")},
			},
		},
	}
	return command
}

func artifactCommand(actionType string) *agentworkbenchv2.CommandEnvelope {
	command := testCommand()
	command.Command = &agentworkbenchv2.CommandEnvelope_ArtifactAction{
		ArtifactAction: &agentworkbenchv2.ArtifactActionCommand{
			ArtifactId:          "image-edit-1",
			RepresentationId:    "result",
			BaseRevision:        7,
			ClientActionId:      "client-action-1",
			ActionType:          actionType,
			Payload:             &agentworkbenchv2.StructuredPayload{MediaType: "application/json", Data: []byte(`{}`)},
			ActionSchemaVersion: "1",
		},
	}
	return command
}

func withImageEditArtifact(
	t *testing.T,
	state workbenchdomain.SessionState,
) workbenchdomain.SessionState {
	t.Helper()
	snapshot := &agentworkbenchv2.SessionSnapshot{}
	require.NoError(t, proto.Unmarshal(state.Projection, snapshot))
	snapshot.Artifacts = []*agentworkbenchv2.ArtifactDescriptor{testImageEditArtifact()}
	snapshot.Digest = nil
	digest := testProtoDigest(t, snapshot)
	snapshot.Digest = &digest
	projection, err := (proto.MarshalOptions{Deterministic: true}).Marshal(snapshot)
	require.NoError(t, err)
	state.Projection = projection
	state.Digest = digest
	return state
}

func withArtifactOperations(
	t *testing.T,
	state workbenchdomain.SessionState,
	operations ...string,
) workbenchdomain.SessionState {
	t.Helper()
	snapshot := &agentworkbenchv2.SessionSnapshot{}
	require.NoError(t, proto.Unmarshal(state.Projection, snapshot))
	snapshot.Capabilities.ArtifactOperations = operations
	snapshot.Digest = nil
	digest := testProtoDigest(t, snapshot)
	snapshot.Digest = &digest
	projection, err := (proto.MarshalOptions{Deterministic: true}).Marshal(snapshot)
	require.NoError(t, err)
	state.Projection = projection
	state.Digest = digest
	return state
}
