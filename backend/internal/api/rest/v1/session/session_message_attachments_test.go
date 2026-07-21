package sessionapi

import (
	"context"
	"encoding/json"
	"testing"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	fileDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/sessionfile"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/require"
)

func TestStageMessageAttachmentsDownloadsOwnedFileIntoPodWorkspace(t *testing.T) {
	files := &attachmentFilesStub{
		row:         &fileDomain.File{ID: "file_123", SessionID: "session-1", Filename: "release notes.pdf"},
		downloadURL: "https://storage.example.com/signed-download",
	}
	sandbox := &attachmentSandboxStub{}

	paths, err := stageMessageAttachments(
		context.Background(),
		files,
		sandbox,
		&podDomain.Pod{RunnerID: 7, PodKey: "pod-1"},
		"session-1",
		[]messageAttachment{{FileID: "file_123"}},
	)

	require.NoError(t, err)
	require.Equal(t, []string{"uploads/file_123-release-notes.pdf"}, paths)
	require.Equal(t, &runnerv1.SandboxFsCommand{
		Op: "download", PodKey: "pod-1", Path: "uploads/file_123-release-notes.pdf",
		Payload: "https://storage.example.com/signed-download",
	}, sandbox.command)
}

func TestMaterializedMessagePromptUsesDeliveredWorkspacePaths(t *testing.T) {
	data := json.RawMessage(`{
		"content": [
			{"type":"input_image","file_id":"file_123","filename":"client-name.png"},
			{"type":"input_text","text":"Please inspect this."}
		]
	}`)

	prompt := materializedMessagePrompt(data, []string{"uploads/file_123-server-name.png"})

	require.Equal(t, "[Attached: uploads/file_123-server-name.png]\n\nPlease inspect this.", prompt)
}

type attachmentFilesStub struct {
	row         *fileDomain.File
	downloadURL string
}

func (s *attachmentFilesStub) GetForSession(_ context.Context, _, _ string) (*fileDomain.File, error) {
	return s.row, nil
}

func (s *attachmentFilesStub) RunnerDownloadURL(_ context.Context, _ *fileDomain.File) (string, error) {
	return s.downloadURL, nil
}

type attachmentSandboxStub struct {
	command *runnerv1.SandboxFsCommand
}

func (s *attachmentSandboxStub) IsConnected(runnerID int64) bool { return runnerID == 7 }

func (s *attachmentSandboxStub) Exec(
	_ context.Context,
	_ int64,
	command *runnerv1.SandboxFsCommand,
) (*runnerv1.SandboxFsResultEvent, error) {
	s.command = command
	return &runnerv1.SandboxFsResultEvent{}, nil
}
