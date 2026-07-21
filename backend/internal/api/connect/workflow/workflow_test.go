package workflowconnect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	workflowv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/workflow/v1"
)

func connectCodeOf(t *testing.T, err error) connect.Code {
	t.Helper()
	var ce *connect.Error
	require.True(t, errors.As(err, &ce), "expected *connect.Error, got %v", err)
	return ce.Code()
}

func TestListWorkflows_NoOrgSlug(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)
	_, err := srv.ListWorkflows(context.Background(), connect.NewRequest(&workflowv1.ListWorkflowsRequest{}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connectCodeOf(t, err))
}

func TestGetWorkflow_NoOrgSlug(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)
	_, err := srv.GetWorkflow(context.Background(), connect.NewRequest(&workflowv1.GetWorkflowRequest{WorkflowSlug: "s"}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connectCodeOf(t, err))
}

func TestCreateWorkflow_NoOrgSlug(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)
	_, err := srv.CreateWorkflow(context.Background(), connect.NewRequest(&workflowv1.CreateWorkflowRequest{Name: "n"}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connectCodeOf(t, err))
}

func TestCreateWorkflowRequiresResourceApply(t *testing.T) {
	srv := NewServer(nil, nil, nil, loopOrgService{}, nil)
	_, err := srv.CreateWorkflow(
		loopContext(),
		connect.NewRequest(&workflowv1.CreateWorkflowRequest{
			OrgSlug: "acme",
			Name:    "legacy",
		}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeFailedPrecondition, connectCodeOf(t, err))
	assert.Contains(t, err.Error(), "validate-plan-apply")
}

func TestUpdateWorkflow_NoOrgSlug(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)
	_, err := srv.UpdateWorkflow(context.Background(), connect.NewRequest(&workflowv1.UpdateWorkflowRequest{WorkflowSlug: "s"}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connectCodeOf(t, err))
}

func TestDeleteWorkflow_NoOrgSlug(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)
	_, err := srv.DeleteWorkflow(context.Background(), connect.NewRequest(&workflowv1.DeleteWorkflowRequest{WorkflowSlug: "s"}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connectCodeOf(t, err))
}

func TestEnableWorkflow_NoOrgSlug(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)
	_, err := srv.EnableWorkflow(context.Background(), connect.NewRequest(&workflowv1.WorkflowActionRequest{WorkflowSlug: "s"}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connectCodeOf(t, err))
}

func TestDisableWorkflow_NoOrgSlug(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)
	_, err := srv.DisableWorkflow(context.Background(), connect.NewRequest(&workflowv1.WorkflowActionRequest{WorkflowSlug: "s"}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connectCodeOf(t, err))
}

func TestTriggerWorkflow_NoOrgSlug(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)
	_, err := srv.TriggerWorkflow(context.Background(), connect.NewRequest(&workflowv1.TriggerWorkflowRequest{WorkflowSlug: "s"}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connectCodeOf(t, err))
}

func TestListWorkflowRuns_NoOrgSlug(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)
	_, err := srv.ListWorkflowRuns(context.Background(), connect.NewRequest(&workflowv1.ListWorkflowRunsRequest{WorkflowSlug: "s"}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connectCodeOf(t, err))
}

func TestCancelWorkflowRun_NoOrgSlug(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)
	_, err := srv.CancelWorkflowRun(context.Background(), connect.NewRequest(&workflowv1.CancelWorkflowRunRequest{WorkflowSlug: "s", RunId: 1}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connectCodeOf(t, err))
}

func TestProcedureConstants(t *testing.T) {
	cases := map[string]string{
		"/proto.workflow.v1.WorkflowService/ListWorkflows":     ListWorkflowsProcedure,
		"/proto.workflow.v1.WorkflowService/GetWorkflow":       GetWorkflowProcedure,
		"/proto.workflow.v1.WorkflowService/CreateWorkflow":    CreateWorkflowProcedure,
		"/proto.workflow.v1.WorkflowService/UpdateWorkflow":    UpdateWorkflowProcedure,
		"/proto.workflow.v1.WorkflowService/DeleteWorkflow":    DeleteWorkflowProcedure,
		"/proto.workflow.v1.WorkflowService/EnableWorkflow":    EnableWorkflowProcedure,
		"/proto.workflow.v1.WorkflowService/DisableWorkflow":   DisableWorkflowProcedure,
		"/proto.workflow.v1.WorkflowService/TriggerWorkflow":   TriggerWorkflowProcedure,
		"/proto.workflow.v1.WorkflowService/ListWorkflowRuns":  ListWorkflowRunsProcedure,
		"/proto.workflow.v1.WorkflowService/CancelWorkflowRun": CancelWorkflowRunProcedure,
	}
	for want, got := range cases {
		assert.Equal(t, want, got)
	}
}
