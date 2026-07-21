package runner

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

const gitFsTimeout = 30 * time.Second

func (h *RunnerMessageHandler) runGitInWorkspace(
	workspace *sandboxWorkspace,
	args ...string,
) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitFsTimeout)
	defer cancel()
	return runSandboxCommand(ctx, workspace, "git", args...)
}

func (h *RunnerMessageHandler) gitPathStatus(
	workspace *sandboxWorkspace,
	rel string,
) (string, error) {
	out, err := h.runGitInWorkspace(workspace, "status", "--porcelain", "--", rel)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(out) == "" {
		return "modified", nil
	}
	line := strings.TrimSpace(strings.Split(out, "\n")[0])
	if len(line) < 3 {
		return "modified", nil
	}
	code := strings.TrimSpace(line[:2])
	switch {
	case strings.ContainsAny(code, "?ARC"):
		return "created", nil
	case strings.Contains(code, "D"):
		return "deleted", nil
	default:
		return "modified", nil
	}
}

func parseGitPorcelain(
	workspace *sandboxWorkspace,
	out string,
) []*runnerv1.SandboxFsChange {
	var changes []*runnerv1.SandboxFsChange
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if len(line) < 4 {
			continue
		}
		code := line[:2]
		path := strings.TrimSpace(line[3:])
		status := "modified"
		switch {
		case strings.ContainsAny(code, "?A"):
			status = "created"
		case strings.Contains(code, "D"):
			status = "deleted"
		}
		var size, modified int64
		if info, err := workspace.root.Stat(filepath.FromSlash(path)); err == nil {
			if !info.IsDir() {
				size = info.Size()
			}
			modified = info.ModTime().Unix()
		}
		changes = append(changes, &runnerv1.SandboxFsChange{
			Path: filepath.ToSlash(path), Name: filepath.Base(path), Status: status,
			Bytes: size, ModifiedAt: modified,
		})
	}
	return changes
}
