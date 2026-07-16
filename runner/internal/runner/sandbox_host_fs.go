package runner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func hostDirEntries(root string) ([]*runnerv1.SandboxFsEntry, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	out := make([]*runnerv1.SandboxFsEntry, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || info.Mode()&fs.ModeSymlink != 0 {
			continue
		}
		abs := filepath.Join(root, entry.Name())
		out = append(out, &runnerv1.SandboxFsEntry{
			Path:       abs,
			Name:       entry.Name(),
			Type:       hostEntryType(info),
			Bytes:      hostEntrySize(info),
			ModifiedAt: info.ModTime().Unix(),
		})
	}
	return out, nil
}

func hostEntryType(info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}
	return "file"
}

func hostEntrySize(info os.FileInfo) int64 {
	if info.IsDir() {
		return 0
	}
	return info.Size()
}

func listHostWorkspaceEntries(workspaceRoot, path string) ([]*runnerv1.SandboxFsEntry, error) {
	abs := workspaceRoot
	if path != "" {
		abs = filepath.Join(workspaceRoot, strings.TrimPrefix(path, "/"))
	}
	abs, err := filepath.Abs(abs)
	if err != nil {
		return nil, err
	}
	if abs != workspaceRoot &&
		!strings.HasPrefix(abs, workspaceRoot+string(filepath.Separator)) {
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
