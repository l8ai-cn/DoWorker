package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/l8ai-cn/agentcloud/agentfile/extract"
	"github.com/l8ai-cn/agentcloud/agentfile/merge"
	"github.com/l8ai-cn/agentcloud/agentfile/parser"
	"github.com/l8ai-cn/agentcloud/agentfile/serialize"
)

func TestParse_KnowledgeDecl(t *testing.T) {
	prog, errs := parser.Parse("AGENT claude\nKNOWLEDGE team-docs [rw], product-wiki\n")
	require.Empty(t, errs)
	require.Len(t, prog.Declarations, 2)

	kd := prog.Declarations[1].(*parser.KnowledgeDecl)
	require.Len(t, kd.Mounts, 2)
	assert.Equal(t, parser.KnowledgeMountRef{Slug: "team-docs", Mode: "rw"}, kd.Mounts[0])
	assert.Equal(t, parser.KnowledgeMountRef{Slug: "product-wiki", Mode: "ro"}, kd.Mounts[1])

	spec := extract.Extract(prog)
	require.Len(t, spec.Knowledge, 2)
	assert.Equal(t, "rw", spec.Knowledge[0].Mode)
}

func TestParse_KnowledgeDecl_rejectsBadMode(t *testing.T) {
	_, errs := parser.Parse("KNOWLEDGE team-docs [wx]\n")
	require.NotEmpty(t, errs)
}

func TestSerialize_KnowledgeDecl_roundTrip(t *testing.T) {
	src := "AGENT claude\nKNOWLEDGE team-docs [rw], product-wiki\n"
	prog, errs := parser.Parse(src)
	require.Empty(t, errs)
	out := serialize.Serialize(prog)
	prog2, errs2 := parser.Parse(out)
	require.Empty(t, errs2)
	assert.Equal(t, prog.Declarations[1], prog2.Declarations[1])
}

func TestMerge_KnowledgeDecl_unionAndModeOverride(t *testing.T) {
	base, errs := parser.Parse("AGENT claude\nKNOWLEDGE team-docs, product-wiki\n")
	require.Empty(t, errs)
	layer, errs := parser.Parse("KNOWLEDGE team-docs [rw], runbooks\n")
	require.Empty(t, errs)

	merge.Merge(base, layer)
	spec := extract.Extract(base)
	require.Len(t, spec.Knowledge, 3)
	assert.Equal(t, "team-docs", spec.Knowledge[0].Slug)
	assert.Equal(t, "rw", spec.Knowledge[0].Mode, "layer rw must upgrade base ro")
	assert.Equal(t, "product-wiki", spec.Knowledge[1].Slug)
	assert.Equal(t, "ro", spec.Knowledge[1].Mode)
	assert.Equal(t, "runbooks", spec.Knowledge[2].Slug)
}
