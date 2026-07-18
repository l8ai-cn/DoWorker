package runner

import (
	"encoding/base64"
	"mime"
	"os"
	"path/filepath"
	"strings"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
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
	relative, display, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	root, err := openSandboxWorkspaceRoot(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer root.Close()
	info, err := root.Stat(relative)
	if err != nil {
		if os.IsNotExist(err) {
			return fsErrResult("not found"), nil
		}
		return fsErrResult(err.Error()), nil
	}
	if !info.IsDir() {
		return h.sandboxFsReadWorkspace(workspace, rel)
	}
	dir, err := root.Open(relative)
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
	return &runnerv1.SandboxFsResultEvent{Entries: out, WorkspaceRoot: workspaceRoot}, nil
}

func (h *RunnerMessageHandler) sandboxFsRead(workspaceRoot, rel string) (*runnerv1.SandboxFsResultEvent, error) {
	relative, _, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	root, err := openSandboxWorkspaceRoot(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer root.Close()
	info, err := root.Stat(relative)
	if err != nil {
		if os.IsNotExist(err) {
			return fsErrResult("not found"), nil
		}
		return fsErrResult(err.Error()), nil
	}
	if info.IsDir() {
		return fsErrResult("is a directory"), nil
	}
	data, err := root.ReadFile(relative)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	truncated := false
	if len(data) > maxSandboxFsReadBytes {
		data = data[:maxSandboxFsReadBytes]
		truncated = true
	}
	encoding := "utf-8"
	content := string(data)
	if !isUTF8(data) {
		encoding = "base64"
		content = base64.StdEncoding.EncodeToString(data)
	}
	ct := mime.TypeByExtension(filepath.Ext(relative))
	return &runnerv1.SandboxFsResultEvent{
		Entries:       out,
		WorkspaceRoot: workspace.displayPath(),
	}, nil
}

func (h *RunnerMessageHandler) sandboxFsWrite(workspaceRoot, rel, content string) (*runnerv1.SandboxFsResultEvent, error) {
	relative, _, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	root, err := openSandboxWorkspaceRoot(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer root.Close()
	if err := root.MkdirAll(filepath.Dir(relative), 0o755); err != nil {
		return fsErrResult(err.Error()), nil
	}
	if err := root.WriteFile(relative, []byte(content), 0o644); err != nil {
		return fsErrResult(err.Error()), nil
	}
	if err := workspace.root.WriteFile(relative, []byte(content), 0o644); err != nil {
		return fsErrResult(err.Error()), nil
	}
	return &runnerv1.SandboxFsResultEvent{WorkspaceRoot: workspace.displayPath()}, nil
}

func (h *RunnerMessageHandler) sandboxFsMkdir(workspaceRoot, rel string) (*runnerv1.SandboxFsResultEvent, error) {
	relative, _, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	root, err := openSandboxWorkspaceRoot(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer root.Close()
	if err := root.MkdirAll(relative, 0o755); err != nil {
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
