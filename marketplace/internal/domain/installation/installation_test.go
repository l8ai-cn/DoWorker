package installation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstallationOperationLifecycle(t *testing.T) {
	item, err := New("installation-1", 42, 108, 301, "entitlement-1", 9, 14)
	require.NoError(t, err)
	operation, err := item.Plan("operation-1", "plan-1", "digest", []byte(`{"mutations":[]}`))
	require.NoError(t, err)

	require.NoError(t, operation.Start())
	require.Equal(t, OperationRunning, operation.Status())
	require.Equal(t, StatusInstalling, item.Status())

	require.NoError(t, operation.Succeed([]byte(`{"runtime_ref":"expert-18"}`)))
	require.NoError(t, item.Activate("expert-18"))
	require.Equal(t, OperationSucceeded, operation.Status())
	require.Equal(t, StatusActive, item.Status())
}

func TestFailedOperationCannotActivateInstallation(t *testing.T) {
	item, err := New("installation-1", 42, 108, 301, "entitlement-1", 9, 14)
	require.NoError(t, err)
	operation, err := item.Plan("operation-1", "plan-1", "digest", []byte(`{}`))
	require.NoError(t, err)
	require.NoError(t, operation.Start())
	require.NoError(t, operation.Fail("RUNTIME_FAILED", "运行时安装失败"))
	require.NoError(t, item.Fail())

	require.ErrorIs(t, item.Activate("expert-18"), ErrInvalidTransition)
	require.Equal(t, StatusFailed, item.Status())
}
