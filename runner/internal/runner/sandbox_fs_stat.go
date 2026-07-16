package runner

import (
	"mime"
	"os"
	"path/filepath"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (h *RunnerMessageHandler) sandboxFsStat(
	workspaceRoot string,
	rel string,
) (*runnerv1.SandboxFsResultEvent, error) {
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
	if !info.Mode().IsRegular() {
		return fsErrResult("not a regular file"), nil
	}
	return &runnerv1.SandboxFsResultEvent{
		ContentType:   mime.TypeByExtension(filepath.Ext(relative)),
		FileBytes:     info.Size(),
		WorkspaceRoot: workspaceRoot,
	}, nil
}
