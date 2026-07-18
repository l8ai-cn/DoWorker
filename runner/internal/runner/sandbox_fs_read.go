package runner

import (
	"encoding/base64"
	"io"
	"os"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (h *RunnerMessageHandler) sandboxFsRead(
	workspaceRoot, rel string,
) (*runnerv1.SandboxFsResultEvent, error) {
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.sandboxFsReadWorkspace(workspace, rel)
}

func (h *RunnerMessageHandler) sandboxFsReadWorkspace(
	workspace *sandboxWorkspace,
	rel string,
) (*runnerv1.SandboxFsResultEvent, error) {
	relative, _, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	file, info, err := openSandboxRegularFile(workspace.root, relative)
	if err != nil {
		if os.IsNotExist(err) {
			return fsErrResult("not found"), nil
		}
		return fsErrResult(err.Error()), nil
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxSandboxFsReadBytes+1))
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	truncated := len(data) > maxSandboxFsReadBytes
	if truncated {
		data = data[:maxSandboxFsReadBytes]
	}
	encoding := "utf-8"
	content := string(data)
	if !isUTF8(data) {
		encoding = "base64"
		content = base64.StdEncoding.EncodeToString(data)
	}
	return &runnerv1.SandboxFsResultEvent{
		Content:       content,
		Encoding:      encoding,
		ContentType:   sandboxFsContentType(relative),
		FileBytes:     info.Size(),
		Truncated:     truncated,
		WorkspaceRoot: workspace.displayPath(),
	}, nil
}
