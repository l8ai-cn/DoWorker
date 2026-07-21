package runner

import (
	"os"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func (h *RunnerMessageHandler) sandboxFsStatWorkspace(
	workspace *sandboxWorkspace,
	rel string,
) (*runnerv1.SandboxFsResultEvent, error) {
	relative, _, err := resolveSandboxWorkspaceRelativePath(rel)
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
	if !info.Mode().IsRegular() {
		return fsErrResult("not a regular file"), nil
	}
	return &runnerv1.SandboxFsResultEvent{
		ContentType:   sandboxFsContentType(relative),
		FileBytes:     info.Size(),
		WorkspaceRoot: workspace.displayPath(),
	}, nil
}
