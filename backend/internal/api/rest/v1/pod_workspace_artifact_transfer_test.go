package v1

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	fileservice "github.com/l8ai-cn/agentcloud/backend/internal/service/file"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type workspaceTransferSandbox struct {
	commands []*runnerv1.SandboxFsCommand
	results  []*runnerv1.SandboxFsResultEvent
	exec     func(context.Context, *runnerv1.SandboxFsCommand) (*runnerv1.SandboxFsResultEvent, error)
}

func (*workspaceTransferSandbox) IsConnected(int64) bool { return true }

func (s *workspaceTransferSandbox) Exec(
	ctx context.Context,
	_ int64,
	command *runnerv1.SandboxFsCommand,
) (*runnerv1.SandboxFsResultEvent, error) {
	s.commands = append(s.commands, command)
	if s.exec != nil {
		return s.exec(ctx, command)
	}
	result := s.results[0]
	s.results = s.results[1:]
	return result, nil
}

type workspaceArtifactTransferStub struct {
	data     []byte
	prepared *fileservice.WorkspaceArtifactTransfer
	deleted  bool
}

func (s *workspaceArtifactTransferStub) PrepareWorkspaceArtifactTransfer(
	_ context.Context,
	_ int64,
	_ string,
	_ string,
	_ int64,
) (*fileservice.WorkspaceArtifactTransfer, error) {
	s.prepared = &fileservice.WorkspaceArtifactTransfer{
		Key: "temporary-video", PutURL: "https://storage/upload",
		ContentType: "video/mp4", Size: int64(len(s.data)),
	}
	return s.prepared, nil
}

func (s *workspaceArtifactTransferStub) OpenWorkspaceArtifact(
	context.Context,
	*fileservice.WorkspaceArtifactTransfer,
) (io.ReadCloser, int64, error) {
	return io.NopCloser(bytes.NewReader(s.data)), int64(len(s.data)), nil
}

func (s *workspaceArtifactTransferStub) DeleteWorkspaceArtifact(
	context.Context,
	*fileservice.WorkspaceArtifactTransfer,
) error {
	s.deleted = true
	return nil
}

func TestTransferWorkspaceArtifactStreamsCompleteVideo(t *testing.T) {
	video := bytes.Repeat([]byte{0x4d}, 2<<20)
	sandbox := &workspaceTransferSandbox{results: []*runnerv1.SandboxFsResultEvent{
		{FileBytes: int64(len(video)), ContentType: "video/mp4"},
		{FileBytes: int64(len(video)), ContentType: "video/mp4"},
	}}
	transfer := &workspaceArtifactTransferStub{data: video}
	handler := podWorkspaceHandler(nil)
	handler.sandboxFs = sandbox
	handler.workspaceArtifacts = transfer
	recorder, ctx := podWorkspaceRequest(
		http.MethodGet,
		"/pods/worker-1/resources/workspace/artifacts/output/result.mp4",
		"worker-1",
	)
	ctx.Params = append(ctx.Params, ginParam("filepath", "/output/result.mp4"))

	handler.TransferWorkspaceArtifact(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, video, recorder.Body.Bytes())
	assert.Equal(t, "video/mp4", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	assert.True(t, transfer.deleted)
	require.Len(t, sandbox.commands, 2)
	assert.Equal(t, "stat", sandbox.commands[0].GetOp())
	assert.Equal(t, "upload", sandbox.commands[1].GetOp())
	assert.Equal(t, transfer.prepared.PutURL, sandbox.commands[1].GetPayload())
}

func TestTransferWorkspaceArtifactRejectsConcurrentTransferForPod(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	sandbox := &workspaceTransferSandbox{
		exec: func(context.Context, *runnerv1.SandboxFsCommand) (*runnerv1.SandboxFsResultEvent, error) {
			once.Do(func() {
				close(started)
				<-release
			})
			return &runnerv1.SandboxFsResultEvent{FileBytes: 5, ContentType: "video/mp4"}, nil
		},
	}
	handler := podWorkspaceHandler(nil)
	handler.sandboxFs = sandbox
	handler.workspaceArtifacts = &workspaceArtifactTransferStub{data: []byte("video")}
	firstRecorder, firstCtx := workspaceArtifactRequest()
	done := make(chan struct{})
	go func() {
		handler.TransferWorkspaceArtifact(firstCtx)
		close(done)
	}()
	<-started

	secondRecorder, secondCtx := workspaceArtifactRequest()
	handler.TransferWorkspaceArtifact(secondCtx)

	assert.Equal(t, http.StatusTooManyRequests, secondRecorder.Code)
	close(release)
	<-done
	assert.Equal(t, http.StatusOK, firstRecorder.Code)
}

func TestTransferWorkspaceArtifactMapsUploadFailureToBadGateway(t *testing.T) {
	sandbox := &workspaceTransferSandbox{results: []*runnerv1.SandboxFsResultEvent{
		{FileBytes: 5, ContentType: "video/mp4"},
		{Error: "upload failed: HTTP 500"},
	}}
	handler := podWorkspaceHandler(nil)
	handler.sandboxFs = sandbox
	handler.workspaceArtifacts = &workspaceArtifactTransferStub{data: []byte("video")}
	recorder, ctx := workspaceArtifactRequest()

	handler.TransferWorkspaceArtifact(ctx)

	assert.Equal(t, http.StatusBadGateway, recorder.Code)
}

func TestTransferWorkspaceArtifactMapsTimeoutToGatewayTimeout(t *testing.T) {
	sandbox := &workspaceTransferSandbox{
		exec: func(context.Context, *runnerv1.SandboxFsCommand) (*runnerv1.SandboxFsResultEvent, error) {
			return nil, context.DeadlineExceeded
		},
	}
	handler := podWorkspaceHandler(nil)
	handler.sandboxFs = sandbox
	handler.workspaceArtifacts = &workspaceArtifactTransferStub{data: []byte("video")}
	recorder, ctx := workspaceArtifactRequest()

	handler.TransferWorkspaceArtifact(ctx)

	assert.Equal(t, http.StatusGatewayTimeout, recorder.Code)
}

func TestTransferWorkspaceArtifactMapsSizeLimitToPayloadTooLarge(t *testing.T) {
	sandbox := &workspaceTransferSandbox{results: []*runnerv1.SandboxFsResultEvent{
		{FileBytes: 5, ContentType: "video/mp4"},
		{Error: "upload exceeds maximum file size"},
	}}
	handler := podWorkspaceHandler(nil)
	handler.sandboxFs = sandbox
	handler.workspaceArtifacts = &workspaceArtifactTransferStub{data: []byte("video")}
	recorder, ctx := workspaceArtifactRequest()

	handler.TransferWorkspaceArtifact(ctx)

	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
}

func TestTransferWorkspaceArtifactReturnsUnavailableWhenPrepareFails(t *testing.T) {
	sandbox := &workspaceTransferSandbox{results: []*runnerv1.SandboxFsResultEvent{
		{FileBytes: 5, ContentType: "video/mp4"},
	}}
	handler := podWorkspaceHandler(nil)
	handler.sandboxFs = sandbox
	handler.workspaceArtifacts = &workspaceArtifactTransferErrorStub{err: errors.New("storage unavailable")}
	recorder, ctx := workspaceArtifactRequest()

	handler.TransferWorkspaceArtifact(ctx)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

type workspaceArtifactTransferErrorStub struct {
	workspaceArtifactTransferStub
	err error
}

func (s *workspaceArtifactTransferErrorStub) PrepareWorkspaceArtifactTransfer(
	context.Context,
	int64,
	string,
	string,
	int64,
) (*fileservice.WorkspaceArtifactTransfer, error) {
	return nil, s.err
}

func workspaceArtifactRequest() (*httptest.ResponseRecorder, *gin.Context) {
	recorder, ctx := podWorkspaceRequest(
		http.MethodGet,
		"/pods/worker-1/resources/workspace/artifacts/output/result.mp4",
		"worker-1",
	)
	ctx.Params = append(ctx.Params, ginParam("filepath", "/output/result.mp4"))
	return recorder, ctx
}

func ginParam(key, value string) gin.Param {
	return gin.Param{Key: key, Value: value}
}
