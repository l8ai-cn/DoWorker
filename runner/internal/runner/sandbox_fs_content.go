package runner

import (
	"fmt"
	"io"
)

func readSandboxWorkspaceFileIn(
	workspace *sandboxWorkspace,
	rel string,
) ([]byte, error) {
	relative, _, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return nil, err
	}
	file, info, err := openSandboxRegularFile(workspace.root, relative)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if info.Size() > maxSandboxFsReadBytes {
		return nil, sandboxFsContentLimitError()
	}
	data, err := io.ReadAll(io.LimitReader(file, maxSandboxFsReadBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxSandboxFsReadBytes {
		return nil, sandboxFsContentLimitError()
	}
	return data, nil
}

func sandboxFsContentLimitError() error {
	return fmt.Errorf(
		"sandbox filesystem content exceeds %d byte limit",
		maxSandboxFsReadBytes,
	)
}
