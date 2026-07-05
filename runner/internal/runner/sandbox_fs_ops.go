package runner

import (
	"encoding/base64"
	"fmt"
	"io/fs"
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
	abs, display, err := resolveWorkspacePath(workspaceRoot, rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return fsErrResult("not found"), nil
		}
		return fsErrResult(err.Error()), nil
	}
	if !info.IsDir() {
		return h.sandboxFsRead(workspaceRoot, rel)
	}
	entries, err := os.ReadDir(abs)
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
		childRel := display
		if childRel == "." {
			childRel = e.Name()
		} else {
			childRel = filepath.ToSlash(filepath.Join(childRel, e.Name()))
		}
		out = append(out, fsEntryFromInfo(childRel, fi))
	}
	return &runnerv1.SandboxFsResultEvent{Entries: out, WorkspaceRoot: workspaceRoot}, nil
}

func (h *RunnerMessageHandler) sandboxFsRead(workspaceRoot, rel string) (*runnerv1.SandboxFsResultEvent, error) {
	abs, _, err := resolveWorkspacePath(workspaceRoot, rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return fsErrResult("not found"), nil
		}
		return fsErrResult(err.Error()), nil
	}
	if info.IsDir() {
		return fsErrResult("is a directory"), nil
	}
	data, err := os.ReadFile(abs)
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
	ct := mime.TypeByExtension(filepath.Ext(abs))
	return &runnerv1.SandboxFsResultEvent{
		Content:       content,
		Encoding:      encoding,
		ContentType:   ct,
		FileBytes:     info.Size(),
		Truncated:     truncated,
		WorkspaceRoot: workspaceRoot,
	}, nil
}

func (h *RunnerMessageHandler) sandboxFsWrite(workspaceRoot, rel, content string) (*runnerv1.SandboxFsResultEvent, error) {
	abs, _, err := resolveWorkspacePath(workspaceRoot, rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return fsErrResult(err.Error()), nil
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return fsErrResult(err.Error()), nil
	}
	return &runnerv1.SandboxFsResultEvent{WorkspaceRoot: workspaceRoot}, nil
}

func (h *RunnerMessageHandler) sandboxFsMkdir(workspaceRoot, rel string) (*runnerv1.SandboxFsResultEvent, error) {
	abs, _, err := resolveWorkspacePath(workspaceRoot, rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return fsErrResult(err.Error()), nil
	}
	return &runnerv1.SandboxFsResultEvent{WorkspaceRoot: workspaceRoot}, nil
}

func fsErrResult(msg string) *runnerv1.SandboxFsResultEvent {
	return &runnerv1.SandboxFsResultEvent{Error: msg}
}

func isUTF8(b []byte) bool {
	return len(b) == 0 || strings.ToValidUTF8(string(b), "") == string(b)
}

func hostDirEntries(root string) ([]*runnerv1.SandboxFsEntry, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	out := make([]*runnerv1.SandboxFsEntry, 0, len(entries))
	for _, e := range entries {
		fi, err := e.Info()
		if err != nil {
			continue
		}
		if fi.Mode()&fs.ModeSymlink != 0 {
			continue
		}
		abs := filepath.Join(root, e.Name())
		out = append(out, &runnerv1.SandboxFsEntry{
			Path:       abs,
			Name:       e.Name(),
			Type:       entryType(fi),
			Bytes:      fileSize(fi),
			ModifiedAt: fi.ModTime().Unix(),
		})
	}
	return out, nil
}

func entryType(fi os.FileInfo) string {
	if fi.IsDir() {
		return "directory"
	}
	return "file"
}

func fileSize(fi os.FileInfo) int64 {
	if fi.IsDir() {
		return 0
	}
	return fi.Size()
}

func listHostWorkspaceEntries(workspaceRoot, path string) ([]*runnerv1.SandboxFsEntry, error) {
	root := workspaceRoot
	abs := root
	if path != "" {
		abs = filepath.Join(root, strings.TrimPrefix(path, "/"))
	}
	abs, err := filepath.Abs(abs)
	if err != nil {
		return nil, err
	}
	if abs != root && !strings.HasPrefix(abs, root+string(filepath.Separator)) {
		return nil, fmt.Errorf("path escapes workspace root")
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory")
	}
	return hostDirEntries(abs)
}
