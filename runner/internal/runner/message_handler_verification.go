package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/envpath"
)

const (
	defaultVerificationTimeout = 300 * time.Second
	maxVerificationTimeout     = 900 * time.Second
	maxVerificationOutputBytes = 64 << 10
)

func (h *RunnerMessageHandler) OnRunVerification(cmd *runnerv1.RunVerificationCommand) error {
	result, err := h.verificationResult(cmd)
	if err != nil {
		return err
	}
	if err := h.conn.SendVerificationResult(result); err != nil {
		return fmt.Errorf("send verification result: %w", err)
	}
	return nil
}

func (h *RunnerMessageHandler) runVerification(cmd *runnerv1.RunVerificationCommand) *runnerv1.VerificationResultEvent {
	result := &runnerv1.VerificationResultEvent{
		RequestId: cmd.GetRequestId(),
		PodKey:    cmd.GetPodKey(),
	}
	if strings.TrimSpace(cmd.GetRequestId()) == "" || strings.TrimSpace(cmd.GetPodKey()) == "" {
		result.Error = "request_id and pod_key are required"
		return result
	}
	if strings.TrimSpace(cmd.GetCommand()) == "" {
		result.Error = "verification command is required"
		return result
	}
	pod, ok := h.podStore.Get(cmd.GetPodKey())
	if !ok {
		result.Error = "pod not found"
		return result
	}
	workDir, err := podWorkspaceRoot(pod)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	timeout := verificationTimeout(cmd.GetTimeoutSeconds())
	ctx, cancel := context.WithTimeout(h.runner.GetRunContext(), timeout)
	defer cancel()

	output := &cappedOutputBuffer{limit: maxVerificationOutputBytes}
	shell, flag := envpath.ShellCommand()
	process := exec.CommandContext(ctx, shell, flag, cmd.GetCommand())
	process.Dir = workDir
	process.Stdout = output
	process.Stderr = output
	err = process.Run()

	result.Output = output.String()
	result.OutputTruncated = output.Truncated()
	if err == nil {
		return result
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		result.Error = "verification timed out"
		return result
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = int32(exitErr.ExitCode())
		return result
	}
	result.Error = err.Error()
	return result
}

func verificationTimeout(seconds int32) time.Duration {
	if seconds <= 0 {
		return defaultVerificationTimeout
	}
	timeout := time.Duration(seconds) * time.Second
	if timeout > maxVerificationTimeout {
		return maxVerificationTimeout
	}
	return timeout
}

type cappedOutputBuffer struct {
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func (buffer *cappedOutputBuffer) Write(data []byte) (int, error) {
	remaining := buffer.limit - buffer.buffer.Len()
	if remaining <= 0 {
		buffer.truncated = true
		return len(data), nil
	}
	if len(data) > remaining {
		buffer.truncated = true
		_, _ = buffer.buffer.Write(data[:remaining])
		return len(data), nil
	}
	_, _ = buffer.buffer.Write(data)
	return len(data), nil
}

func (buffer *cappedOutputBuffer) String() string {
	output := buffer.buffer.Bytes()
	end := len(output)
	for end > 0 && !utf8.Valid(output[:end]) {
		end--
	}
	if end != len(output) {
		buffer.truncated = true
	}
	return string(output[:end])
}

func (buffer *cappedOutputBuffer) Truncated() bool {
	return buffer.truncated
}
