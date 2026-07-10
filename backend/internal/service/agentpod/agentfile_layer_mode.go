package agentpod

import "github.com/anthropics/agentsmesh/agentfile/parser"

// Cross-module contract: an explicitly requested mode must win over the
// automation-level adapter's forced MODE (see CreatePod), otherwise a user (or
// the session API `pty_only` path) asking for CLI/PTY is silently downgraded to
// ACP under the default autonomous level.
func agentfileLayerHasModeDecl(layer *string) bool {
	if layer == nil {
		return false
	}
	program, _ := parser.Parse(*layer)
	for _, declaration := range program.Declarations {
		if _, ok := declaration.(*parser.ModeDecl); ok {
			return true
		}
	}
	return false
}
