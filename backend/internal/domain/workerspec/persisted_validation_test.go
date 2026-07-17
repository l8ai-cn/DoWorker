package workerspec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistedWorkerSpecAcceptsLegacyMissingProtocolAdapter(t *testing.T) {
	spec := validWorkerSpec()
	spec.Runtime.ModelBinding.ProtocolAdapter = ""
	raw, err := json.Marshal(spec)
	require.NoError(t, err)

	_, strictErr := DecodeSpec(raw)
	persisted, persistedErr := DecodePersistedSpec(raw)

	require.ErrorContains(t, strictErr, "protocol adapter")
	require.NoError(t, persistedErr)
	assert.False(t, HasResolvedProtocolAdapters(persisted))
}
