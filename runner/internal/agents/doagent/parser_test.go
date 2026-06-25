package doagent

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoAgentParser_SumsLLMResponseUsage(t *testing.T) {
	sandbox := t.TempDir()
	logsDir := sandbox + "/do-agent-home/logs"
	require.NoError(t, os.MkdirAll(logsDir, 0o755))

	log := `{"event":"llm_response","model":"openai/gpt-5.4","usage":{"prompt_tokens":100,"completion_tokens":40,"total_tokens":140}}
{"event":"llm_response","model":"openai/gpt-5.4","usage":{"prompt_tokens":50,"completion_tokens":10,"total_tokens":60}}
{"event":"task_result","total_input_tokens":150,"total_output_tokens":50}
`
	require.NoError(t, os.WriteFile(logsDir+"/sess.jsonl", []byte(log), 0o644))

	usage, err := (&doagentParser{}).Parse(sandbox, time.Unix(0, 0))
	require.NoError(t, err)
	require.NotNil(t, usage)

	m := usage.Models["openai/gpt-5.4"]
	require.NotNil(t, m)
	assert.Equal(t, int64(150), m.InputTokens)
	assert.Equal(t, int64(50), m.OutputTokens)
}

func TestDoAgentParser_SkipsTaskResultTotals(t *testing.T) {
	sandbox := t.TempDir()
	logsDir := sandbox + "/do-agent-home/logs"
	require.NoError(t, os.MkdirAll(logsDir, 0o755))
	require.NoError(t, os.WriteFile(logsDir+"/sess.jsonl",
		[]byte(`{"event":"task_result","total_input_tokens":999,"total_output_tokens":888}`+"\n"), 0o644))

	usage, err := (&doagentParser{}).Parse(sandbox, time.Unix(0, 0))
	require.NoError(t, err)
	assert.Nil(t, usage)
}
