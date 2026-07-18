package agentworkbenchconnect

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	sessiondomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	workbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	workbenchsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentworkbench"
	authpkg "github.com/anthropics/agentsmesh/backend/pkg/auth"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

const (
	testJWTSecret = "agent-workbench-connect-test-secret"
	testOrgSlug   = "acme"
	testSessionID = "conv_1"
	testEpoch     = "epoch-1"
	testUserID    = int64(42)
	testAudience  = "agentsmesh-api"
)

var (
	testManagerOnce sync.Once
	testManager     *authpkg.AccessTokenManager
	testManagerErr  error
)

type fakeOrganization struct{}

func (fakeOrganization) GetID() int64    { return 7 }
func (fakeOrganization) GetSlug() string { return testOrgSlug }
func (fakeOrganization) GetName() string { return "Acme" }

type fakeOrganizationService struct {
	member bool
}

func (f fakeOrganizationService) GetBySlug(
	context.Context,
	string,
) (middleware.OrganizationGetter, error) {
	return fakeOrganization{}, nil
}

func (f fakeOrganizationService) IsMember(context.Context, int64, int64) (bool, error) {
	return f.member, nil
}

func (f fakeOrganizationService) GetMemberRole(context.Context, int64, int64) (string, error) {
	if !f.member {
		return "", context.Canceled
	}
	return "member", nil
}

type fakeSessionLookup struct {
	session *sessiondomain.Session
	err     error
}

func (f fakeSessionLookup) Get(context.Context, string) (*sessiondomain.Session, error) {
	return f.session, f.err
}

type fakeRepository struct {
	mu               sync.Mutex
	state            *workbenchdomain.SessionState
	events           []workbenchdomain.Event
	ensureCalls      int
	getCalls         int
	blockGetCall     int
	blockGetEntered  chan struct{}
	blockGetRelease  chan struct{}
	blockGetSignaled bool
}

func (f *fakeRepository) Append(
	context.Context,
	workbenchdomain.AppendRequest,
) (workbenchdomain.AppendResult, error) {
	return workbenchdomain.AppendResult{}, nil
}

func (f *fakeRepository) GetSnapshot(
	ctx context.Context,
	_ string,
) (*workbenchdomain.SessionState, error) {
	f.mu.Lock()
	f.getCalls++
	call := f.getCalls
	shouldBlock := call == f.blockGetCall
	if shouldBlock && !f.blockGetSignaled {
		close(f.blockGetEntered)
		f.blockGetSignaled = true
	}
	state := cloneState(f.state)
	release := f.blockGetRelease
	f.mu.Unlock()
	if shouldBlock {
		select {
		case <-release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return state, nil
}

func (f *fakeRepository) EnsureSnapshot(
	_ context.Context,
	initial workbenchdomain.SessionState,
) (*workbenchdomain.SessionState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ensureCalls++
	if f.state == nil {
		f.state = cloneState(&initial)
	}
	return cloneState(f.state), nil
}

func (f *fakeRepository) ListAfter(
	_ context.Context,
	sessionID string,
	streamEpoch string,
	sequence uint64,
	limit int,
) ([]workbenchdomain.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make([]workbenchdomain.Event, 0, limit)
	for _, event := range f.events {
		if event.SessionID == sessionID &&
			event.StreamEpoch == streamEpoch &&
			event.Sequence > sequence {
			result = append(result, event)
			if len(result) == limit {
				break
			}
		}
	}
	return result, nil
}

func (*fakeRepository) PutCommandReceipt(
	context.Context,
	workbenchdomain.CommandReceipt,
) (*workbenchdomain.CommandReceipt, error) {
	return nil, nil
}

func (*fakeRepository) GetCommandReceipt(
	context.Context,
	string,
	string,
) (*workbenchdomain.CommandReceipt, error) {
	return nil, nil
}

func cloneState(state *workbenchdomain.SessionState) *workbenchdomain.SessionState {
	if state == nil {
		return nil
	}
	cloned := *state
	cloned.Projection = append([]byte(nil), state.Projection...)
	return &cloned
}

func testServer(
	repository PersistenceRepository,
	executor CommandExecutor,
) *Server {
	return NewServer(
		repository,
		workbenchsvc.NewDeltaHub(8),
		fakeSessionLookup{session: activeSession(testUserID, 7)},
		fakeOrganizationService{member: true},
		executor,
	)
}

func activeSession(userID, orgID int64) *sessiondomain.Session {
	return &sessiondomain.Session{
		ID:             testSessionID,
		OrganizationID: orgID,
		UserID:         userID,
		PodKey:         "pod-1",
	}
}

func authorizedRequest[T any](t *testing.T, message *T) *connect.Request[T] {
	t.Helper()
	token, err := workbenchTestAccessTokenManager(t).GenerateToken(
		testUserID,
		"owner@example.com",
		"owner",
		7,
		"member",
	)
	require.NoError(t, err)
	request := connect.NewRequest(message)
	request.Header().Set("Authorization", "Bearer "+token)
	return request
}

func authOptions(t *testing.T) connect.HandlerOption {
	return connect.WithInterceptors(NewAuthInterceptor(
		workbenchTestAccessTokenManager(t),
		testAudience,
		nil,
	))
}

func workbenchTestAccessTokenManager(t *testing.T) *authpkg.AccessTokenManager {
	testManagerOnce.Do(func() {
		var privateKey *rsa.PrivateKey
		privateKey, testManagerErr = rsa.GenerateKey(rand.Reader, 2048)
		if testManagerErr != nil {
			return
		}
		testManager, testManagerErr = authpkg.NewAccessTokenManager(
			authpkg.AccessTokenConfig{
				PrivateKey: privateKey,
				PublicKey:  &privateKey.PublicKey,
				KeyID:      "agent-workbench-test",
				Issuer:     "agent-workbench-test",
				Audiences:  []string{testAudience},
				Duration:   time.Hour,
			},
		)
	})
	if t != nil {
		t.Helper()
		require.NoError(t, testManagerErr)
		require.NotNil(t, testManager)
	}
	return testManager
}

func testSnapshotState(
	t *testing.T,
	revision uint64,
	sequence uint64,
) workbenchdomain.SessionState {
	t.Helper()
	snapshot := &agentworkbenchv2.SessionSnapshot{
		SessionId:      testSessionID,
		StreamEpoch:    testEpoch,
		Revision:       revision,
		LatestSequence: sequence,
		Status:         agentworkbenchv2.SessionStatus_SESSION_STATUS_IDLE,
		Capabilities:   testCapabilities(),
	}
	digest := testProtoDigest(t, snapshot)
	snapshot.Digest = &digest
	projection, err := (proto.MarshalOptions{Deterministic: true}).Marshal(snapshot)
	require.NoError(t, err)
	return workbenchdomain.SessionState{
		SessionID:      testSessionID,
		StreamEpoch:    testEpoch,
		Revision:       revision,
		LatestSequence: sequence,
		Projection:     projection,
		Digest:         digest,
	}
}

func testCapabilities() *agentworkbenchv2.SupportCapabilities {
	return &agentworkbenchv2.SupportCapabilities{
		ProtocolVersion: "2",
		CommandSchemas: []*agentworkbenchv2.CapabilityDescriptor{
			{
				Namespace: "proto.agent_workbench.v2", SemanticKey: "send_prompt",
				SchemaVersion: "2", Actions: []string{"session.send"},
			},
			{
				Namespace: "proto.agent_workbench.v2", SemanticKey: "interrupt",
				SchemaVersion: "2", Actions: []string{"session.interrupt"},
			},
			{
				Namespace: "proto.agent_workbench.v2", SemanticKey: "change_configuration",
				SchemaVersion: "2", Actions: []string{"session.configure"},
			},
			{
				Namespace: "proto.agent_workbench.v2", SemanticKey: "resolve_permission",
				SchemaVersion: "2", Actions: []string{"session.permission.resolve"},
			},
		},
		TerminalOperations: []string{
			"terminal.input",
			"terminal.resize",
			"terminal.signal",
			"terminal.control",
		},
		ArtifactOperations: []string{
			"image.edit",
			"presentation.regenerate_slide",
		},
	}
}

func testEvent(
	t *testing.T,
	revision uint64,
	sequence uint64,
) workbenchdomain.Event {
	t.Helper()
	event := &agentworkbenchv2.AgentEvent{
		Envelope: &agentworkbenchv2.EventEnvelope{
			SessionId:   testSessionID,
			StreamEpoch: testEpoch,
			Revision:    revision,
			Sequence:    sequence,
			ItemId:      "status-" + string(rune('0'+sequence)),
			CreatedAt:   "2026-07-16T10:00:00Z",
		},
		Event: &agentworkbenchv2.AgentEvent_SessionStatusChanged{
			SessionStatusChanged: &agentworkbenchv2.SessionStatusChanged{
				Status: agentworkbenchv2.SessionStatus_SESSION_STATUS_RUNNING,
			},
		},
	}
	payload, err := (proto.MarshalOptions{Deterministic: true}).Marshal(event)
	require.NoError(t, err)
	return workbenchdomain.Event{
		SessionID:   testSessionID,
		StreamEpoch: testEpoch,
		Revision:    revision,
		Sequence:    sequence,
		Payload:     payload,
		Digest:      testProtoDigest(t, event),
	}
}

func testDelta(
	t *testing.T,
	revision uint64,
	sequence uint64,
) *agentworkbenchv2.SessionDeltaBatch {
	t.Helper()
	eventRecord := testEvent(t, revision, sequence)
	event := &agentworkbenchv2.AgentEvent{}
	require.NoError(t, proto.Unmarshal(eventRecord.Payload, event))
	delta := &agentworkbenchv2.SessionDeltaBatch{
		SessionId:     testSessionID,
		StreamEpoch:   testEpoch,
		BaseRevision:  revision - 1,
		Revision:      revision,
		FirstSequence: sequence,
		LastSequence:  sequence,
		Events:        []*agentworkbenchv2.AgentEvent{event},
	}
	delta.Digest = testProtoDigest(t, delta)
	return delta
}

func testArtifactEvent(
	t *testing.T,
	revision uint64,
	sequence uint64,
) workbenchdomain.Event {
	t.Helper()
	event := &agentworkbenchv2.AgentEvent{
		Envelope: &agentworkbenchv2.EventEnvelope{
			SessionId:   testSessionID,
			StreamEpoch: testEpoch,
			Revision:    revision,
			Sequence:    sequence,
			ItemId:      "artifact:image-edit-1",
			CreatedAt:   "2026-07-16T10:00:00Z",
		},
		Event: &agentworkbenchv2.AgentEvent_ArtifactChanged{
			ArtifactChanged: &agentworkbenchv2.ArtifactChanged{
				Artifact: testImageEditArtifact(),
			},
		},
	}
	payload, err := (proto.MarshalOptions{Deterministic: true}).Marshal(event)
	require.NoError(t, err)
	return workbenchdomain.Event{
		SessionID: testSessionID, StreamEpoch: testEpoch,
		Revision: revision, Sequence: sequence,
		Payload: payload, Digest: testProtoDigest(t, event),
	}
}

func testImageEditArtifact() *agentworkbenchv2.ArtifactDescriptor {
	result := "result"
	return &agentworkbenchv2.ArtifactDescriptor{
		ArtifactId: "image-edit-1",
		Revision:   7,
		Filename:   "edited.png",
		MediaType:  "image/png",
		Status:     agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
		Representations: []*agentworkbenchv2.ArtifactRepresentation{
			{RepresentationId: "source", Revision: 7, MediaType: "image/png"},
			{RepresentationId: "result", Revision: 7, MediaType: "image/png"},
		},
		Manifest: &agentworkbenchv2.ArtifactManifest{
			Manifest: &agentworkbenchv2.ArtifactManifest_ImageEdit{
				ImageEdit: &agentworkbenchv2.ImageEditManifest{
					SourceRepresentationId: "source",
					ResultRepresentationId: &result,
					SourceWidth:            100,
					SourceHeight:           100,
				},
			},
		},
	}
}

func testProtoDigest(t *testing.T, message proto.Message) string {
	t.Helper()
	data, err := (proto.MarshalOptions{Deterministic: true}).Marshal(message)
	require.NoError(t, err)
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
