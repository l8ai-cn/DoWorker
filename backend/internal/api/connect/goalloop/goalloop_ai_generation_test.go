package goalloopconnect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	airesourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	goalloopsvc "github.com/anthropics/agentsmesh/backend/internal/service/goalloop"
	goalloopv1 "github.com/anthropics/agentsmesh/proto/gen/go/goalloop/v1"
)

func TestGenerateLoopProgramReturnsCompiledProposal(t *testing.T) {
	program, canonical, diagnostics := compileLoopSource(loopSource)
	require.Empty(t, diagnostics)
	generator := &loopDraftGenerator{
		proposal: goalloopsvc.DraftProposal{
			Program:         program,
			CanonicalSource: canonical,
		},
	}
	server := NewServer(
		&loopGoalLoopService{},
		loopOrgService{},
		WithAIGeneration(generator),
	)

	response, err := server.GenerateLoopProgram(
		loopContext(),
		connect.NewRequest(&goalloopv1.GenerateLoopProgramRequest{
			OrgSlug:         "acme",
			Prompt:          "Create a professional PPT loop",
			CurrentSource:   loopSource,
			ModelResourceId: 42,
			Locale:          "zh-CN",
			Revision:        9,
		}),
	)

	require.NoError(t, err)
	require.Empty(t, response.Msg.Diagnostics)
	require.NotNil(t, response.Msg.Program)
	require.Equal(t, uint64(9), response.Msg.Revision)
	require.Equal(t, int64(7), generator.scope.OrganizationID)
	require.Equal(t, int64(9), generator.scope.UserID)
	require.Equal(t, int64(42), generator.input.ModelResourceID)
	require.Equal(t, "Create a professional PPT loop", generator.input.Prompt)
	require.Equal(t, loopSource, generator.input.CurrentSource)
	require.Equal(t, "zh-CN", generator.input.Locale)
}

func TestRepairLoopProgramReturnsRevisionedProposalAndPatch(t *testing.T) {
	program, canonical, diagnostics := compileLoopSource(loopSource)
	require.Empty(t, diagnostics)
	generator := &loopDraftGenerator{repairProposal: goalloopsvc.DraftRepairProposal{
		Program: program, CanonicalSource: canonical,
		Patch: goalloopsvc.DraftIntegerPatch{
			NodeID: "n-fix-cycle", FieldPath: "repeat.max",
			OldValue: 6, NewValue: 4,
		},
	}}
	server := NewServer(
		&loopGoalLoopService{},
		loopOrgService{},
		WithAIGeneration(generator),
	)

	response, err := server.RepairLoopProgram(
		loopContext(),
		connect.NewRequest(&goalloopv1.RepairLoopProgramRequest{
			OrgSlug: "acme", Source: loopSource,
			ModelResourceId: 42, Locale: "zh-CN", Revision: 11,
			DiagnosticCode: "loop.repeat.max-exceeds-limit",
			NodeId:         "n-fix-cycle", FieldPath: "repeat.max",
			Prompt: "减少循环次数",
		}),
	)

	require.NoError(t, err)
	require.Equal(t, uint64(11), response.Msg.Proposal.Revision)
	require.Equal(t, int64(6), response.Msg.Patch.OldValue)
	require.Equal(t, int64(4), response.Msg.Patch.NewValue)
	require.Equal(t, loopSource, generator.repairInput.Source)
	require.Equal(t, "repeat.max", generator.repairInput.FieldPath)
}

func TestGenerateLoopProgramMapsFailuresWithoutLeakingDetails(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code connect.Code
	}{
		{
			name: "invalid input",
			err:  goalloopsvc.ErrInvalidDraftGenerationInput,
			code: connect.CodeInvalidArgument,
		},
		{
			name: "secret-like input",
			err:  goalloopsvc.ErrDraftContainsSecret,
			code: connect.CodeInvalidArgument,
		},
		{
			name: "resource forbidden",
			err:  airesourceservice.ErrForbidden,
			code: connect.CodePermissionDenied,
		},
		{
			name: "resource missing",
			err:  airesourceservice.ErrNotFound,
			code: connect.CodeNotFound,
		},
		{
			name: "resource unhealthy",
			err:  airesourceservice.ErrUnhealthy,
			code: connect.CodeFailedPrecondition,
		},
		{
			name: "generated source invalid",
			err:  goalloopsvc.ErrGeneratedDraftInvalid,
			code: connect.CodeFailedPrecondition,
		},
		{
			name: "provider request",
			err:  errors.Join(goalloopsvc.ErrDraftProviderUnavailable, errors.New("secret provider body")),
			code: connect.CodeUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer(
				&loopGoalLoopService{},
				loopOrgService{},
				WithAIGeneration(&loopDraftGenerator{err: test.err}),
			)
			_, err := server.GenerateLoopProgram(
				loopContext(),
				connect.NewRequest(&goalloopv1.GenerateLoopProgramRequest{
					OrgSlug: "acme", Prompt: "make loop", ModelResourceId: 42, Locale: "zh-CN",
				}),
			)

			require.Equal(t, test.code, connect.CodeOf(err))
			require.NotContains(t, err.Error(), "secret")
		})
	}
}

func TestGenerateLoopProgramRequiresConfiguredGenerator(t *testing.T) {
	server := NewServer(&loopGoalLoopService{}, loopOrgService{})

	_, err := server.GenerateLoopProgram(
		loopContext(),
		connect.NewRequest(&goalloopv1.GenerateLoopProgramRequest{
			OrgSlug: "acme", Prompt: "make loop", ModelResourceId: 42, Locale: "zh-CN",
		}),
	)

	require.Equal(t, connect.CodeUnavailable, connect.CodeOf(err))
}

type loopDraftGenerator struct {
	proposal       goalloopsvc.DraftProposal
	repairProposal goalloopsvc.DraftRepairProposal
	err            error
	scope          goalloopsvc.DraftGenerationScope
	input          goalloopsvc.DraftGenerationInput
	repairInput    goalloopsvc.DraftRepairInput
}

func (generator *loopDraftGenerator) Generate(
	_ context.Context,
	scope goalloopsvc.DraftGenerationScope,
	input goalloopsvc.DraftGenerationInput,
) (goalloopsvc.DraftProposal, error) {
	generator.scope = scope
	generator.input = input
	return generator.proposal, generator.err
}

func (generator *loopDraftGenerator) Repair(
	_ context.Context,
	scope goalloopsvc.DraftGenerationScope,
	input goalloopsvc.DraftRepairInput,
) (goalloopsvc.DraftRepairProposal, error) {
	generator.scope = scope
	generator.repairInput = input
	return generator.repairProposal, generator.err
}
