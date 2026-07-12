package sessionapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgentInteractionModes(t *testing.T) {
	t.Run("preserves configured order", func(t *testing.T) {
		modes, err := agentInteractionModes("pty,acp")
		require.NoError(t, err)
		require.Equal(t, []string{"pty", "acp"}, modes)
	})

	for _, raw := range []string{"", "acp,acp", "pty,unknown"} {
		t.Run(raw, func(t *testing.T) {
			_, err := agentInteractionModes(raw)
			require.Error(t, err)
		})
	}
}
