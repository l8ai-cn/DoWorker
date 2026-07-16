package loopscript

import (
	"strings"
	"testing"
)

func TestAnalyzeReturnsProgramWithSemanticDiagnostics(t *testing.T) {
	source := strings.Replace(canonicalSource, "max: 5", "max: 6", 1)

	program, diagnostics := Analyze(source)

	if program == nil {
		t.Fatal("Analyze() program = nil, want parsed program")
	}
	if program.Loop.Repeat.Max != 6 {
		t.Fatalf("Analyze() repeat max = %d, want 6", program.Loop.Repeat.Max)
	}
	requireDiagnostic(t, diagnostics, "loop.repeat.max-exceeds-limit", "repeat.max")

	parsed, parseDiagnostics := Parse(source)
	if parsed != nil {
		t.Fatalf("Parse() program = %#v, want nil", parsed)
	}
	requireDiagnostic(t, parseDiagnostics, "loop.repeat.max-exceeds-limit", "repeat.max")
}

func TestAnalyzeReturnsNilForSyntaxDiagnostics(t *testing.T) {
	source := strings.Replace(canonicalSource, "  on_failure pause", "  parallel work {}", 1)

	program, diagnostics := Analyze(source)

	if program != nil {
		t.Fatalf("Analyze() program = %#v, want nil", program)
	}
	requireDiagnostic(t, diagnostics, "loop.syntax.unknown", "")
}

func TestAnalyzeRejectsCredentialLiteralsWithoutRetainingThem(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantNodeID string
		fieldValue func(*Program) string
	}{
		{
			name: "agent prompt",
			source: strings.Replace(
				canonicalSource,
				`prompt """修复结算页税额计算，并补充边界测试。"""`,
				`prompt """use sk-ant-api03-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"""`,
				1,
			),
			wantNodeID: "n-fix-tax",
			fieldValue: func(program *Program) string {
				return program.Loop.Repeat.Agent.Prompt
			},
		},
		{
			name: "verifier command",
			source: strings.Replace(
				canonicalSource,
				`command "pnpm test --filter billing"`,
				`command "ghp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`,
				1,
			),
			wantNodeID: "n-tests",
			fieldValue: func(program *Program) string {
				return program.Loop.Repeat.Verifier.Command
			},
		},
		{
			name: "verifier accept",
			source: strings.Replace(
				canonicalSource,
				`accept "完整测试集通过"`,
				`accept "Authorization: Bearer abcdefghijklmnopqrstuvwxyz012345"`,
				1,
			),
			wantNodeID: "n-tests",
			fieldValue: func(program *Program) string {
				return program.Loop.Repeat.Verifier.Accept
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, diagnostics := Analyze(test.source)
			if program == nil {
				t.Fatal("Analyze() program = nil, want redacted program")
			}
			if value := test.fieldValue(program); value != "" {
				t.Fatalf("credential field retained in AST: %q", value)
			}
			diagnostic := requireDiagnostic(t, diagnostics, "loop.secret.literal-forbidden", "")
			if diagnostic.NodeID != test.wantNodeID {
				t.Fatalf("diagnostic node id = %q, want %q", diagnostic.NodeID, test.wantNodeID)
			}
			if hasDiagnostic(diagnostics, "loop.text.empty") {
				t.Fatalf("Analyze() added empty-text diagnostic after redaction: %#v", diagnostics)
			}

			parsed, parseDiagnostics := Parse(test.source)
			if parsed != nil {
				t.Fatalf("Parse() program = %#v, want nil", parsed)
			}
			requireDiagnostic(t, parseDiagnostics, "loop.secret.literal-forbidden", "")
		})
	}
}

func requireDiagnostic(t *testing.T, diagnostics []Diagnostic, code, fieldPath string) Diagnostic {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code && diagnostic.FieldPath == fieldPath {
			return diagnostic
		}
	}
	t.Fatalf("diagnostic %q with field path %q not found in %#v", code, fieldPath, diagnostics)
	return Diagnostic{}
}

func hasDiagnostic(diagnostics []Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}
