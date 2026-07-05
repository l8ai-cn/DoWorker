package capability

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate_knownAxes(t *testing.T) {
	require.NoError(t, Validate("resume", "cli"))
	require.NoError(t, Validate("resume", "acp"))
	require.NoError(t, Validate("permission", "acp"))
	require.NoError(t, Validate("control", "set_model,set_permission_mode"))
}

func TestValidate_rejectsUnknown(t *testing.T) {
	require.Error(t, Validate("harness_mode", "native"))
	require.Error(t, Validate("resume", "warm-reattach"))
	require.Error(t, Validate("control", "set-model"))
}
