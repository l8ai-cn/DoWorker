package orchestrationresource

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeYAMLSubmissionEnforcesSeparateSourceLimit(t *testing.T) {
	atLimit := bytes.Repeat([]byte{'\n'}, maxYAMLManifestBytes)
	require.NoError(t, preflightYAML(atLimit))

	_, err := DecodeYAMLSubmission(append(atLimit, '\n'))
	require.Error(t, err)
	require.Contains(t, err.Error(), "262144")
}

func TestDecodeJSONSubmissionStillAcceptsOneMiBSource(t *testing.T) {
	source := append([]byte(validSubmissionJSON),
		bytes.Repeat([]byte{' '}, maxManifestBytes-len(validSubmissionJSON))...)
	require.Len(t, source, maxManifestBytes)

	manifest, err := DecodeJSONSubmission(source)
	require.NoError(t, err)
	require.NotEmpty(t, manifest.Spec)
}
