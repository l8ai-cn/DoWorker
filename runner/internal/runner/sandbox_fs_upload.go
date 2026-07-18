package runner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.sandboxFsUploadWorkspace(context.Background(), workspace, rel, uploadURL)
}

func (h *RunnerMessageHandler) sandboxFsUploadWorkspace(
	ctx context.Context,
	workspace *sandboxWorkspace,
	rel string,
	uploadURL string,
) (*runnerv1.SandboxFsResultEvent, error) {
	relative, _, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	parsed, err := url.ParseRequestURI(uploadURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fsErrResult("invalid upload URL"), nil
	}
	file, info, err := openSandboxRegularFile(workspace.root, relative)
	if err != nil {
		if os.IsNotExist(err) {
			return fsErrResult("not found"), nil
		}
		return fsErrResult(err.Error()), nil
	}
	defer file.Close()
	if info.Size() > maxSandboxFsUploadBytes {
		return fsErrResult("upload exceeds maximum file size"), nil
	}
	contentType := sandboxFsContentType(relative)
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, parsed.String(), file)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	hasher := sha256.New()
	request.Body = io.NopCloser(io.TeeReader(file, hasher))
	request.ContentLength = info.Size()
	request.Header.Set("Content-Type", contentType)
	response, err := sandboxFsUploadClient.Do(request)
	if err != nil {
		return fsErrResult(fmt.Sprintf("upload failed: %v", err)), nil
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		detail := strings.TrimSpace(string(responseBody))
		if detail == "" {
			detail = http.StatusText(response.StatusCode)
		}
		return fsErrResult(fmt.Sprintf(
			"upload failed: HTTP %d: %s",
			response.StatusCode,
			detail,
		)), nil
	}
	return &runnerv1.SandboxFsResultEvent{
		ContentType:   contentType,
		ContentDigest: "sha256:" + hex.EncodeToString(hasher.Sum(nil)),
		FileBytes:     info.Size(),
		WorkspaceRoot: workspace.displayPath(),
	}, nil
}
