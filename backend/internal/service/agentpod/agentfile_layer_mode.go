package agentpod

import "github.com/anthropics/agentsmesh/agentfile/parser"

// agentfileLayerHasModeDecl reports whether the user-supplied AgentFile layer
// already selects an interaction mode via a bare `MODE pty` / `MODE acp`
// declaration. Per-mode argument declarations (`MODE acp "app-server"`) carry a
// third token and are not mode selectors, so they are ignored.
//
// Cross-module contract: an explicitly requested mode must win over the
// automation-level adapter's forced MODE (see CreatePod), otherwise a user (or
// the session API `pty_only` path) asking for CLI/PTY is silently downgraded to
// ACP under the default autonomous level.
func agentfileLayerHasModeDecl(layer *string) bool {
	return agentfileLayerMode(layer) != ""
}

func agentfileLayerMode(layer *string) string {
	if layer == nil {
		return ""
	}
	program, _ := parser.Parse(*layer)
	mode := ""
	for _, declaration := range program.Declarations {
		if modeDeclaration, ok := declaration.(*parser.ModeDecl); ok {
			mode = modeDeclaration.Mode
		}
	}
	return mode
}
