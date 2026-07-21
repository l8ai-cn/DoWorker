package sessionapi

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"unicode"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	fileDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/sessionfile"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

type messageAttachment struct {
	FileID string
}

type sessionAttachmentFiles interface {
	GetForSession(context.Context, string, string) (*fileDomain.File, error)
	RunnerDownloadURL(context.Context, *fileDomain.File) (string, error)
}

type sessionAttachmentSandbox interface {
	IsConnected(int64) bool
	Exec(context.Context, int64, *runnerv1.SandboxFsCommand) (*runnerv1.SandboxFsResultEvent, error)
}

func stageMessageAttachments(
	ctx context.Context,
	files sessionAttachmentFiles,
	sandbox sessionAttachmentSandbox,
	pod *podDomain.Pod,
	sessionID string,
	attachments []messageAttachment,
) ([]string, error) {
	if len(attachments) == 0 {
		return nil, nil
	}
	if files == nil || sandbox == nil || pod == nil || pod.RunnerID == 0 || pod.PodKey == "" {
		return nil, fmt.Errorf("attachment delivery unavailable")
	}
	if !sandbox.IsConnected(pod.RunnerID) {
		return nil, fmt.Errorf("runner unavailable")
	}

	paths := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		row, err := files.GetForSession(ctx, sessionID, attachment.FileID)
		if err != nil {
			return nil, fmt.Errorf("resolve attachment %q: %w", attachment.FileID, err)
		}
		if row == nil || row.SessionID != sessionID {
			return nil, fmt.Errorf("attachment %q is unavailable", attachment.FileID)
		}
		downloadURL, err := files.RunnerDownloadURL(ctx, row)
		if err != nil {
			return nil, fmt.Errorf("prepare attachment %q: %w", attachment.FileID, err)
		}
		workspacePath := sessionAttachmentWorkspacePath(row.ID, row.Filename)
		result, err := sandbox.Exec(ctx, pod.RunnerID, &runnerv1.SandboxFsCommand{
			Op: "download", PodKey: pod.PodKey, Path: workspacePath, Payload: downloadURL,
		})
		if err != nil {
			return nil, fmt.Errorf("deliver attachment %q: %w", attachment.FileID, err)
		}
		if result == nil || result.GetError() != "" {
			return nil, fmt.Errorf("deliver attachment %q: %s", attachment.FileID, sandboxResultError(result))
		}
		paths = append(paths, workspacePath)
	}
	return paths, nil
}

func sessionAttachmentWorkspacePath(fileID, filename string) string {
	return path.Join("uploads", attachmentPathPart(fileID)+"-"+attachmentPathPart(filename))
}

func attachmentPathPart(value string) string {
	value = path.Base(strings.ReplaceAll(strings.TrimSpace(value), "\\", "/"))
	var out strings.Builder
	previousDash := false
	for _, r := range strings.ToLower(value) {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '.', r == '_':
			out.WriteRune(r)
			previousDash = false
		case !previousDash:
			out.WriteByte('-')
			previousDash = true
		}
	}
	result := strings.Trim(out.String(), "-.")
	if result == "" {
		return "attachment"
	}
	return result
}

func sandboxResultError(result *runnerv1.SandboxFsResultEvent) string {
	if result == nil || result.GetError() == "" {
		return "runner returned no result"
	}
	return result.GetError()
}

func messageAttachments(data json.RawMessage) []messageAttachment {
	var msg struct {
		Content []messageContentBlock `json:"content"`
	}
	if json.Unmarshal(data, &msg) != nil {
		return nil
	}
	attachments := make([]messageAttachment, 0)
	for _, block := range msg.Content {
		if (block.Type == "input_image" || block.Type == "input_file") && strings.TrimSpace(block.FileID) != "" {
			attachments = append(attachments, messageAttachment{FileID: strings.TrimSpace(block.FileID)})
		}
	}
	return attachments
}

func materializedMessagePrompt(data json.RawMessage, paths []string) string {
	var msg struct {
		Content []messageContentBlock `json:"content"`
	}
	if json.Unmarshal(data, &msg) != nil {
		return ""
	}
	parts := make([]string, 0, len(paths)+len(msg.Content))
	for _, workspacePath := range paths {
		parts = append(parts, "[Attached: "+workspacePath+"]")
	}
	text := make([]string, 0, len(msg.Content))
	for _, block := range msg.Content {
		if block.Type != "text" && block.Type != "input_text" {
			continue
		}
		if value := strings.TrimSpace(block.Text); value != "" {
			text = append(text, value)
		}
	}
	if len(text) > 0 {
		if len(parts) > 0 {
			parts = append(parts, "")
		}
		parts = append(parts, strings.Join(text, "\n"))
	}
	return strings.Join(parts, "\n")
}
