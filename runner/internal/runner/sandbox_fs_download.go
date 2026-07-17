package runner

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

var sandboxFsDownloadClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func (h *RunnerMessageHandler) sandboxFsDownload(
	workspaceRoot, rel, downloadURL string,
) (*runnerv1.SandboxFsResultEvent, error) {
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.sandboxFsDownloadWorkspace(context.Background(), workspace, rel, downloadURL)
}

func (h *RunnerMessageHandler) sandboxFsDownloadWorkspace(
	ctx context.Context,
	workspace *sandboxWorkspace,
	rel, downloadURL string,
) (*runnerv1.SandboxFsResultEvent, error) {
	relative, _, err := resolveSandboxWorkspaceRelativePath(rel)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	parsed, err := url.ParseRequestURI(downloadURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fsErrResult("invalid download URL"), nil
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	resp, err := sandboxFsDownloadClient.Do(request)
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
	if err := writeDownloadedFile(workspace.root, relative, resp.Body); err != nil {
		return fsErrResult(err.Error()), nil
	}
	return &runnerv1.SandboxFsResultEvent{WorkspaceRoot: workspace.displayPath()}, nil
}

func writeDownloadedFile(root *os.Root, destination string, body io.Reader) error {
	if err := root.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("prepare download path: %w", err)
	}
	tempName, err := sandboxDownloadTempName(destination)
	if err != nil {
		return err
	}
	temp, err := root.OpenFile(tempName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create download file: %w", err)
	}
	defer root.Remove(tempName)

	written, copyErr := io.Copy(temp, io.LimitReader(body, maxSandboxFsDownloadBytes+1))
	if closeErr := temp.Close(); copyErr != nil {
		return fmt.Errorf("write download: %w", copyErr)
	} else if closeErr != nil {
		return fmt.Errorf("close download: %w", closeErr)
	}
	if written > maxSandboxFsDownloadBytes {
		return fmt.Errorf("download exceeds maximum file size")
	}
	if err := root.Rename(tempName, destination); err != nil {
		return fmt.Errorf("commit download: %w", err)
	}
	return nil
}

func sandboxDownloadTempName(destination string) (string, error) {
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", fmt.Errorf("create download file name: %w", err)
	}
	name := "." + filepath.Base(destination) + ".download-" + hex.EncodeToString(suffix[:])
	return filepath.Join(filepath.Dir(destination), name), nil
}
