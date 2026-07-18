package agentpod

import (
	"context"
	"testing"

	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePodProjectsSelectedRunnerClusterID(t *testing.T) {
	selector := &mockRunnerSelector{
		resolveRunner: &runnerDomain.Runner{ID: 1, ClusterID: 37},
	}
	orchestrator, _, _ := setupOrchestrator(t, withRunnerSelector(selector))

	result, err := createPodWithPlanSourceForTest(t, orchestrator, context.Background(), adapterTestCreateRequest())

	require.NoError(t, err)
	assert.Equal(t, int64(37), result.Pod.ClusterID)
}
