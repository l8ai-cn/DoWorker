package loopscript

import (
	"reflect"
	"testing"
)

const canonicalSource = `@id(n-checkout-fix)
loop checkout-fix {
  @id(n-coder)
  worker coder = snapshot(42)
  limits(iterations: 5, tokens: 80000, timeout: 60m, no_progress: 3, same_error: 2)
  @id(n-fix-cycle)
  repeat fix-cycle(max: 5, until: tests.passed) {
    @id(n-fix-tax)
    agent fix-tax(using: coder) { prompt """修复结算页税额计算，并补充边界测试。""" }
    @id(n-tests)
    verify tests { command "pnpm test --filter billing" accept "完整测试集通过" }
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
			Worker: WorkerNode{
				NodeID:     "n-coder",
				LocalID:    "coder",
				SnapshotID: 42,
			},
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
					Using:   "coder",
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

func TestCompileGoalLoopV1MapsEveryField(t *testing.T) {
	program := mustParse(t, canonicalSource)

	spec, diagnostics := CompileGoalLoopV1(program)
	if len(diagnostics) != 0 {
		t.Fatalf("CompileGoalLoopV1() diagnostics = %#v", diagnostics)
	}

	want := &GoalLoopLaunchSpec{
		Name:                "checkout-fix",
		Slug:                "checkout-fix",
		WorkerSnapshotID:    42,
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
