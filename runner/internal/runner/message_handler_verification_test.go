package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/client"
)

func TestRunVerificationUsesPodWorkspaceAndReturnsExitCode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell assertions use POSIX exit syntax")
	}
	handler, _ := verificationHandler(t)

	success := handler.runVerification(&runnerv1.RunVerificationCommand{
		RequestId: "verify-success",
		PodKey:    "pod-1",
		Command:   "printf verified",
	})
	require.Empty(t, success.Error)
	require.Zero(t, success.ExitCode)
	require.Equal(t, "verified", success.Output)

	failure := handler.runVerification(&runnerv1.RunVerificationCommand{
		RequestId: "verify-failure",
		PodKey:    "pod-1",
		Command:   "exit 23",
	})
	require.Empty(t, failure.Error)
	require.Equal(t, int32(23), failure.ExitCode)
}

func TestRunVerificationRejectsUnknownPod(t *testing.T) {
	handler, _ := verificationHandler(t)

	result := handler.runVerification(&runnerv1.RunVerificationCommand{
		RequestId: "verify-missing",
		PodKey:    "missing",
		Command:   "echo ignored",
	})

	require.Equal(t, "pod not found", result.Error)
}

func TestOnRunVerificationSendsResult(t *testing.T) {
	handler, conn := verificationHandler(t)
	if runtime.GOOS == "windows" {
		t.Skip("shell assertions use POSIX exit syntax")
	}

	require.NoError(t, handler.OnRunVerification(&runnerv1.RunVerificationCommand{
		RequestId: "verify-send",
		PodKey:    "pod-1",
		Command:   "exit 3",
	}))
	require.Len(t, conn.Events, 1)
	result, ok := conn.Events[0].Data.(*runnerv1.VerificationResultEvent)
	require.True(t, ok)
	require.Equal(t, int32(3), result.ExitCode)
}

func TestOnRunVerificationDuplicateRequestUsesCachedResult(t *testing.T) {
	handler, conn := verificationHandler(t)
	if runtime.GOOS == "windows" {
		t.Skip("shell assertions use POSIX redirection")
	}
	counterPath := filepath.Join(t.TempDir(), "count")
	command := fmt.Sprintf("printf x >> %q", counterPath)
	request := &runnerv1.RunVerificationCommand{
		RequestId: "verify-once",
		PodKey:    "pod-1",
		Command:   command,
	}

	require.NoError(t, handler.OnRunVerification(request))
	require.NoError(t, handler.OnRunVerification(request))

	content, err := os.ReadFile(counterPath)
	require.NoError(t, err)
	require.Equal(t, "x", string(content))
	require.Len(t, conn.Events, 2)
}

func TestOnRunVerificationCachedResultSurvivesHandlerRestart(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell assertions use POSIX redirection")
	}
	receiptRoot := t.TempDir()
	counterPath := filepath.Join(t.TempDir(), "count")
	request := &runnerv1.RunVerificationCommand{
		RequestId: "verify-durable",
		PodKey:    "pod-1",
		Command:   fmt.Sprintf("printf x >> %q", counterPath),
	}

	first, firstConn := verificationHandler(t)
	first.receipts = newCommandReceiptStore(receiptRoot)
	require.NoError(t, first.OnRunVerification(request))
	require.Len(t, firstConn.Events, 1)

	second, secondConn := verificationHandler(t)
	second.receipts = newCommandReceiptStore(receiptRoot)
	require.NoError(t, second.OnRunVerification(request))
	require.Len(t, secondConn.Events, 1)

	content, err := os.ReadFile(counterPath)
	require.NoError(t, err)
	require.Equal(t, "x", string(content))
}

func TestCappedOutputBufferPreservesUTF8AtLimit(t *testing.T) {
	buffer := &cappedOutputBuffer{limit: maxVerificationOutputBytes}
	output := strings.Repeat("a", maxVerificationOutputBytes-1) + "界"

	written, err := buffer.Write([]byte(output))

	require.NoError(t, err)
	require.Equal(t, len(output), written)
	require.True(t, utf8.ValidString(buffer.String()))
	require.Len(t, buffer.String(), maxVerificationOutputBytes-1)
	require.True(t, buffer.Truncated())
}

func verificationHandler(t *testing.T) (*RunnerMessageHandler, *client.MockConnection) {
	t.Helper()
	store := NewInMemoryPodStore()
	store.Put("pod-1", &Pod{PodKey: "pod-1", WorkDir: t.TempDir()})
	conn := client.NewMockConnection()
	return &RunnerMessageHandler{
		runner:   &Runner{},
		podStore: store,
		conn:     conn,
	}, conn
}
