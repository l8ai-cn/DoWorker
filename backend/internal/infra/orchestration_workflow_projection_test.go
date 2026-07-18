package infra

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkflowProjectionUpdatesResetRevisionScopedRuntimeState(t *testing.T) {
	updates := workflowProjectionUpdates(orchestrationWorkflowRecord{
		OrchestrationResourceRevision: 2,
	})

	require.Contains(t, updates, "last_pod_key")
	require.Nil(t, updates["last_pod_key"])
	require.Contains(t, updates, "sandbox_path")
	require.Nil(t, updates["sandbox_path"])
}
