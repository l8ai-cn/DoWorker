package goalloopconnect

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	goalloopsvc "github.com/anthropics/agentsmesh/backend/internal/service/goalloop"
	goalloopv1 "github.com/anthropics/agentsmesh/proto/gen/go/goalloop/v1"
)

const loopSource = `@id(n-checkout-fix)
loop checkout-fix {
  limits(iterations: 5, tokens: 80000, timeout: 60m, no_progress: 3, same_error: 2)
  @id(n-fix-cycle)
  repeat fix-cycle(max: 5, until: tests.passed) {
    @id(n-fix-tax)
    agent fix-tax { prompt """fix checkout tax""" }
    @id(n-tests)
    verify tests { command "pnpm test" accept "tests pass" }
  }
  on_failure pause
}`

func TestCompileLoopProgramReturnsCanonicalAST(t *testing.T) {
	service := &loopGoalLoopService{}
	server := NewServer(service, loopOrgService{})

	response, err := server.CompileLoopProgram(
		loopContext(), connect.NewRequest(&goalloopv1.CompileLoopProgramRequest{
			OrgSlug:  "acme",
			Source:   loopSource,
			Revision: 7,
		}),
	)

	require.NoError(t, err)
	require.Empty(t, response.Msg.Diagnostics)
	require.NotEmpty(t, response.Msg.CanonicalSource)
	require.Equal(t, "n-checkout-fix", response.Msg.Program.Loop.NodeId)
	require.Equal(t, "fix checkout tax", response.Msg.Program.Repeat.Agent.Prompt)
	require.Equal(t, uint64(7), response.Msg.Revision)
	require.Zero(t, service.validateCalls)
}

func TestCompileLoopProgramReturnsDiagnosticsForInvalidSource(t *testing.T) {
	server := NewServer(&loopGoalLoopService{}, loopOrgService{})

	response, err := server.CompileLoopProgram(
		loopContext(), connect.NewRequest(&goalloopv1.CompileLoopProgramRequest{
			OrgSlug: "acme",
			Source:  "loop broken {}",
		}),
	)

	require.NoError(t, err)
	require.Nil(t, response.Msg.Program)
	require.Empty(t, response.Msg.CanonicalSource)
	require.NotEmpty(t, response.Msg.Diagnostics)
	require.NotEmpty(t, response.Msg.Diagnostics[0].Code)
}

func TestCompileLoopProgramDoesNotRequireWorkerSnapshot(t *testing.T) {
	service := &loopGoalLoopService{validateErr: goalloopsvc.ErrInvalidInput}
	server := NewServer(service, loopOrgService{})

	response, err := server.CompileLoopProgram(
		loopContext(), connect.NewRequest(&goalloopv1.CompileLoopProgramRequest{
			OrgSlug: "acme",
			Source:  loopSource,
		}),
	)

	require.NoError(t, err)
	require.NotNil(t, response.Msg.Program)
	require.Empty(t, response.Msg.Diagnostics)
	require.Zero(t, service.validateCalls)
}

func TestCompileLoopProgramWorksWithoutExecutionService(t *testing.T) {
	server := NewServer(nil, loopOrgService{})

	response, err := server.CompileLoopProgram(
		loopContext(), connect.NewRequest(&goalloopv1.CompileLoopProgramRequest{
			OrgSlug: "acme",
			Source:  loopSource,
		}),
	)

	require.NoError(t, err)
	require.NotNil(t, response.Msg.Program)
	require.Empty(t, response.Msg.Diagnostics)
}

func TestRunLoopProgramRequiresResourceApplyBeforeExplicitStart(t *testing.T) {
	service := &loopGoalLoopService{}
	server := NewServer(service, loopOrgService{})

	_, err := server.RunLoopProgram(
		loopContext(), connect.NewRequest(&goalloopv1.RunLoopProgramRequest{
			OrgSlug:              "acme",
			Source:               loopSource,
			WorkerSpecSnapshotId: 42,
		}),
	)

	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	require.Contains(t, err.Error(), "validate-plan-apply")
	require.Contains(t, err.Error(), "Start")
	require.Zero(t, service.createCalls)
	require.Zero(t, service.startCalls)
}

func TestRunLoopProgramRejectsInvalidSourceWithoutCreating(t *testing.T) {
	service := &loopGoalLoopService{}
	server := NewServer(service, loopOrgService{})

	_, err := server.RunLoopProgram(
		loopContext(), connect.NewRequest(&goalloopv1.RunLoopProgramRequest{
			OrgSlug:              "acme",
			Source:               "loop broken {}",
			WorkerSpecSnapshotId: 42,
		}),
	)

	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	require.Contains(t, err.Error(), "validate-plan-apply")
	require.Zero(t, service.createCalls)
	require.Zero(t, service.startCalls)
}

func TestRunLoopProgramRequiresWorkerSnapshot(t *testing.T) {
	service := &loopGoalLoopService{}
	server := NewServer(service, loopOrgService{})

	_, err := server.RunLoopProgram(
		loopContext(), connect.NewRequest(&goalloopv1.RunLoopProgramRequest{
			OrgSlug: "acme",
			Source:  loopSource,
		}),
	)

	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	require.Contains(t, err.Error(), "validate-plan-apply")
	require.Contains(t, err.Error(), "Start")
	require.Zero(t, service.validateCalls)
	require.Zero(t, service.createCalls)
	require.Zero(t, service.startCalls)
}

func TestToProtoRejectsNilGoalLoop(t *testing.T) {
	item, err := toProto(nil)

	require.Nil(t, item)
	require.Error(t, err)
}

type loopGoalLoopService struct {
	GoalLoopService
	createReq           goalloopsvc.CreateRequest
	createErr           error
	createCalls         int
	startedSlug         string
	startCalls          int
	created             *domain.GoalLoop
	startErr            error
	validateErr         error
	readyErr            error
	validateCalls       int
	validatedSnapshotID int64
}

func (s *loopGoalLoopService) ValidateExecutionReady() error {
	return s.readyErr
}

func (s *loopGoalLoopService) ValidateWorkerSnapshotForExecution(
	_ context.Context, _, _, snapshotID int64,
) error {
	s.validateCalls++
	s.validatedSnapshotID = snapshotID
	return s.validateErr
}

func (s *loopGoalLoopService) Create(_ context.Context, req goalloopsvc.CreateRequest) (*domain.GoalLoop, error) {
	s.createCalls++
	s.createReq = req
	if s.createErr != nil {
		return nil, s.createErr
	}
	criteria, _ := json.Marshal(req.AcceptanceCriteria)
	now := time.Now()
	slug := req.Slug
	if slug == "" {
		slug = req.Name
	}
	s.created = &domain.GoalLoop{
		ID: 1, OrganizationID: req.OrganizationID, CreatedByID: req.CreatedByID,
		Slug: slug, Name: req.Name, WorkerSpecSnapshotID: req.WorkerSpecSnapshotID,
		Objective: req.Objective, AcceptanceCriteria: criteria,
		VerificationCommand: req.VerificationCommand, Status: domain.StatusDraft,
		MaxIterations: req.MaxIterations, TokenBudget: req.TokenBudget,
		TimeoutMinutes: req.TimeoutMinutes, NoProgressLimit: req.NoProgressLimit,
		SameErrorLimit: req.SameErrorLimit, EscalationPolicy: req.EscalationPolicy,
		CreatedAt: now, UpdatedAt: now,
	}
	return s.created, nil
}

func (s *loopGoalLoopService) Start(
	_ context.Context, _, _ int64, slug string,
) (*domain.GoalLoop, error) {
	s.startCalls++
	s.startedSlug = slug
	if s.startErr != nil {
		return nil, s.startErr
	}
	s.created.Status = domain.StatusActive
	return s.created, nil
}

type loopOrg struct{}

func (loopOrg) GetID() int64    { return 7 }
func (loopOrg) GetSlug() string { return "acme" }
func (loopOrg) GetName() string { return "Acme" }

type loopOrgService struct{}

func (loopOrgService) GetBySlug(context.Context, string) (middleware.OrganizationGetter, error) {
	return loopOrg{}, nil
}
func (loopOrgService) IsMember(context.Context, int64, int64) (bool, error) { return true, nil }
func (loopOrgService) GetMemberRole(context.Context, int64, int64) (string, error) {
	return "owner", nil
}

func loopContext() context.Context {
	return middleware.SetTenant(context.Background(), &middleware.TenantContext{UserID: 9})
}
