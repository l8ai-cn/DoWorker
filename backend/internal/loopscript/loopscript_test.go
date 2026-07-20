package loopscript

import (
	"reflect"
	"testing"
)

const canonicalSource = `@id(n-checkout-fix)
loop checkout-fix {
  limits(iterations: 5, tokens: 80000, timeout: 60m, no_progress: 3, same_error: 2)
  @id(n-fix-cycle)
  repeat fix-cycle(max: 5, until: tests.passed) {
    @id(n-fix-tax)
    agent fix-tax { prompt """修复结算页税额计算，并补充边界测试。""" }
    @id(n-tests)
    verify tests { command "pnpm test --filter billing" accept "完整测试集通过" }
  }
  on_failure pause
}`

const customBlockSource = `@id(n-checkout-fix)
loop checkout-fix {
  limits(iterations: 5, tokens: 80000, timeout: 60m, no_progress: 3, same_error: 2)
  @id(n-fix-cycle)
  repeat fix-cycle(max: 5, until: ppt-step-check.passed) {
    custom_block(node_id: n-ppt-step, definition_id: "e54112b4-6a22-4ec4-b14d-dc3ac7c527a4", slug: ppt-step, version: 2, digest: "a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0")
    @id(n-ppt-step-ppt-step-task)
    agent ppt-step-task { prompt """制作季度复盘的专业 PPT""" }
    @id(n-ppt-step-ppt-step-check)
    verify ppt-step-check { command "test -f output.pptx" accept "output.pptx 存在且可打开" }
  }
  on_failure pause
}`

func TestParseCanonicalGoalLoopV1(t *testing.T) {
	program, diagnostics := Parse(canonicalSource)
	if len(diagnostics) != 0 {
		t.Fatalf("Parse() diagnostics = %#v", diagnostics)
	}

	want := &Program{
		SchemaVersion: 1,
		Loop: LoopNode{
			NodeID:  "n-checkout-fix",
			LocalID: "checkout-fix",
			Limits: Limits{
				Iterations:  5,
				Tokens:      80000,
				TimeoutMins: 60,
				NoProgress:  3,
				SameError:   2,
			},
			Repeat: RepeatNode{
				NodeID:  "n-fix-cycle",
				LocalID: "fix-cycle",
				Max:     5,
				Until: Reference{
					LocalID: "tests",
					Field:   "passed",
				},
				Agent: AgentNode{
					NodeID:  "n-fix-tax",
					LocalID: "fix-tax",
					Prompt:  "修复结算页税额计算，并补充边界测试。",
				},
				Verifier: VerifierNode{
					NodeID:  "n-tests",
					LocalID: "tests",
					Command: "pnpm test --filter billing",
					Accept:  "完整测试集通过",
				},
			},
			FailurePolicy: FailurePause,
		},
	}
	if !reflect.DeepEqual(program, want) {
		t.Fatalf("Parse() program mismatch\n got: %#v\nwant: %#v", program, want)
	}
}

func TestFormatRoundTripPreservesAST(t *testing.T) {
	program := mustParse(t, canonicalSource)

	formatted, diagnostics := Format(program)
	if len(diagnostics) != 0 {
		t.Fatalf("Format() diagnostics = %#v", diagnostics)
	}
	reparsed := mustParse(t, formatted)

	if !reflect.DeepEqual(reparsed, program) {
		t.Fatalf("round trip AST mismatch\n got: %#v\nwant: %#v", reparsed, program)
	}
}

func TestFormatIsStableForSameAST(t *testing.T) {
	program := mustParse(t, canonicalSource)

	first, diagnostics := Format(program)
	if len(diagnostics) != 0 {
		t.Fatalf("first Format() diagnostics = %#v", diagnostics)
	}
	second, diagnostics := Format(program)
	if len(diagnostics) != 0 {
		t.Fatalf("second Format() diagnostics = %#v", diagnostics)
	}

	if first != canonicalSource {
		t.Fatalf("Format() = %q, want canonical source %q", first, canonicalSource)
	}
	if second != first {
		t.Fatalf("Format() was not stable\nfirst:  %q\nsecond: %q", first, second)
	}
}

func TestFormatRoundTripsPromptsEndingInQuotes(t *testing.T) {
	for _, prompt := range []string{
		`finish with "`,
		`finish with ""`,
		"  preserve edge spaces  ",
		"first line\n  indented line\n",
		`contains """ delimiter`,
	} {
		program := mustParse(t, canonicalSource)
		program.Loop.Repeat.Agent.Prompt = prompt

		formatted, diagnostics := Format(program)
		if len(diagnostics) != 0 {
			t.Fatalf("Format() diagnostics = %#v", diagnostics)
		}
		reparsed := mustParse(t, formatted)
		if reparsed.Loop.Repeat.Agent.Prompt != prompt {
			t.Fatalf("prompt = %q, want %q", reparsed.Loop.Repeat.Agent.Prompt, prompt)
		}
	}
}

func TestCustomBlockReferenceRoundTripsWithoutVersionDrift(t *testing.T) {
	program := mustParse(t, customBlockSource)
	custom := program.Loop.Repeat.CustomBlock
	if custom == nil {
		t.Fatal("custom block reference is missing")
	}
	if custom.DefinitionID != "e54112b4-6a22-4ec4-b14d-dc3ac7c527a4" ||
		custom.Slug != "ppt-step" ||
		custom.Version != 2 ||
		custom.DefinitionDigest != "a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0" {
		t.Fatalf("custom block reference = %#v", custom)
	}

	formatted, diagnostics := Format(program)
	if len(diagnostics) != 0 {
		t.Fatalf("Format() diagnostics = %#v", diagnostics)
	}
	if formatted != customBlockSource {
		t.Fatalf("Format() = %q, want %q", formatted, customBlockSource)
	}
	reparsed := mustParse(t, formatted)
	if !reflect.DeepEqual(reparsed.Loop.Repeat.CustomBlock, custom) {
		t.Fatalf("custom block changed after format\n got: %#v\nwant: %#v", reparsed.Loop.Repeat.CustomBlock, custom)
	}
}

func TestCompileGoalLoopV1MapsEveryField(t *testing.T) {
	program := mustParse(t, canonicalSource)

	spec, diagnostics := CompileGoalLoopV1(program)
	if len(diagnostics) != 0 {
		t.Fatalf("CompileGoalLoopV1() diagnostics = %#v", diagnostics)
	}

	want := &GoalLoopLaunchSpec{
		Name:                "checkout-fix",
		Slug:                "checkout-fix",
		Objective:           "修复结算页税额计算，并补充边界测试。",
		AcceptanceCriteria:  []string{"完整测试集通过"},
		VerificationCommand: "pnpm test --filter billing",
		MaxIterations:       5,
		TokenBudget:         80000,
		TimeoutMinutes:      60,
		NoProgressLimit:     3,
		SameErrorLimit:      2,
		EscalationPolicy:    "pause",
	}
	if !reflect.DeepEqual(spec, want) {
		t.Fatalf("CompileGoalLoopV1() mismatch\n got: %#v\nwant: %#v", spec, want)
	}
}

func mustParse(t *testing.T, source string) *Program {
	t.Helper()
	program, diagnostics := Parse(source)
	if len(diagnostics) != 0 {
		t.Fatalf("Parse() diagnostics = %#v", diagnostics)
	}
	if program == nil {
		t.Fatal("Parse() returned nil program without diagnostics")
	}
	return program
}
