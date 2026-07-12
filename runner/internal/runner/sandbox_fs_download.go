package runner

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

const maxSandboxFsDownloadBytes = 128 << 20

var sandboxFsDownloadClient = &http.Client{Timeout: 30 * time.Second}

func (h *RunnerMessageHandler) sandboxFsDownload(
	workspaceRoot, rel, downloadURL string,
) (*runnerv1.SandboxFsResultEvent, error) {
	abs, _, err := resolveWorkspacePath(workspaceRoot, rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	parsed, err := url.ParseRequestURI(downloadURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fsErrResult("invalid download URL"), nil
	}
	resp, err := sandboxFsDownloadClient.Get(parsed.String())
	if err != nil {
		return fsErrResult(fmt.Sprintf("download failed: %v", err)), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fsErrResult(fmt.Sprintf("download failed: HTTP %d", resp.StatusCode)), nil
	}
	if resp.ContentLength > maxSandboxFsDownloadBytes {
		return fsErrResult("download exceeds maximum file size"), nil
	}
	if err := writeDownloadedFile(abs, resp.Body); err != nil {
		return fsErrResult(err.Error()), nil
	}
	return &runnerv1.SandboxFsResultEvent{WorkspaceRoot: workspaceRoot}, nil
}

func writeDownloadedFile(destination string, body io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("prepare download path: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(destination), ".download-*")
	if err != nil {
		return fmt.Errorf("create download file: %w", err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	written, copyErr := io.Copy(temp, io.LimitReader(body, maxSandboxFsDownloadBytes+1))
	if closeErr := temp.Close(); copyErr != nil {
		return fmt.Errorf("write download: %w", copyErr)
	} else if closeErr != nil {
		return fmt.Errorf("close download: %w", closeErr)
	}
	if written > maxSandboxFsDownloadBytes {
		return fmt.Errorf("download exceeds maximum file size")
	}
	if err := os.Chmod(tempName, 0o644); err != nil {
		return fmt.Errorf("set download permissions: %w", err)
	}
	if err := os.Rename(tempName, destination); err != nil {
		return fmt.Errorf("commit download: %w", err)
	}
	return nil
}
