package runner

import (
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

const maxSandboxFsReadChunkBytes int64 = 4 << 20

func (h *RunnerMessageHandler) sandboxFsStat(
	workspaceRoot string,
	rel string,
) (*runnerv1.SandboxFsResultEvent, error) {
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.sandboxFsStatWorkspace(workspace, rel)
}

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

func (h *RunnerMessageHandler) sandboxFsReadBytes(
	workspaceRoot string,
	rel string,
	offset int64,
	length int64,
) (*runnerv1.SandboxFsResultEvent, error) {
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.sandboxFsReadBytesWorkspace(workspace, rel, offset, length)
}

func (h *RunnerMessageHandler) sandboxFsReadBytesWorkspace(
	workspace *sandboxWorkspace,
	rel string,
	offset int64,
	length int64,
) (*runnerv1.SandboxFsResultEvent, error) {
	if offset < 0 {
		return fsErrResult("offset must not be negative"), nil
	}
	if length <= 0 || length > maxSandboxFsReadChunkBytes {
		return fsErrResult(fmt.Sprintf(
			"length must be between 1 and %d bytes",
			maxSandboxFsReadChunkBytes,
		)), nil
	}
	relative, _, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	file, info, err := openSandboxRegularFile(workspace.root, relative)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer file.Close()
	if offset > info.Size() {
		return fsErrResult("offset exceeds file size"), nil
	}
	data := make([]byte, min(length, info.Size()-offset))
	read, err := file.ReadAt(data, offset)
	if err != nil && err != io.EOF {
		return fsErrResult(err.Error()), nil
	}
	data = data[:read]
	return &runnerv1.SandboxFsResultEvent{
		ContentBytes:  data,
		ContentOffset: offset,
		ContentType:   sandboxFsContentType(relative),
		FileBytes:     info.Size(),
		Eof:           offset+int64(read) >= info.Size(),
		WorkspaceRoot: workspace.displayPath(),
	}, nil
}

func sandboxFsContentType(path string) string {
	if strings.EqualFold(filepath.Ext(path), ".pptx") {
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	}
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		return "application/octet-stream"
	}
	return contentType
}
