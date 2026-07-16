package runner

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

const maxSandboxFsUploadBytes int64 = 128 << 20

var sandboxFsUploadClient = &http.Client{
	Timeout: 2 * time.Minute,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func (h *RunnerMessageHandler) sandboxFsUpload(
	workspaceRoot string,
	rel string,
	uploadURL string,
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
	parsed, err := url.ParseRequestURI(uploadURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fsErrResult("invalid upload URL"), nil
	}
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
	if info.Size() > maxSandboxFsUploadBytes {
		return fsErrResult("upload exceeds maximum file size"), nil
	}
	file, err := root.Open(relative)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer file.Close()
	contentType := mime.TypeByExtension(filepath.Ext(relative))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	request, err := http.NewRequest(http.MethodPut, parsed.String(), file)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	request.ContentLength = info.Size()
	request.Header.Set("Content-Type", contentType)
	response, err := sandboxFsUploadClient.Do(request)
	if err != nil {
		return fsErrResult(fmt.Sprintf("upload failed: %v", err)), nil
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fsErrResult(fmt.Sprintf("upload failed: HTTP %d", response.StatusCode)), nil
	}
	return &runnerv1.SandboxFsResultEvent{
		ContentType:   contentType,
		FileBytes:     info.Size(),
		WorkspaceRoot: workspaceRoot,
	}, nil
}
