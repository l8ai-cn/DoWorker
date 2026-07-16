package goalloop

import "github.com/anthropics/agentsmesh/backend/internal/loopscript"

type draftIntegerTarget struct {
	current int64
	minimum int64
	maximum int64
	apply   func(*loopscript.Program, int64)
}

func integerRepairTarget(
	program *loopscript.Program,
	diagnostic loopscript.Diagnostic,
) (draftIntegerTarget, bool) {
	if diagnostic.Code != "loop.value.out-of-range" &&
		diagnostic.Code != "loop.repeat.max-exceeds-limit" {
		return draftIntegerTarget{}, false
	}
	loop := program.Loop
	switch diagnostic.FieldPath {
	case "limits.iterations":
		minimum := max(int64(1), loop.Repeat.Max)
		return draftIntegerTarget{
			current: loop.Limits.Iterations, minimum: minimum, maximum: 100,
			apply: func(value *loopscript.Program, next int64) {
				value.Loop.Limits.Iterations = next
			},
		}, true
	case "limits.tokens":
		return integerLimitTarget(
			loop.Limits.Tokens, 1, int64(^uint64(0)>>1),
			func(value *loopscript.Program, next int64) {
				value.Loop.Limits.Tokens = next
			},
		), true
	case "limits.timeout":
		return integerLimitTarget(
			loop.Limits.TimeoutMins, 1, 1440,
			func(value *loopscript.Program, next int64) {
				value.Loop.Limits.TimeoutMins = next
			},
		), true
	case "limits.no_progress":
		return integerLimitTarget(
			loop.Limits.NoProgress, 1, 20,
			func(value *loopscript.Program, next int64) {
				value.Loop.Limits.NoProgress = next
			},
		), true
	case "limits.same_error":
		return integerLimitTarget(
			loop.Limits.SameError, 1, 20,
			func(value *loopscript.Program, next int64) {
				value.Loop.Limits.SameError = next
			},
		), true
	case "repeat.max":
		return integerLimitTarget(
			loop.Repeat.Max, 1, minInt64(100, loop.Limits.Iterations),
			func(value *loopscript.Program, next int64) {
				value.Loop.Repeat.Max = next
			},
		), true
	default:
		return draftIntegerTarget{}, false
	}
}

func minInt64(left, right int64) int64 {
	if left < right {
		return left
	}
	return right
}

func integerLimitTarget(
	current, minimum, maximum int64,
	apply func(*loopscript.Program, int64),
) draftIntegerTarget {
	return draftIntegerTarget{
		current: current, minimum: minimum, maximum: maximum, apply: apply,
	}
}
