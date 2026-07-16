package runner

import (
	"io/fs"
	"os"
	"path/filepath"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (h *RunnerMessageHandler) sandboxFsChanges(workspaceRoot string) (*runnerv1.SandboxFsResultEvent, error) {
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.sandboxFsChangesWorkspace(workspace)
}

func (h *RunnerMessageHandler) sandboxFsChangesWorkspace(
	workspace *sandboxWorkspace,
) (*runnerv1.SandboxFsResultEvent, error) {
	if _, err := workspace.root.Lstat(".git"); os.IsNotExist(err) {
		return sandboxFsStandaloneChanges(workspace)
	} else if err != nil {
		return fsErrResult(err.Error()), nil
	}
	out, err := h.runGitInWorkspace(workspace, "status", "--porcelain")
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	changes := parseGitPorcelain(workspace, out)
	return &runnerv1.SandboxFsResultEvent{
		Changes:       changes,
		WorkspaceRoot: workspace.displayPath(),
	}, nil
}

func sandboxFsStandaloneChanges(
	workspace *sandboxWorkspace,
) (*runnerv1.SandboxFsResultEvent, error) {
	changes := make([]*runnerv1.SandboxFsChange, 0)
	err := fs.WalkDir(workspace.root.FS(), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		changes = append(changes, &runnerv1.SandboxFsChange{
			Path:       filepath.ToSlash(path),
			Name:       entry.Name(),
			Status:     "created",
			Bytes:      info.Size(),
			ModifiedAt: info.ModTime().Unix(),
		})
		return nil
	})
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	return &runnerv1.SandboxFsResultEvent{
		Changes:       changes,
		WorkspaceRoot: workspace.displayPath(),
	}, nil
}

func (h *RunnerMessageHandler) sandboxFsDiff(workspaceRoot, rel string) (*runnerv1.SandboxFsResultEvent, error) {
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.sandboxFsDiffWorkspace(workspace, rel)
}

func (h *RunnerMessageHandler) sandboxFsDiffWorkspace(
	workspace *sandboxWorkspace,
	rel string,
) (*runnerv1.SandboxFsResultEvent, error) {
	_, display, err := resolveWorkspacePath(workspace.displayPath(), rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	status, err := h.gitPathStatus(workspace, display)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	var before, after *string
	switch status {
	case "deleted":
		head, showErr := h.runGitInWorkspace(workspace, "show", "HEAD:"+display)
		if showErr != nil {
			return fsErrResult(showErr.Error()), nil
		}
		b := head
		before = &b
	case "created":
		data, readErr := readSandboxWorkspaceFileIn(workspace, rel)
		if readErr != nil {
			return fsErrResult(readErr.Error()), nil
		}
		s := string(data)
		after = &s
	default:
		head, showErr := h.runGitInWorkspace(workspace, "show", "HEAD:"+display)
		if showErr != nil {
			return fsErrResult(showErr.Error()), nil
		}
		b := head
		before = &b
		data, readErr := readSandboxWorkspaceFileIn(workspace, rel)
		if readErr != nil {
			return fsErrResult(readErr.Error()), nil
		}
		s := string(data)
		after = &s
	}
	res := &runnerv1.SandboxFsResultEvent{WorkspaceRoot: workspace.displayPath()}
	if before != nil {
		res.Before = *before
	}
	if after != nil {
		res.After = *after
	}
	return res, nil
}
