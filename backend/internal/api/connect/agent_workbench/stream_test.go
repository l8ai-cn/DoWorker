package agentworkbenchconnect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	workbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	workbenchsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentworkbench"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

func TestStreamSessionDeltasReplaysThenStreamsLive(t *testing.T) {
	repository := &fakeRepository{
		state:           statePointer(testSnapshotState(t, 2, 2)),
		events:          []workbenchdomain.Event{testEvent(t, 1, 1), testEvent(t, 2, 2)},
		blockGetCall:    2,
		blockGetEntered: make(chan struct{}),
		blockGetRelease: make(chan struct{}),
	}
	hub := workbenchsvc.NewDeltaHub(4)
	server := NewServer(
		repository,
		hub,
		fakeSessionLookup{session: activeSession(testUserID, 7)},
		fakeOrganizationService{member: true},
		nil,
	)
	client, closeServer := streamClient(t, server)
	defer closeServer()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.CallServerStream(ctx, authorizedRequest(t,
		&agentworkbenchv2.StreamSessionDeltasRequest{
			OrgSlug: testOrgSlug,
			Cursor: &agentworkbenchv2.SessionCursor{
				SessionId: testSessionID, StreamEpoch: testEpoch,
			},
		},
	))
	require.NoError(t, err)
	defer stream.Close()
	require.True(t, stream.Receive(), "first replay: %v", stream.Err())
	require.Equal(t, uint64(1), stream.Msg().GetRevision())
	require.True(t, stream.Receive(), "second replay: %v", stream.Err())
	require.Equal(t, uint64(2), stream.Msg().GetRevision())

	<-repository.blockGetEntered
	close(repository.blockGetRelease)
	hub.Publish(testSessionID, testDelta(t, 3, 3))

	require.True(t, stream.Receive(), "live delta: %v", stream.Err())
	require.Equal(t, uint64(3), stream.Msg().GetRevision())
	require.Equal(t, uint64(3), stream.Msg().GetFirstSequence())
}

func TestStreamSessionDeltasRejectsCursorMismatch(t *testing.T) {
	repository := &fakeRepository{state: statePointer(testSnapshotState(t, 2, 2))}
	server := testServer(repository, nil)
	client, closeServer := streamClient(t, server)
	defer closeServer()

	stream, err := client.CallServerStream(
		context.Background(),
		authorizedRequest(t, &agentworkbenchv2.StreamSessionDeltasRequest{
			OrgSlug: testOrgSlug,
			Cursor: &agentworkbenchv2.SessionCursor{
				SessionId:   testSessionID,
				StreamEpoch: "stale-epoch",
			},
		}),
	)
	require.NoError(t, err)
	require.False(t, stream.Receive())
	requireConnectCode(t, stream.Err(), connect.CodeFailedPrecondition)
}

func TestStreamSessionDeltasReportsSubscriberLag(t *testing.T) {
	repository := &fakeRepository{
		state:           statePointer(testSnapshotState(t, 1, 1)),
		events:          []workbenchdomain.Event{testEvent(t, 1, 1)},
		blockGetCall:    2,
		blockGetEntered: make(chan struct{}),
		blockGetRelease: make(chan struct{}),
	}
	hub := workbenchsvc.NewDeltaHub(1)
	server := NewServer(
		repository,
		hub,
		fakeSessionLookup{session: activeSession(testUserID, 7)},
		fakeOrganizationService{member: true},
		nil,
	)
	client, closeServer := streamClient(t, server)
	defer closeServer()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.CallServerStream(ctx, authorizedRequest(t,
		&agentworkbenchv2.StreamSessionDeltasRequest{
			OrgSlug: testOrgSlug,
			Cursor: &agentworkbenchv2.SessionCursor{
				SessionId: testSessionID, StreamEpoch: testEpoch,
			},
		},
	))
	require.NoError(t, err)
	require.True(t, stream.Receive(), "replay frame: %v", stream.Err())
	<-repository.blockGetEntered
	hub.Publish(testSessionID, testDelta(t, 2, 2))
	hub.Publish(testSessionID, testDelta(t, 3, 3))
	close(repository.blockGetRelease)

	for stream.Receive() {
	}
	requireConnectCode(t, stream.Err(), connect.CodeFailedPrecondition)
}

func TestStreamSessionDeltasDecoratesArtifactChangesForViewer(t *testing.T) {
	state := testSnapshotState(t, 1, 1)
	state = withImageEditArtifact(t, state)
	repository := &fakeRepository{
		state:  &state,
		events: []workbenchdomain.Event{testArtifactEvent(t, 1, 1)},
	}
	client, closeServer := streamClient(t, testServer(repository, nil))
	defer closeServer()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.CallServerStream(ctx, authorizedRequest(t,
		&agentworkbenchv2.StreamSessionDeltasRequest{
			OrgSlug: testOrgSlug,
			Cursor: &agentworkbenchv2.SessionCursor{
				SessionId: testSessionID, StreamEpoch: testEpoch,
			},
		},
	))
	require.NoError(t, err)
	require.True(t, stream.Receive(), "artifact replay: %v", stream.Err())
	artifact := stream.Msg().GetEvents()[0].GetArtifactChanged().GetArtifact()
	require.Len(t, artifact.GetGrants(), 1)
	require.Contains(t, artifact.GetGrants()[0].GetActions(), "image.edit")
	require.NoError(t, validateDeltaDigest(stream.Msg()))
}

func streamClient(
	t *testing.T,
	server *Server,
) (*connect.Client[
	agentworkbenchv2.StreamSessionDeltasRequest,
	agentworkbenchv2.SessionDeltaBatch,
], func()) {
	t.Helper()
	mux := http.NewServeMux()
	Mount(mux, server, authOptions(t))
	httpServer := httptest.NewServer(mux)
	client := connect.NewClient[
		agentworkbenchv2.StreamSessionDeltasRequest,
		agentworkbenchv2.SessionDeltaBatch,
	](httpServer.Client(), httpServer.URL+StreamSessionDeltasProcedure)
	return client, httpServer.Close
}
