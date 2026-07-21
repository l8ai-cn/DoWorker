package agentpod

import (
	"github.com/l8ai-cn/agentcloud/agentfile/parser"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

func planSourceArtifactLayer(
	spec specdomain.Spec,
	agentfileLayer string,
) string {
	layer := normalizedPlanAgentfileLayer(agentfileLayer)
	program, parseErrors := parser.Parse(layer)
	if len(parseErrors) > 0 || planLayerHasMode(program) {
		return layer
	}
	return appendPlanLayerLineForTest(
		layer,
		"MODE "+string(spec.TypeConfig.InteractionMode),
	)
}

func planLayerHasMode(program *parser.Program) bool {
	for _, declaration := range program.Declarations {
		switch declaration.(type) {
		case *parser.ModeDecl, *parser.ModeArgsDecl:
			return true
		}
	}
	return false
}
