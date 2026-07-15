package loopscript

import (
	"fmt"
	"strconv"
	"strings"
)

func Format(program *Program) (string, []Diagnostic) {
	if diagnostics := validateProgram(program, nil); len(diagnostics) != 0 {
		return "", diagnostics
	}

	loop := program.Loop
	repeat := loop.Repeat
	var output strings.Builder
	fmt.Fprintf(&output, "@id(%s)\n", loop.NodeID)
	fmt.Fprintf(&output, "loop %s {\n", loop.LocalID)
	fmt.Fprintf(&output, "  @id(%s)\n", loop.Worker.NodeID)
	fmt.Fprintf(
		&output, "  worker %s = snapshot(%d)\n",
		loop.Worker.LocalID, loop.Worker.SnapshotID,
	)
	fmt.Fprintf(
		&output,
		"  limits(iterations: %d, tokens: %d, timeout: %dm, no_progress: %d, same_error: %d)\n",
		loop.Limits.Iterations, loop.Limits.Tokens, loop.Limits.TimeoutMins,
		loop.Limits.NoProgress, loop.Limits.SameError,
	)
	fmt.Fprintf(&output, "  @id(%s)\n", repeat.NodeID)
	fmt.Fprintf(
		&output, "  repeat %s(max: %d, until: %s.%s) {\n",
		repeat.LocalID, repeat.Max, repeat.Until.LocalID, repeat.Until.Field,
	)
	fmt.Fprintf(&output, "    @id(%s)\n", repeat.Agent.NodeID)
	fmt.Fprintf(
		&output, "    agent %s(using: %s) { prompt %s }\n",
		repeat.Agent.LocalID, repeat.Agent.Using, formatPrompt(repeat.Agent.Prompt),
	)
	fmt.Fprintf(&output, "    @id(%s)\n", repeat.Verifier.NodeID)
	fmt.Fprintf(
		&output, "    verify %s { command %s accept %s }\n",
		repeat.Verifier.LocalID, strconv.Quote(repeat.Verifier.Command),
		strconv.Quote(repeat.Verifier.Accept),
	)
	output.WriteString("  }\n")
	fmt.Fprintf(&output, "  on_failure %s\n", loop.FailurePolicy)
	output.WriteString("}")
	return output.String(), nil
}

func formatPrompt(prompt string) string {
	if strings.Contains(prompt, `"""`) || strings.HasSuffix(prompt, `"`) ||
		strings.HasPrefix(prompt, " ") || strings.HasSuffix(prompt, " ") ||
		strings.Contains(prompt, "\n") {
		return strconv.Quote(prompt)
	}
	return `"""` + prompt + `"""`
}
