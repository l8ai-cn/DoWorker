package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/agentfile/extract"
	"github.com/anthropics/agentsmesh/agentfile/parser"
	"github.com/anthropics/agentsmesh/agentfile/serialize"
)

func TestParse_CapabilityDecl(t *testing.T) {
	src := `AGENT do-agent
CAPABILITY resume cli
CAPABILITY permission notification
CAPABILITY control set_model,set_permission_mode
CAPABILITY usage live
CAPABILITY interrupt true
`
	prog, errs := parser.Parse(src)
	require.Empty(t, errs)
	require.Len(t, prog.Declarations, 6)

	cap0 := prog.Declarations[1].(*parser.CapabilityDecl)
	assert.Equal(t, "resume", cap0.Axis)
	assert.Equal(t, "cli", cap0.Value)

	spec := extract.Extract(prog)
	assert.Equal(t, map[string]string{
		"resume":     "cli",
		"permission": "notification",
		"control":    "set_model,set_permission_mode",
		"usage":      "live",
		"interrupt":  "true",
	}, spec.Capabilities)
}

func TestParse_CapabilityDecl_rejectsUnknownAxis(t *testing.T) {
	_, errs := parser.Parse("CAPABILITY harness_mode native\n")
	require.NotEmpty(t, errs)
}

func TestSerialize_CapabilityDecl_roundTrip(t *testing.T) {
	src := "AGENT claude\nCAPABILITY resume cli\nCAPABILITY model_family claude\n"
	prog, errs := parser.Parse(src)
	require.Empty(t, errs)
	out := serialize.Serialize(prog)
	prog2, errs2 := parser.Parse(out)
	require.Empty(t, errs2)
	assert.Equal(t, prog.Declarations[1], prog2.Declarations[1])
}
