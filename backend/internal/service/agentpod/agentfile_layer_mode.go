package agentpod

import "strings"

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
	mode := ""
	for _, line := range strings.Split(*layer, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "MODE" &&
			(fields[1] == "pty" || fields[1] == "acp") {
			mode = fields[1]
		}
	}
	return mode
}
