package loopscript

import (
	"strings"
	"testing"
)

func TestRangeDiagnosticsIncludeStableFieldPath(t *testing.T) {
	tests := []struct {
		name      string
		oldValue  string
		newValue  string
		fieldPath string
	}{
		{"iterations", "iterations: 5", "iterations: 101", "limits.iterations"},
		{"tokens", "tokens: 80000", "tokens: 0", "limits.tokens"},
		{"timeout", "timeout: 60m", "timeout: 0m", "limits.timeout"},
		{"no progress", "no_progress: 3", "no_progress: 0", "limits.no_progress"},
		{"same error", "same_error: 2", "same_error: 0", "limits.same_error"},
		{"repeat max", "max: 5", "max: 101", "repeat.max"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := strings.Replace(canonicalSource, test.oldValue, test.newValue, 1)

			program, diagnostics := Analyze(source)

			if program == nil {
				t.Fatal("Analyze() program = nil, want parsed program")
			}
			requireDiagnostic(t, diagnostics, "loop.value.out-of-range", test.fieldPath)
		})
	}
}

func TestRepeatMaxExceedsLimitDiagnosticIncludesFieldPath(t *testing.T) {
	source := strings.Replace(canonicalSource, "max: 5", "max: 6", 1)

	program, diagnostics := Analyze(source)

	if program == nil {
		t.Fatal("Analyze() program = nil, want parsed program")
	}
	requireDiagnostic(t, diagnostics, "loop.repeat.max-exceeds-limit", "repeat.max")
}
