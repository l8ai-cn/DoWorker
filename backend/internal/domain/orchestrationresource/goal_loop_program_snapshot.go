package orchestrationresource

import (
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/loopscript"
)

type GoalLoopProgramSnapshot struct {
	CanonicalSource string                  `json:"canonicalSource" yaml:"canonicalSource"`
	CustomBlock     *GoalLoopCustomBlockPin `json:"customBlock,omitempty" yaml:"customBlock,omitempty"`
}

type GoalLoopCustomBlockPin struct {
	NodeID           string `json:"nodeId" yaml:"nodeId"`
	DefinitionID     string `json:"definitionId" yaml:"definitionId"`
	Slug             string `json:"slug" yaml:"slug"`
	Version          int64  `json:"version" yaml:"version"`
	DefinitionDigest string `json:"definitionDigest" yaml:"definitionDigest"`
}

func validateGoalLoopProgramSnapshot(
	metadata Metadata,
	spec *GoalLoopResourceSpec,
) error {
	snapshot := spec.LoopProgram
	if snapshot == nil {
		return nil
	}
	if err := validateDefinitionText(
		"loopProgram.canonicalSource",
		snapshot.CanonicalSource,
		65_536,
		true,
	); err != nil {
		return err
	}
	program, diagnostics := loopscript.Parse(snapshot.CanonicalSource)
	if len(diagnostics) != 0 {
		return fmt.Errorf("loopProgram.canonicalSource is invalid")
	}
	canonical, diagnostics := loopscript.Format(program)
	if len(diagnostics) != 0 || canonical != snapshot.CanonicalSource {
		return fmt.Errorf("loopProgram.canonicalSource must be canonical")
	}
	if err := validateGoalLoopProgramProjection(metadata, spec, program); err != nil {
		return err
	}
	return validateGoalLoopCustomBlockPin(snapshot.CustomBlock, program.Loop.Repeat.CustomBlock)
}

func validateGoalLoopProgramProjection(
	metadata Metadata,
	spec *GoalLoopResourceSpec,
	program *loopscript.Program,
) error {
	loop := program.Loop
	if loop.LocalID != metadata.Name.String() ||
		loop.Limits.Iterations != int64(spec.MaxIterations) ||
		loop.Repeat.Max != loop.Limits.Iterations ||
		loop.Limits.TimeoutMins != int64(spec.TimeoutMinutes) ||
		loop.Limits.NoProgress != int64(spec.NoProgressLimit) ||
		loop.Limits.SameError != int64(spec.SameErrorLimit) ||
		string(loop.FailurePolicy) != spec.EscalationPolicy {
		return fmt.Errorf("loopProgram.canonicalSource does not match GoalLoop execution fields")
	}
	if spec.TokenBudget == nil || loop.Limits.Tokens != *spec.TokenBudget {
		return fmt.Errorf("loopProgram.canonicalSource does not match tokenBudget")
	}
	if loop.Repeat.Agent.Prompt != spec.Objective ||
		loop.Repeat.Verifier.Command != spec.VerificationCommand ||
		len(spec.AcceptanceCriteria) != 1 ||
		loop.Repeat.Verifier.Accept != spec.AcceptanceCriteria[0] {
		return fmt.Errorf("loopProgram.canonicalSource does not match GoalLoop task fields")
	}
	return nil
}

func validateGoalLoopCustomBlockPin(
	pin *GoalLoopCustomBlockPin,
	source *loopscript.CustomBlockRef,
) error {
	if source == nil && pin == nil {
		return nil
	}
	if source == nil || pin == nil {
		return fmt.Errorf("loopProgram.customBlock must exactly match canonicalSource")
	}
	if pin.NodeID != source.NodeID ||
		pin.DefinitionID != source.DefinitionID ||
		pin.Slug != source.Slug ||
		pin.Version != source.Version ||
		pin.DefinitionDigest != source.DefinitionDigest {
		return fmt.Errorf("loopProgram.customBlock must exactly match canonicalSource")
	}
	return nil
}
