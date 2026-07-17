package orchestrationresource

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeYAMLRejectsGeneratedNodeCountOverLimit(t *testing.T) {
	manifest := validManifestForTest()
	items := strings.TrimSuffix(strings.Repeat("null,", maxYAMLNodes), ",")
	manifest.Spec = json.RawMessage(`{"runtime":"codex","items":[` + items + `]}`)

	encoded, err := EncodeYAML(manifest)

	require.Error(t, err)
	require.Nil(t, encoded)
	require.Contains(t, err.Error(), "10000 nodes")
}

func TestCountYAMLNodesFromJSONIsExact(t *testing.T) {
	count, err := countYAMLNodesFromJSON([]byte(`{"a":[1,true,null,{"b":"x"}]}`))

	require.NoError(t, err)
	require.Equal(t, 10, count)
}

func TestCountYAMLNodesFromJSONRejectsOverLimit(t *testing.T) {
	atLimit := strings.TrimSuffix(strings.Repeat("null,", maxYAMLNodes-2), ",")
	count, err := countYAMLNodesFromJSON([]byte(`[` + atLimit + `]`))
	require.NoError(t, err)
	require.Equal(t, maxYAMLNodes, count)

	overLimit := strings.TrimSuffix(strings.Repeat("null,", maxYAMLNodes-1), ",")
	count, err = countYAMLNodesFromJSON([]byte(`[` + overLimit + `]`))

	require.Error(t, err)
	require.Zero(t, count)
	require.Contains(t, err.Error(), "10000 nodes")
}

func TestEncodeYAMLRejectsOutputOverLimit(t *testing.T) {
	manifest := validManifestForTest()
	manifest.Spec = oversizedYAMLOutputSpec(t)

	encoded, err := EncodeYAML(manifest)

	require.ErrorIs(t, err, errYAMLOutputTooLarge)
	require.Nil(t, encoded)
}

func TestEncodeYAMLRejectsPhysicalLineOverLimit(t *testing.T) {
	manifest := submissionManifestForYAMLTest()
	manifest.Spec = json.RawMessage(
		`{"runtime":"codex","text":"` + strings.Repeat("x", 70<<10) + `"}`,
	)

	encoded, err := EncodeYAML(manifest)

	require.ErrorIs(t, err, errYAMLLineTooLong)
	require.Nil(t, encoded)
	require.Less(t, len(err.Error()), 200)
}

func TestEncodeYAMLAcceptsPhysicalLineAtLimit(t *testing.T) {
	manifest := submissionManifestForYAMLTest()
	manifest.Spec = json.RawMessage(
		`{"runtime":"codex","text":"` +
			strings.Repeat("x", maxYAMLLineBytes-len("  text: ")) + `"}`,
	)

	encoded, err := EncodeYAML(manifest)

	require.NoError(t, err)
	require.NoError(t, preflightYAML(encoded))
	_, err = DecodeYAMLSubmission(encoded)
	require.NoError(t, err)
}

func TestLimitedYAMLBufferRejectsLineAtomicallyAcrossWrites(t *testing.T) {
	buffer := newLimitedYAMLBuffer(maxYAMLManifestBytes)
	line := bytes.Repeat([]byte{'x'}, maxYAMLLineBytes)

	written, err := buffer.Write(line)
	require.NoError(t, err)
	require.Equal(t, len(line), written)
	before := append([]byte(nil), buffer.Bytes()...)

	written, err = buffer.Write([]byte("x"))
	require.ErrorIs(t, err, errYAMLLineTooLong)
	require.Zero(t, written)
	require.Equal(t, before, buffer.Bytes())
}

func TestLimitedYAMLBufferHandlesSplitCRLFAtLineBoundary(t *testing.T) {
	buffer := newLimitedYAMLBuffer(maxYAMLManifestBytes)
	line := bytes.Repeat([]byte{'x'}, maxYAMLLineBytes)

	for _, chunk := range [][]byte{line, {'\r'}, {'\n'}, line} {
		_, err := buffer.Write(chunk)
		require.NoError(t, err)
	}
	require.NoError(t, preflightYAML(buffer.Bytes()))
}

func TestEncodeYAMLNormalOutputStaysWithinLimit(t *testing.T) {
	encoded, err := EncodeYAML(submissionManifestForYAMLTest())

	require.NoError(t, err)
	require.NotEmpty(t, encoded)
	require.LessOrEqual(t, len(encoded), maxYAMLManifestBytes)
	require.NoError(t, preflightYAML(encoded))
	_, err = DecodeYAMLSubmission(encoded)
	require.NoError(t, err)
}

func oversizedYAMLOutputSpec(t *testing.T) json.RawMessage {
	t.Helper()
	items := make([]string, 8_000)
	for index := range items {
		items[index] = strings.Repeat("x", 40)
	}
	spec, err := json.Marshal(map[string]any{
		"runtime": "codex",
		"items":   items,
	})
	require.NoError(t, err)
	require.Less(t, len(spec), maxManifestBytes)
	return spec
}

func submissionManifestForYAMLTest() Manifest {
	manifest := validManifestForTest()
	manifest.Metadata.UID = ""
	manifest.Metadata.ResourceVersion = ""
	manifest.Metadata.Generation = 0
	return manifest
}
