package goalloop

import (
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/loopscript"
)

func enforceDraftMutationPolicy(
	current *loopscript.Program,
	proposed *loopscript.Program,
) error {
	if current == nil {
		if proposed != nil && proposed.Loop.Repeat.Verifier.Command == "false" {
			return nil
		}
		return fmt.Errorf("%w: new Loop verifier must be inert", ErrGeneratedDraftInvalid)
	}
	if proposed == nil || !sameDraftIdentity(current, proposed) ||
		!sameDraftVerifier(current, proposed) ||
		weakensDraftLimits(current, proposed) ||
		weakensFailurePolicy(current, proposed) {
		return fmt.Errorf("%w: protected Loop semantics changed", ErrGeneratedDraftInvalid)
	}
	return nil
}

func sameDraftIdentity(current, proposed *loopscript.Program) bool {
	left, right := current.Loop, proposed.Loop
	return left.NodeID == right.NodeID &&
		left.LocalID == right.LocalID &&
		left.Repeat.NodeID == right.Repeat.NodeID &&
		left.Repeat.LocalID == right.Repeat.LocalID &&
		left.Repeat.Agent.NodeID == right.Repeat.Agent.NodeID &&
		left.Repeat.Agent.LocalID == right.Repeat.Agent.LocalID &&
		left.Repeat.Verifier.NodeID == right.Repeat.Verifier.NodeID &&
		left.Repeat.Verifier.LocalID == right.Repeat.Verifier.LocalID &&
		sameCustomBlockReference(left.Repeat.CustomBlock, right.Repeat.CustomBlock)
}

func sameCustomBlockReference(
	left *loopscript.CustomBlockRef,
	right *loopscript.CustomBlockRef,
) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.NodeID == right.NodeID &&
		left.DefinitionID == right.DefinitionID &&
		left.Slug == right.Slug &&
		left.Version == right.Version &&
		left.DefinitionDigest == right.DefinitionDigest
}

func sameDraftVerifier(current, proposed *loopscript.Program) bool {
	left, right := current.Loop.Repeat, proposed.Loop.Repeat
	return left.Until == right.Until &&
		left.Verifier.Command == right.Verifier.Command &&
		left.Verifier.Accept == right.Verifier.Accept
}

func weakensDraftLimits(current, proposed *loopscript.Program) bool {
	left, right := current.Loop, proposed.Loop
	return right.Limits.Iterations > left.Limits.Iterations ||
		right.Limits.Tokens > left.Limits.Tokens ||
		right.Limits.TimeoutMins > left.Limits.TimeoutMins ||
		right.Limits.NoProgress > left.Limits.NoProgress ||
		right.Limits.SameError > left.Limits.SameError ||
		right.Repeat.Max > left.Repeat.Max
}

func weakensFailurePolicy(current, proposed *loopscript.Program) bool {
	return current.Loop.FailurePolicy == loopscript.FailureFail &&
		proposed.Loop.FailurePolicy != loopscript.FailureFail
}
