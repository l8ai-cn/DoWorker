package agentworkbench

import (
	"testing"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

func TestCommandPayloadDigestMatchesBrowserGolden(t *testing.T) {
	expectedRevision := uint64(42)
	command := &agentworkbenchv2.CommandEnvelope{
		SessionId:        "conv-1",
		StreamEpoch:      "epoch-1",
		CommandId:        "command-1",
		ExpectedRevision: &expectedRevision,
		IssuedAt:         "2026-07-16T10:00:00Z",
		Command: &agentworkbenchv2.CommandEnvelope_SendPrompt{
			SendPrompt: &agentworkbenchv2.SendPromptCommand{
				Text: "创建一个视频预览",
			},
		},
	}

	digest, err := CommandPayloadDigest(command)

	require.NoError(t, err)
	require.Equal(
		t,
		"sha256:f47c597114f8300cba123545827f460f98e98978a9e5e7036ac437fdd9c1e47b",
		digest,
	)
}
