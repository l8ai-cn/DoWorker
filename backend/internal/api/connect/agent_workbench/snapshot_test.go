package agentworkbenchconnect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	sessiondomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	workbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

func TestGetSessionSnapshotRequiresBearer(t *testing.T) {
	repository := &fakeRepository{state: statePointer(testSnapshotState(t, 1, 1))}
	mux := http.NewServeMux()
	Mount(mux, testServer(repository, nil), authOptions(t))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client := connect.NewClient[
		agentworkbenchv2.GetSessionSnapshotRequest,
		agentworkbenchv2.SessionSnapshot,
	](server.Client(), server.URL+GetSessionSnapshotProcedure)

	_, err := client.CallUnary(context.Background(), connect.NewRequest(
		&agentworkbenchv2.GetSessionSnapshotRequest{
			OrgSlug: testOrgSlug, SessionId: testSessionID,
		},
	))

	requireConnectCode(t, err, connect.CodeUnauthenticated)
}

func TestGetSessionSnapshotEnforcesOrganizationAndOwner(t *testing.T) {
	repository := &fakeRepository{state: statePointer(testSnapshotState(t, 1, 1))}
	tests := []struct {
		name    string
		session *sessiondomain.Session
		code    connect.Code
	}{
		{
			name:    "different organization",
			session: activeSession(testUserID, 8),
			code:    connect.CodeNotFound,
		},
		{
			name:    "different owner",
			session: activeSession(99, 7),
			code:    connect.CodePermissionDenied,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer(
				repository,
				nil,
				fakeSessionLookup{session: test.session},
				fakeOrganizationService{member: true},
				nil,
			)
			_, err := server.GetSessionSnapshot(
				ownerContext(),
				connect.NewRequest(&agentworkbenchv2.GetSessionSnapshotRequest{
					OrgSlug: testOrgSlug, SessionId: testSessionID,
				}),
			)
			requireConnectCode(t, err, test.code)
		})
	}
}

func TestGetSessionSnapshotEnforcesEmbedScopeAndReadCapability(t *testing.T) {
	repository := &fakeRepository{state: statePointer(testSnapshotState(t, 1, 1))}
	server := testServer(repository, nil)
	tests := []struct {
		name    string
		ctx     context.Context
		orgSlug string
		code    connect.Code
	}{
		{"missing read", embedContext("write"), testOrgSlug, connect.CodePermissionDenied},
		{"different organization", embedContext("read"), "other", connect.CodeNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := server.GetSessionSnapshot(
				test.ctx,
				connect.NewRequest(&agentworkbenchv2.GetSessionSnapshotRequest{
					OrgSlug: test.orgSlug, SessionId: testSessionID,
				}),
			)
			requireConnectCode(t, err, test.code)
		})
	}
}

func TestGetSessionSnapshotAddsViewerScopedGrants(t *testing.T) {
	state := testSnapshotState(t, 1, 1)
	repository := &fakeRepository{state: &state}
	server := testServer(repository, nil)

	owner, err := server.GetSessionSnapshot(
		ownerContext(),
		connect.NewRequest(&agentworkbenchv2.GetSessionSnapshotRequest{
			OrgSlug: testOrgSlug, SessionId: testSessionID,
		}),
	)
	require.NoError(t, err)
	require.Len(t, owner.Msg.GetGrants(), 1)
	require.ElementsMatch(t, []string{
		"session.send",
		"session.interrupt",
		"session.configure",
		"session.permission.resolve",
		"terminal.input",
		"terminal.resize",
		"terminal.signal",
		"terminal.control",
	}, owner.Msg.GetGrants()[0].GetActions())

	embedded, err := server.GetSessionSnapshot(
		embedContext("read", "write", "approve", "terminal", "control"),
		connect.NewRequest(&agentworkbenchv2.GetSessionSnapshotRequest{
			OrgSlug: testOrgSlug, SessionId: testSessionID,
		}),
	)
	require.NoError(t, err)
	require.Len(t, embedded.Msg.GetGrants(), 1)
	require.ElementsMatch(t, owner.Msg.GetGrants()[0].GetActions(), embedded.Msg.GetGrants()[0].GetActions())
	require.NotEqual(t, owner.Msg.GetGrants()[0].GetSubject(), embedded.Msg.GetGrants()[0].GetSubject())
}

func TestGetSessionSnapshotCreatesPersistentInitialSnapshot(t *testing.T) {
	repository := &fakeRepository{}
	server := testServer(repository, nil)

	response, err := server.GetSessionSnapshot(
		ownerContext(),
		connect.NewRequest(&agentworkbenchv2.GetSessionSnapshotRequest{
			OrgSlug: testOrgSlug, SessionId: testSessionID,
		}),
	)

	require.NoError(t, err)
	require.Equal(t, testSessionID, response.Msg.GetSessionId())
	require.NotEmpty(t, response.Msg.GetStreamEpoch())
	require.Equal(t, uint64(0), response.Msg.GetRevision())
	require.Equal(t, uint64(0), response.Msg.GetLatestSequence())
	require.NotEmpty(t, response.Msg.GetDigest())
	require.Equal(t, 1, repository.ensureCalls)
}

func TestGetSessionSnapshotDecodesExistingSnapshot(t *testing.T) {
	state := testSnapshotState(t, 2, 3)
	repository := &fakeRepository{state: &state}

	response, err := testServer(repository, nil).GetSessionSnapshot(
		ownerContext(),
		connect.NewRequest(&agentworkbenchv2.GetSessionSnapshotRequest{
			OrgSlug: testOrgSlug, SessionId: testSessionID,
		}),
	)

	require.NoError(t, err)
	require.Equal(t, uint64(2), response.Msg.GetRevision())
	require.Equal(t, uint64(3), response.Msg.GetLatestSequence())
	expectedDigest, digestErr := snapshotDigest(response.Msg)
	require.NoError(t, digestErr)
	require.Equal(t, expectedDigest, response.Msg.GetDigest())
	require.NotEqual(t, state.Digest, response.Msg.GetDigest())
	require.Zero(t, repository.ensureCalls)
}

func TestGetSessionSnapshotRejectsProjectionMetadataMismatch(t *testing.T) {
	state := testSnapshotState(t, 2, 3)
	state.Revision = 4
	repository := &fakeRepository{state: &state}

	_, err := testServer(repository, nil).GetSessionSnapshot(
		ownerContext(),
		connect.NewRequest(&agentworkbenchv2.GetSessionSnapshotRequest{
			OrgSlug: testOrgSlug, SessionId: testSessionID,
		}),
	)

	requireConnectCode(t, err, connect.CodeDataLoss)
}

func TestGetSessionSnapshotRejectsImpossibleRevisionSequence(t *testing.T) {
	state := testSnapshotState(t, 2, 1)
	repository := &fakeRepository{state: &state}

	_, err := testServer(repository, nil).GetSessionSnapshot(
		ownerContext(),
		connect.NewRequest(&agentworkbenchv2.GetSessionSnapshotRequest{
			OrgSlug: testOrgSlug, SessionId: testSessionID,
		}),
	)

	requireConnectCode(t, err, connect.CodeDataLoss)
}

func statePointer(state workbenchdomain.SessionState) *workbenchdomain.SessionState {
	return &state
}
