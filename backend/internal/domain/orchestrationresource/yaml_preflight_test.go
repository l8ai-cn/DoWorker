package orchestrationresource

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYAMLPreflightRejectsInvalidUTF8(t *testing.T) {
	source := append([]byte(validSubmissionYAML), 0xff)

	_, err := DecodeYAMLSubmission(source)

	require.Error(t, err)
	require.Contains(t, err.Error(), "UTF-8")
	require.Less(t, len(err.Error()), 200)
}

func TestYAMLPreflightRejectsOversizedPhysicalLine(t *testing.T) {
	tests := map[string]string{
		"flow":   "value: [" + strings.Repeat("0,", maxYAMLLineBytes/2+1) + "0]\n",
		"quoted": "value: \"" + strings.Repeat("x", maxYAMLLineBytes+1) + "\"\n",
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			err := preflightYAML([]byte(source))
			require.Error(t, err)
			require.Contains(t, err.Error(), "line 1")
			require.Contains(t, err.Error(), "65536 bytes")
		})
	}
}

func TestYAMLPreflightAllowsLargeBlockScalarWithShortLines(t *testing.T) {
	blockLine := "    " + strings.Repeat("x", 1024) + "\n"
	source := strings.Replace(validSubmissionYAML, "  runtime: codex",
		"  text: |\n"+strings.Repeat(blockLine, 240), 1)

	require.Greater(t, len(source), maxYAMLLineBytes)
	require.LessOrEqual(t, len(source), maxYAMLManifestBytes)
	manifest, err := DecodeYAMLSubmission([]byte(source))
	require.NoError(t, err)
	require.NotEmpty(t, manifest.Spec)
}

func TestYAMLErrorsAreBoundedAndDoNotLeakLongTokens(t *testing.T) {
	longValue := strings.Repeat("untrusted", 6_000)
	tests := map[string]string{
		"alias": strings.Replace(validSubmissionYAML, "  runtime: codex",
			"  runtime: codex\n  payload: *"+longValue, 1),
		"tag": strings.Replace(validSubmissionYAML, "  runtime: codex",
			"  runtime: codex\n  payload: !"+longValue+" literal", 1),
		"parser": "value: !<tag:" + longValue + "\n",
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			require.LessOrEqual(t, len(source), maxYAMLManifestBytes)
			for line := range strings.SplitSeq(source, "\n") {
				require.LessOrEqual(t, len(line), maxYAMLLineBytes)
			}

			_, err := DecodeYAMLSubmission([]byte(source))
			require.Error(t, err)
			require.Less(t, len(err.Error()), 200)
			require.NotContains(t, err.Error(), strings.Repeat("untrusted", 20))
		})
	}
}
