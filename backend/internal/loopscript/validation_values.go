package loopscript

import (
	"fmt"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/pkg/secretguard"
)

type textRedactions struct {
	agentPrompt     bool
	verifierCommand bool
	verifierAccept  bool
}

func appendRangeDiagnostics(diagnostics []Diagnostic, loop LoopNode, positions *programPositions) []Diagnostic {
	values := []struct {
		name     string
		value    int64
		minimum  int64
		maximum  int64
		nodeID   string
		position sourcePosition
	}{
		{"limits.iterations", loop.Limits.Iterations, 1, 100, loop.NodeID, limitsPosition(positions)},
		{"limits.tokens", loop.Limits.Tokens, 1, int64(^uint64(0) >> 1), loop.NodeID, limitsPosition(positions)},
		{"limits.timeout", loop.Limits.TimeoutMins, 1, 1440, loop.NodeID, limitsPosition(positions)},
		{"limits.no_progress", loop.Limits.NoProgress, 1, 20, loop.NodeID, limitsPosition(positions)},
		{"limits.same_error", loop.Limits.SameError, 1, 20, loop.NodeID, limitsPosition(positions)},
		{"repeat.max", loop.Repeat.Max, 1, 100, loop.Repeat.NodeID, repeatPosition(positions)},
	}
	for _, value := range values {
		if value.value < value.minimum || value.value > value.maximum {
			diagnostics = append(diagnostics, diagnosticForField(
				"loop.value.out-of-range",
				fmt.Sprintf("%s must be between %d and %d", value.name, value.minimum, value.maximum),
				value.nodeID, value.name, value.position,
			))
		}
	}
	return diagnostics
}

func appendTextDiagnostics(
	diagnostics []Diagnostic,
	loop LoopNode,
	positions *programPositions,
	redactions textRedactions,
) []Diagnostic {
	texts := []struct {
		name, value, nodeID string
		position            sourcePosition
		redacted            bool
	}{
		{
			"agent prompt", loop.Repeat.Agent.Prompt, loop.Repeat.Agent.NodeID,
			agentPosition(positions), redactions.agentPrompt,
		},
		{
			"verifier command", loop.Repeat.Verifier.Command, loop.Repeat.Verifier.NodeID,
			verifierPosition(positions), redactions.verifierCommand,
		},
		{
			"verifier acceptance", loop.Repeat.Verifier.Accept, loop.Repeat.Verifier.NodeID,
			verifierPosition(positions), redactions.verifierAccept,
		},
	}
	for _, text := range texts {
		if text.redacted {
			continue
		}
		if secretguard.ContainsCredentialLiteral(text.value) {
			diagnostics = append(diagnostics, diagnosticFor(
				"loop.secret.literal-forbidden",
				"secret literals are forbidden",
				text.nodeID, text.position,
			))
			continue
		}
		if strings.TrimSpace(text.value) == "" {
			diagnostics = append(diagnostics, diagnosticFor(
				"loop.text.empty", text.name+" must not be empty", text.nodeID, text.position,
			))
		}
	}
	return diagnostics
}
