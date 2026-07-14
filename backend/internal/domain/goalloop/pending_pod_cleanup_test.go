package goalloop

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPendingPodCleanupRoundTrip(t *testing.T) {
	encoded := EncodePendingPodCleanup(
		StatusPaused,
		"verification exited with code 1",
		"runner unavailable",
	)

	state, ok := ParsePendingPodCleanup(&encoded)

	require.True(t, ok)
	require.Equal(t, StatusPaused, state.TargetStatus)
	require.Equal(t, "verification exited with code 1", state.Reason)
	require.Equal(t, "runner unavailable", state.StopError)
}

func TestPendingPodCleanupRejectsInvalidTarget(t *testing.T) {
	encoded := PendingPodCleanupErrorPrefix +
		`{"target_status":"active","reason":"verification failed","stop_error":"offline"}`

	_, ok := ParsePendingPodCleanup(&encoded)

	require.False(t, ok)
}
