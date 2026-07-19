package workercreation

import (
	"strings"

	"github.com/anthropics/agentsmesh/agentfile/parser"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

type freshPodLayer struct {
	config          map[string]any
	interactionMode specdomain.InteractionMode
	branch          string
	prompt          string
}

func parseFreshPodAgentfileLayer(source string) (freshPodLayer, error) {
	out := freshPodLayer{
		config:          map[string]any{},
		interactionMode: specdomain.InteractionModeACP,
	}
	source = strings.TrimSpace(source)
	if source == "" {
		return out, nil
	}
	program, parseErrors := parser.Parse(source)
	if len(parseErrors) > 0 {
		return freshPodLayer{}, invalidFreshPodDraft("agentfile_layer", parseErrors[0])
	}
	for _, decl := range program.Declarations {
		switch typed := decl.(type) {
		case *parser.ConfigDecl:
			out.config[typed.Name] = typed.Default
		case *parser.ModeDecl:
			out.interactionMode = specdomain.InteractionMode(typed.Mode)
		case *parser.BranchDecl:
			out.branch = stringLiteralValue(typed.Value)
		case *parser.PromptDecl:
			out.prompt = typed.Content
		}
	}
	return out, nil
}

func stringLiteralValue(expr parser.Expr) string {
	if value, ok := expr.(*parser.StringLit); ok {
		return strings.TrimSpace(value.Value)
	}
	return ""
}
