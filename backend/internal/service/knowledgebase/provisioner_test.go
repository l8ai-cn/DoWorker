package knowledgebase

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderScaffold(t *testing.T) {
	changes, err := renderScaffold("Team Docs", "Everything the team knows.")
	require.NoError(t, err)

	byPath := map[string]string{}
	for _, ch := range changes {
		byPath[ch.Path] = ch.Content
	}
	require.Len(t, byPath, 5)

	llms := byPath["llms.txt"]
	require.NotEmpty(t, llms)
	assert.True(t, strings.HasPrefix(llms, "# Team Docs\n"), "llms.txt must start with H1 name")
	assert.Contains(t, llms, "> Everything the team knows.")
	assert.Contains(t, llms, "[Overview](wiki/index.md)")

	agents := byPath["AGENTS.md"]
	assert.Contains(t, agents, "raw/")
	assert.Contains(t, agents, "wiki/")
	assert.Contains(t, agents, "llms.txt")

	assert.Contains(t, byPath, "wiki/index.md")
	assert.Contains(t, byPath["wiki/log.md"], "init | Knowledge base created")
	assert.Contains(t, byPath, "raw/README.md")
}

func TestRenderScaffold_DefaultDescription(t *testing.T) {
	changes, err := renderScaffold("KB", "")
	require.NoError(t, err)
	for _, ch := range changes {
		if ch.Path == "llms.txt" {
			assert.Contains(t, ch.Content, "> Knowledge base maintained by AgentsMesh agents.")
		}
	}
}
