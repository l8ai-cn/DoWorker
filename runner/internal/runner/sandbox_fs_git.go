package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

const gitFsTimeout = 30 * time.Second

func (h *RunnerMessageHandler) sandboxFsChanges(workspaceRoot string) (*runnerv1.SandboxFsResultEvent, error) {
	out, err := h.runGitIn(workspaceRoot, "status", "--porcelain")
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	changes := parseGitPorcelain(workspaceRoot, out)
	return &runnerv1.SandboxFsResultEvent{Changes: changes, WorkspaceRoot: workspaceRoot}, nil
}

func (h *RunnerMessageHandler) sandboxFsDiff(workspaceRoot, rel string) (*runnerv1.SandboxFsResultEvent, error) {
	_, display, err := resolveWorkspacePath(workspaceRoot, rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	status := h.gitPathStatus(workspaceRoot, display)
	var before, after *string
	switch status {
	case "deleted":
		head, _ := h.runGitIn(workspaceRoot, "show", "HEAD:"+display)
		b := head
		before = &b
	case "created":
		abs, _, _ := resolveWorkspacePath(workspaceRoot, rel)
		if data, readErr := os.ReadFile(abs); readErr == nil {
			s := string(data)
			after = &s
		}
	default:
		head, _ := h.runGitIn(workspaceRoot, "show", "HEAD:"+display)
		b := head
		before = &b
		abs, _, _ := resolveWorkspacePath(workspaceRoot, rel)
		if data, readErr := os.ReadFile(abs); readErr == nil {
			s := string(data)
			after = &s
		}
	}
	res := &runnerv1.SandboxFsResultEvent{WorkspaceRoot: workspaceRoot}
	if before != nil {
		res.Before = *before
	}
	if after != nil {
		res.After = *after
	}
	return res, nil
}

func (h *RunnerMessageHandler) runGitIn(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitFsTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (h *RunnerMessageHandler) gitPathStatus(workspaceRoot, rel string) string {
	out, err := h.runGitIn(workspaceRoot, "status", "--porcelain", "--", rel)
	if err != nil || strings.TrimSpace(out) == "" {
		return "modified"
	}
	line := strings.TrimSpace(strings.Split(out, "\n")[0])
	if len(line) < 3 {
		return "modified"
	}
	code := strings.TrimSpace(line[:2])
	switch {
	case strings.Contains(code, "?"):
		return "created"
	case strings.Contains(code, "D"):
		return "deleted"
	default:
		return "modified"
	}
}

func parseGitPorcelain(workspaceRoot, out string) []*runnerv1.SandboxFsChange {
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
		case strings.Contains(code, "?"):
			status = "created"
		case strings.Contains(code, "D"):
			status = "deleted"
		}
		abs := filepath.Join(workspaceRoot, path)
		var size int64
		var mod int64
		if fi, err := os.Stat(abs); err == nil {
			if !fi.IsDir() {
				size = fi.Size()
			}
			mod = fi.ModTime().Unix()
		}
		changes = append(changes, &runnerv1.SandboxFsChange{
			Path:       filepath.ToSlash(path),
			Name:       filepath.Base(path),
			Status:     status,
			Bytes:      size,
			ModifiedAt: mod,
		})
	}
	return changes
}
