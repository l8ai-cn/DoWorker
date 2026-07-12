package agent

import (
	"testing"

	"github.com/anthropics/agentsmesh/agentfile/eval"
	"github.com/anthropics/agentsmesh/agentfile/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildEvalContextExposesPromptForAgentfileArguments(t *testing.T) {
	program, errors := parser.Parse(`AGENT mmx
EXECUTABLE mmx
PROMPT_POSITION none
arg "text"
arg "chat"
arg "--message" prompt
`)
	require.Empty(t, errors)

	context := buildEvalContext(
		&ConfigBuildRequest{Prompt: "Implement the worker"},
		nil,
		nil,
		nil,
		nil,
	)

	require.NoError(t, eval.Eval(program, context))
	assert.Equal(t, []string{"text", "chat", "--message", "Implement the worker"}, context.Result.LaunchArgs)
	assert.Equal(t, "none", context.Result.PromptPosition)
}
