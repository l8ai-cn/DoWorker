package goalloopconnect

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/loopscript"
	goalloopv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/goalloop/v1"
)

func loopProgramToProto(program *loopscript.Program) *goalloopv1.LoopProgram {
	loop := program.Loop
	repeat := loop.Repeat
	return &goalloopv1.LoopProgram{
		SchemaVersion: int32(program.SchemaVersion),
		Loop:          loopIdentity(loop.NodeID, loop.LocalID),
		Limits: &goalloopv1.LoopLimits{
			Iterations:     loop.Limits.Iterations,
			Tokens:         loop.Limits.Tokens,
			TimeoutMinutes: loop.Limits.TimeoutMins,
			NoProgress:     loop.Limits.NoProgress,
			SameError:      loop.Limits.SameError,
		},
		Repeat: &goalloopv1.LoopRepeatNode{
			Identity: loopIdentity(repeat.NodeID, repeat.LocalID),
			Max:      repeat.Max,
			Until: &goalloopv1.LoopReference{
				LocalId: repeat.Until.LocalID,
				Field:   repeat.Until.Field,
			},
			Agent: &goalloopv1.LoopAgentNode{
				Identity: loopIdentity(repeat.Agent.NodeID, repeat.Agent.LocalID),
				Prompt:   repeat.Agent.Prompt,
			},
			Verifier: &goalloopv1.LoopVerifierNode{
				Identity: loopIdentity(repeat.Verifier.NodeID, repeat.Verifier.LocalID),
				Command:  repeat.Verifier.Command,
				Accept:   repeat.Verifier.Accept,
			},
			CustomBlock: customBlockRefToProto(repeat.CustomBlock),
		},
		FailurePolicy: string(loop.FailurePolicy),
	}
}

func customBlockRefToProto(value *loopscript.CustomBlockRef) *goalloopv1.LoopCustomBlockRef {
	if value == nil {
		return nil
	}
	return &goalloopv1.LoopCustomBlockRef{
		NodeId:           value.NodeID,
		DefinitionId:     value.DefinitionID,
		Slug:             value.Slug,
		Version:          uint32(value.Version),
		DefinitionDigest: value.DefinitionDigest,
	}
}

func loopIdentity(nodeID, localID string) *goalloopv1.LoopNodeIdentity {
	return &goalloopv1.LoopNodeIdentity{NodeId: nodeID, LocalId: localID}
}

func loopDiagnosticsToProto(diagnostics []loopscript.Diagnostic) []*goalloopv1.LoopDiagnostic {
	items := make([]*goalloopv1.LoopDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		items = append(items, &goalloopv1.LoopDiagnostic{
			Code:      diagnostic.Code,
			Message:   diagnostic.Message,
			NodeId:    diagnostic.NodeID,
			Line:      int32(diagnostic.Line),
			Column:    int32(diagnostic.Column),
			FieldPath: diagnostic.FieldPath,
		})
	}
	return items
}
