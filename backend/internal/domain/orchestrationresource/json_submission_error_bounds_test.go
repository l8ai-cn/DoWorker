package orchestrationresource

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeJSONSubmissionBoundsGenerationOverflowError(t *testing.T) {
	const tailMarker = "271828182845904523536"
	number := strings.Repeat("9", 50_000) + tailMarker
	source := []byte(`{
		"apiVersion":"agentsmesh.io/v1alpha1",
		"kind":"WorkerTemplate",
		"metadata":{
			"name":"worker-one",
			"namespace":"team-one",
			"generation":` + number + `
		},
		"spec":{"runtime":"codex"}
	}`)

	_, err := DecodeJSONSubmission(source)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrTypedJSONType)
	require.LessOrEqual(t, len(err.Error()), 512)
	require.Contains(t, err.Error(), "decode JSON manifest")
	require.Contains(t, err.Error(), "generation")
	require.Contains(t, err.Error(), "int64")
	require.Contains(t, err.Error(), "offset")
	require.NotContains(t, err.Error(), tailMarker)
}
