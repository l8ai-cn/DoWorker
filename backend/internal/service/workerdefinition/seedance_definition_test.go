package workerdefinition

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSeedanceAgentFileUsesDoAgentUsageLogHome(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "config", "worker-types", "seedance-expert")
	content, err := os.ReadFile(filepath.Join(root, "AgentFile"))
	require.NoError(t, err)

	source := string(content)
	require.Contains(t, source, `sandbox.root + "/seedance-expert-home"`)
}
