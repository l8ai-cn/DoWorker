package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration000232AgentCloudRebrandContract(t *testing.T) {
	up, err := FS.ReadFile("000232_agent_cloud_rebrand.up.sql")
	require.NoError(t, err)
	upSQL := string(up)

	for _, fragment := range []string{
		"orchestration_resources_api_version_check",
		"api_version IN ('agentcloud.io/v1alpha1', 'agentsmesh.io/v1alpha1')",
		"target_api_version IN ('agentcloud.io/v1alpha1', 'agentsmesh.io/v1alpha1')",
		"SET api_version = 'agentcloud.io/v1alpha1'",
		"SET target_api_version = 'agentcloud.io/v1alpha1'",
		"agentsmesh-plugin",
		"agentcloud-plugin",
	} {
		require.Contains(t, upSQL, fragment)
	}
}
