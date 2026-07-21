package runner

import (
	"os"
	"path/filepath"
	"strings"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func fsEntryFromInfo(relDisplay string, info os.FileInfo) *runnerv1.SandboxFsEntry {
	name := info.Name()
	typ := "file"
	var size int64
	if info.IsDir() {
		typ = "directory"
	} else {
		size = info.Size()
	}
	path := relDisplay
	if path == "." {
		path = name
	} else if relDisplay != "" {
		path = filepath.ToSlash(filepath.Join(relDisplay, name))
	}
	return &runnerv1.SandboxFsEntry{
		Path:       path,
		Name:       name,
		Type:       typ,
		Bytes:      size,
		ModifiedAt: info.ModTime().Unix(),
	}
}

func (h *RunnerMessageHandler) sandboxFsList(workspaceRoot, rel string) (*runnerv1.SandboxFsResultEvent, error) {
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.sandboxFsListWorkspace(workspace, rel)
}

func (h *RunnerMessageHandler) sandboxFsListWorkspace(
	workspace *sandboxWorkspace,
	rel string,
) (*runnerv1.SandboxFsResultEvent, error) {
	relative, display, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	info, err := workspace.root.Stat(relative)
	if err != nil {
		if os.IsNotExist(err) {
			return fsErrResult("not found"), nil
		}
		return fsErrResult(err.Error()), nil
	}
	if !info.IsDir() {
		return h.sandboxFsReadWorkspace(workspace, rel)
	}
	dir, err := workspace.root.Open(relative)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer dir.Close()
	entries, err := dir.ReadDir(-1)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	out := make([]*runnerv1.SandboxFsEntry, 0, len(entries))
	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}
		fi, statErr := e.Info()
		if statErr != nil {
			continue
		}
		out = append(out, fsEntryFromInfo(display, fi))
	}
	return &runnerv1.SandboxFsResultEvent{
		Entries:       out,
		WorkspaceRoot: workspace.displayPath(),
	}, nil
}

func (h *RunnerMessageHandler) sandboxFsWrite(workspaceRoot, rel, content string) (*runnerv1.SandboxFsResultEvent, error) {
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.sandboxFsWriteWorkspace(workspace, rel, content)
}

func (h *RunnerMessageHandler) sandboxFsWriteWorkspace(
	workspace *sandboxWorkspace,
	rel, content string,
) (*runnerv1.SandboxFsResultEvent, error) {
	relative, _, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	if err := workspace.root.MkdirAll(filepath.Dir(relative), 0o755); err != nil {
		return fsErrResult(err.Error()), nil
	}
	if err := workspace.root.WriteFile(relative, []byte(content), 0o644); err != nil {
		return fsErrResult(err.Error()), nil
	}
	return &runnerv1.SandboxFsResultEvent{WorkspaceRoot: workspace.displayPath()}, nil
}

func (h *RunnerMessageHandler) sandboxFsMkdir(workspaceRoot, rel string) (*runnerv1.SandboxFsResultEvent, error) {
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.sandboxFsMkdirWorkspace(workspace, rel)
}

func (h *RunnerMessageHandler) sandboxFsMkdirWorkspace(
	workspace *sandboxWorkspace,
	rel string,
) (*runnerv1.SandboxFsResultEvent, error) {
	relative, _, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	if err := workspace.root.MkdirAll(relative, 0o755); err != nil {
		return fsErrResult(err.Error()), nil
	}
	return &runnerv1.SandboxFsResultEvent{WorkspaceRoot: workspace.displayPath()}, nil
}

func fsErrResult(msg string) *runnerv1.SandboxFsResultEvent {
	return &runnerv1.SandboxFsResultEvent{Error: msg}
}

func isUTF8(b []byte) bool {
	return len(b) == 0 || strings.ToValidUTF8(string(b), "") == string(b)
}
