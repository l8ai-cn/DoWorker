package sessionfile

import (
	"context"
	"fmt"
	"path"
	"strings"
	"unicode"

	poddomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

type AttachmentSandbox interface {
	IsConnected(int64) bool
	Exec(context.Context, int64, *runnerv1.SandboxFsCommand) (*runnerv1.SandboxFsResultEvent, error)
}

func (s *Service) Stage(
	ctx context.Context,
	sandbox AttachmentSandbox,
	pod *poddomain.Pod,
	sessionID string,
	fileIDs []string,
) ([]string, error) {
	if len(fileIDs) == 0 {
		return nil, nil
	}
	if s == nil || sandbox == nil || pod == nil ||
		pod.RunnerID == 0 || pod.PodKey == "" {
		return nil, fmt.Errorf("attachment delivery unavailable")
	}
	if !sandbox.IsConnected(pod.RunnerID) {
		return nil, fmt.Errorf("runner unavailable")
	}
	paths := make([]string, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		row, err := s.GetForSession(ctx, sessionID, fileID)
		if err != nil {
			return nil, fmt.Errorf("resolve attachment %q: %w", fileID, err)
		}
		downloadURL, err := s.RunnerDownloadURL(ctx, row)
		if err != nil {
			return nil, fmt.Errorf("prepare attachment %q: %w", fileID, err)
		}
		workspacePath := attachmentWorkspacePath(row.ID, row.Filename)
		result, err := sandbox.Exec(ctx, pod.RunnerID, &runnerv1.SandboxFsCommand{
			Op: "download", PodKey: pod.PodKey,
			Path: workspacePath, Payload: downloadURL,
		})
		if err != nil {
			return nil, fmt.Errorf("deliver attachment %q: %w", fileID, err)
		}
		if result == nil || result.GetError() != "" {
			return nil, fmt.Errorf("deliver attachment %q: %s", fileID, resultError(result))
		}
		paths = append(paths, workspacePath)
	}
	return paths, nil
}

func attachmentWorkspacePath(fileID, filename string) string {
	return path.Join("uploads", safePathPart(fileID)+"-"+safePathPart(filename))
}

func safePathPart(value string) string {
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

func resultError(result *runnerv1.SandboxFsResultEvent) string {
	if result == nil || result.GetError() == "" {
		return "runner returned no result"
	}
	return result.GetError()
}
