package loopscript

import (
	"strings"
	"testing"
)

func TestParseRejectsInvalidPrograms(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantCode   string
		wantNodeID string
	}{
		{
			name:     "duplicate node id",
			source:   strings.Replace(canonicalSource, "n-tests", "n-fix-tax", 1),
			wantCode: "loop.node-id.duplicate",
		},
		{
			name:     "invalid identifier",
			source:   strings.Replace(canonicalSource, "checkout-fix", "Checkout_Fix", 1),
			wantCode: "loop.identifier.invalid",
		},
		{
			name:     "reserved identifier",
			source:   strings.Replace(canonicalSource, "loop checkout-fix", "loop admin", 1),
			wantCode: "loop.identifier.invalid",
		},
		{
			name:       "unknown verifier reference",
			source:     strings.Replace(canonicalSource, "until: tests.passed", "until: checks.passed", 1),
			wantCode:   "loop.reference.until-invalid",
			wantNodeID: "n-fix-cycle",
		},
		{
			name:       "repeat exceeds global limit",
			source:     strings.Replace(canonicalSource, "max: 5", "max: 6", 1),
			wantCode:   "loop.repeat.max-exceeds-limit",
			wantNodeID: "n-fix-cycle",
		},
		{
			name:     "unknown syntax",
			source:   strings.Replace(canonicalSource, "  on_failure pause", "  parallel work {}\n  on_failure pause", 1),
			wantCode: "loop.syntax.unknown",
		},
		{
			name: "worker declaration is not loop syntax",
			source: strings.Replace(
				canonicalSource,
				"  limits(",
				"  @id(n-coder)\n  worker coder = snapshot(42)\n  limits(",
				1,
			),
			wantCode: "loop.syntax.unknown",
		},
		{
			name:     "missing node id",
			source:   strings.Replace(canonicalSource, "    @id(n-tests)\n", "", 1),
			wantCode: "loop.node-id.missing",
		},
		{
			name:       "secret literal",
			source:     strings.Replace(canonicalSource, `prompt """修复结算页税额计算，并补充边界测试。"""`, "prompt secret", 1),
			wantCode:   "loop.secret.literal-forbidden",
			wantNodeID: "n-fix-tax",
		},
		{
			name:     "custom block reference requires a SHA-256 digest",
			source:   strings.Replace(customBlockSource, "a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0", "missing", 1),
			wantCode: "loop.custom-block.digest.invalid",
		},
		{
			name:     "custom block reference requires a version",
			source:   strings.Replace(customBlockSource, "version: 2", "version: 0", 1),
			wantCode: "loop.custom-block.version.invalid",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, diagnostics := Parse(test.source)
			if program != nil {
				t.Fatalf("Parse() program = %#v, want nil", program)
			}
			if len(diagnostics) == 0 {
				t.Fatal("Parse() diagnostics is empty")
			}
			if diagnostics[0].Code != test.wantCode {
				t.Fatalf("diagnostic code = %q, want %q", diagnostics[0].Code, test.wantCode)
			}
			if test.wantNodeID != "" && diagnostics[0].NodeID != test.wantNodeID {
				t.Fatalf("diagnostic node id = %q, want %q", diagnostics[0].NodeID, test.wantNodeID)
			}
			if diagnostics[0].Line < 1 || diagnostics[0].Column < 1 {
				t.Fatalf("diagnostic position = %d:%d", diagnostics[0].Line, diagnostics[0].Column)
			}
		})
	}
}

func TestParseMissingStructureKeepsEOFPosition(t *testing.T) {
	source := strings.Replace(canonicalSource, "  on_failure pause\n", "", 1)

	program, diagnostics := Parse(source)

	if program != nil {
		t.Fatalf("Parse() program = %#v, want nil", program)
	}
	if len(diagnostics) == 0 {
		t.Fatal("Parse() diagnostics is empty")
	}
	if diagnostics[0].Code != "loop.structure.failure-count" {
		t.Fatalf("diagnostic code = %q", diagnostics[0].Code)
	}
	if diagnostics[0].Line < 1 || diagnostics[0].Column < 1 {
		t.Fatalf("diagnostic position = %d:%d", diagnostics[0].Line, diagnostics[0].Column)
	}
}

func TestParseRejectsLargeIntegerWithoutArchitectureTruncation(t *testing.T) {
	source := strings.Replace(canonicalSource, "iterations: 5", "iterations: 4294967297", 1)

	program, diagnostics := Parse(source)

	if program != nil {
		t.Fatalf("Parse() program = %#v, want nil", program)
	}
	if len(diagnostics) == 0 || diagnostics[0].Code != "loop.value.out-of-range" {
		t.Fatalf("Parse() diagnostics = %#v", diagnostics)
	}
}
