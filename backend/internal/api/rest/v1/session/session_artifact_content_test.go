package sessionapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	workbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type artifactContentSandbox struct {
	commands []*runnerv1.SandboxFsCommand
	data     []byte
}

type artifactContentSnapshots struct {
	state *workbenchdomain.SessionState
}

func (s artifactContentSnapshots) GetSnapshot(
	context.Context,
	string,
) (*workbenchdomain.SessionState, error) {
	return s.state, nil
}

func (s *artifactContentSandbox) IsConnected(int64) bool {
	return true
}

func (s *artifactContentSandbox) Exec(
	_ context.Context,
	_ int64,
	command *runnerv1.SandboxFsCommand,
) (*runnerv1.SandboxFsResultEvent, error) {
	s.commands = append(
		s.commands,
		proto.Clone(command).(*runnerv1.SandboxFsCommand),
	)
	switch command.GetOp() {
	case "stat":
		return &runnerv1.SandboxFsResultEvent{
			FileBytes:   int64(len(s.data)),
			ContentType: "video/mp4",
		}, nil
	case "read_verified_bytes":
		var request verifiedArtifactRead
		if err := json.Unmarshal([]byte(command.GetPayload()), &request); err != nil {
			return nil, err
		}
		digest := fmt.Sprintf("sha256:%x", sha256.Sum256(s.data))
		if request.Digest != digest || request.FileBytes != uint64(len(s.data)) {
			return &runnerv1.SandboxFsResultEvent{
				Error: "artifact integrity mismatch",
			}, nil
		}
		start := command.GetOffset()
		if start < 0 || start > int64(len(s.data)) {
			return nil, fmt.Errorf("invalid offset")
		}
		end := min(start+command.GetLength(), int64(len(s.data)))
		return &runnerv1.SandboxFsResultEvent{
			ContentBytes:  append([]byte(nil), s.data[start:end]...),
			ContentOffset: start,
			ContentType:   "video/mp4",
			FileBytes:     int64(len(s.data)),
			Eof:           end == int64(len(s.data)),
		}, nil
	default:
		return nil, fmt.Errorf("unexpected operation %q", command.GetOp())
	}
}

func TestSessionArtifactContentServesExactByteRange(t *testing.T) {
	deps, _ := relayConnectionTestDeps(t, nil)
	sandbox := &artifactContentSandbox{data: []byte("0123456789")}
	deps.SandboxFs = sandbox

	response := sessionArtifactContentRequest(t, deps, "bytes=2-5")

	require.Equal(t, http.StatusPartialContent, response.Code)
	assert.Equal(t, "2345", response.Body.String())
	assert.Equal(t, "bytes 2-5/10", response.Header().Get("Content-Range"))
	assert.Equal(t, "bytes", response.Header().Get("Accept-Ranges"))
	assert.Equal(t, "video/mp4", response.Header().Get("Content-Type"))
	assert.Equal(t, "4", response.Header().Get("Content-Length"))
	require.Len(t, sandbox.commands, 2)
	assert.Equal(t, "stat", sandbox.commands[0].GetOp())
	assert.Equal(t, "read_verified_bytes", sandbox.commands[1].GetOp())
	assert.Equal(t, int64(2), sandbox.commands[1].GetOffset())
	assert.Equal(t, int64(4), sandbox.commands[1].GetLength())
}

func TestSessionArtifactContentSupportsOpenAndSuffixRanges(t *testing.T) {
	for name, header := range map[string]string{
		"open":   "bytes=7-",
		"suffix": "bytes=-3",
	} {
		t.Run(name, func(t *testing.T) {
			deps, _ := relayConnectionTestDeps(t, nil)
			sandbox := &artifactContentSandbox{data: []byte("0123456789")}
			deps.SandboxFs = sandbox

			response := sessionArtifactContentRequest(t, deps, header)

			require.Equal(t, http.StatusPartialContent, response.Code)
			assert.Equal(t, "789", response.Body.String())
			assert.Equal(t, "bytes 7-9/10", response.Header().Get("Content-Range"))
		})
	}
}

func TestSessionArtifactContentStreamsWholeFileInBoundedChunks(t *testing.T) {
	deps, _ := relayConnectionTestDeps(t, nil)
	data := bytes.Repeat([]byte("v"), 4<<20+3)
	sandbox := &artifactContentSandbox{data: data}
	deps.SandboxFs = sandbox

	response := sessionArtifactContentRequest(t, deps, "")

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, data, response.Body.Bytes())
	require.Len(t, sandbox.commands, 3)
	assert.Equal(t, int64(4<<20), sandbox.commands[1].GetLength())
	assert.Equal(t, int64(4<<20), sandbox.commands[2].GetOffset())
	assert.Equal(t, int64(3), sandbox.commands[2].GetLength())
}

func TestSessionArtifactContentRejectsInvalidRanges(t *testing.T) {
	for _, test := range []struct {
		name   string
		header string
		calls  int
	}{
		{name: "multiple", header: "bytes=0-1,4-5", calls: 0},
		{name: "past end", header: "bytes=10-", calls: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			deps, _ := relayConnectionTestDeps(t, nil)
			sandbox := &artifactContentSandbox{data: []byte("0123456789")}
			deps.SandboxFs = sandbox

			response := sessionArtifactContentRequest(t, deps, test.header)

			assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, response.Code)
			assert.Len(t, sandbox.commands, test.calls)
		})
	}
}

func TestSessionArtifactContentServesEmptyFile(t *testing.T) {
	deps, _ := relayConnectionTestDeps(t, nil)
	sandbox := &artifactContentSandbox{}
	deps.SandboxFs = sandbox

	response := sessionArtifactContentRequest(t, deps, "")

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Empty(t, response.Body.Bytes())
	assert.Equal(t, "0", response.Header().Get("Content-Length"))
	assert.Len(t, sandbox.commands, 1)
}

func TestSessionArtifactContentRejectsSameSizeReplacement(t *testing.T) {
	deps, _ := relayConnectionTestDeps(t, nil)
	original := []byte("original-video")
	sandbox := &artifactContentSandbox{data: []byte("replaced-video")}
	deps.SandboxFs = sandbox
	deps.ArtifactSnapshots = artifactContentSnapshots{
		state: artifactContentSnapshot(t, original),
	}

	response := sessionArtifactContentRequest(t, deps, "")

	assert.Equal(t, http.StatusConflict, response.Code)
	assert.NotContains(t, response.Body.String(), "replaced-video")
	require.Len(t, sandbox.commands, 2)
	assert.Equal(t, "read_verified_bytes", sandbox.commands[1].GetOp())
}

func sessionArtifactContentRequest(
	t *testing.T,
	deps *Deps,
	rangeHeader string,
) *httptest.ResponseRecorder {
	t.Helper()
	if deps.ArtifactSnapshots == nil {
		sandbox := deps.SandboxFs.(*artifactContentSandbox)
		deps.ArtifactSnapshots = artifactContentSnapshots{
			state: artifactContentSnapshot(t, sandbox.data),
		}
	}
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/v1/sessions/session-mobile/resources/environments/workspace/artifacts/content/output/demo.mp4?artifact_id=video-1&representation_id=playable&revision=1",
		nil,
	)
	if rangeHeader != "" {
		ctx.Request.Header.Set("Range", rangeHeader)
	}
	ctx.Params = gin.Params{
		{Key: "id", Value: "session-mobile"},
		{Key: "env", Value: "workspace"},
		{Key: "filepath", Value: "/output/demo.mp4"},
	}
	ctx.Set("tenant", &middleware.TenantContext{
		OrganizationID:   21,
		OrganizationSlug: "dev-org",
		UserID:           11,
		UserRole:         "member",
	})
	deps.handleSessionArtifactContent(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}

func artifactContentSnapshot(
	t *testing.T,
	data []byte,
) *workbenchdomain.SessionState {
	t.Helper()
	digest := fmt.Sprintf("sha256:%x", sha256.Sum256(data))
	byteSize := uint64(len(data))
	resourceID := "workspace:output/demo.mp4"
	snapshot := &agentworkbenchv2.SessionSnapshot{
		SessionId: "session-mobile",
		Artifacts: []*agentworkbenchv2.ArtifactDescriptor{{
			ArtifactId: "video-1", Revision: 1,
			Representations: []*agentworkbenchv2.ArtifactRepresentation{{
				RepresentationId: "playable", Revision: 1,
				Status:   agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
				ByteSize: &byteSize, Digest: &digest,
				Transport: &agentworkbenchv2.ArtifactTransport{
					Transport: &agentworkbenchv2.ArtifactTransport_ResourceId{
						ResourceId: resourceID,
					},
				},
			}},
		}},
	}
	projection, err := proto.Marshal(snapshot)
	require.NoError(t, err)
	return &workbenchdomain.SessionState{
		SessionID:  "session-mobile",
		Projection: projection,
	}
}
