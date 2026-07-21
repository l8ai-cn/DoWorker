package orchestrationresourceconnect

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationworker"
	resourcev1 "github.com/l8ai-cn/agentcloud/proto/gen/go/orchestration_resource/v1"
)

func TestApplyBindingResourcePlanUsesActorBoundScope(t *testing.T) {
	binding := &bindingApplyStub{result: testHead()}
	server := NewServer(
		&serviceStub{},
		binding,
		&workerTemplateApplyStub{},
		&workerApplyStub{},
		&promptApplyStub{},
		&expertApplyStub{},
		&workflowApplyStub{},
		&goalLoopApplyStub{},
		testOrganizations(),
	)

	response, err := server.ApplyBindingResourcePlan(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ApplyBindingResourcePlanRequest{
			OrgSlug: "acme",
			PlanId:  testPlanID,
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, testScope(), binding.scope)
	assert.Equal(t, testPlanID, binding.planID)
	assert.Equal(t, testResourceID, response.Msg.Identity.Uid)
}

func TestApplyWorkerTemplatePlanReturnsImmutableSnapshotBinding(t *testing.T) {
	worker := &workerTemplateApplyStub{
		result: workerplanner.AppliedWorkerTemplate{
			Head:                 testHead(),
			WorkerSpecSnapshotID: 91,
		},
	}
	server := NewServer(
		&serviceStub{},
		&bindingApplyStub{},
		worker,
		&workerApplyStub{},
		&promptApplyStub{},
		&expertApplyStub{},
		&workflowApplyStub{},
		&goalLoopApplyStub{},
		testOrganizations(),
	)

	response, err := server.ApplyWorkerTemplatePlan(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ApplyWorkerTemplatePlanRequest{
			OrgSlug: "acme",
			PlanId:  testPlanID,
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, testScope(), worker.scope)
	assert.Equal(t, testPlanID, worker.planID)
	assert.EqualValues(t, 91, response.Msg.WorkerSpecSnapshotId)
	assert.Equal(t, testResourceID, response.Msg.Resource.Identity.Uid)
}

func TestCreateWorkerFromPlanReturnsPinnedLaunchProjection(t *testing.T) {
	worker := &workerApplyStub{
		result: workerplanner.AppliedWorker{
			Head:                 testHead(),
			LaunchID:             71,
			PodID:                73,
			PodKey:               "7-standalone-12345678",
			WorkerSpecSnapshotID: 91,
			ResourceRevision:     3,
			RunnerID:             11,
		},
	}
	server := &Server{
		service:     &serviceStub{},
		workerApply: worker,
		orgs:        testOrganizations(),
	}

	response, err := server.CreateWorkerFromPlan(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.CreateWorkerFromPlanRequest{
			OrgSlug: "acme",
			PlanId:  testPlanID,
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, testScope(), worker.scope)
	assert.Equal(t, testPlanID, worker.planID)
	assert.EqualValues(t, 71, response.Msg.LaunchId)
	assert.EqualValues(t, 73, response.Msg.PodId)
	assert.Equal(t, "7-standalone-12345678", response.Msg.PodKey)
	assert.EqualValues(t, 91, response.Msg.WorkerSpecSnapshotId)
	assert.EqualValues(t, 3, response.Msg.ResourceRevision)
	assert.EqualValues(t, 11, response.Msg.RunnerId)
	assert.Equal(t, testResourceID, response.Msg.Resource.Identity.Uid)
}

func TestApplyPromptPlanUsesActorBoundScope(t *testing.T) {
	prompt := &promptApplyStub{result: testHead()}
	server := NewServer(
		&serviceStub{},
		&bindingApplyStub{},
		&workerTemplateApplyStub{},
		&workerApplyStub{},
		prompt,
		&expertApplyStub{},
		&workflowApplyStub{},
		&goalLoopApplyStub{},
		testOrganizations(),
	)

	response, err := server.ApplyPromptPlan(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ApplyPromptPlanRequest{
			OrgSlug: "acme",
			PlanId:  testPlanID,
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, testScope(), prompt.scope)
	assert.Equal(t, testPlanID, prompt.planID)
	assert.Equal(t, testResourceID, response.Msg.Identity.Uid)
}

func TestApplyExpertPlanReturnsPinnedDomainProjection(t *testing.T) {
	expert := &expertApplyStub{
		result: workerplanner.AppliedExpert{
			Head:                 testHead(),
			ExpertID:             81,
			WorkerSpecSnapshotID: 91,
			ResourceRevision:     3,
		},
	}
	server := NewServer(
		&serviceStub{},
		&bindingApplyStub{},
		&workerTemplateApplyStub{},
		&workerApplyStub{},
		&promptApplyStub{},
		expert,
		&workflowApplyStub{},
		&goalLoopApplyStub{},
		testOrganizations(),
	)

	response, err := server.ApplyExpertPlan(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ApplyExpertPlanRequest{
			OrgSlug: "acme",
			PlanId:  testPlanID,
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, testScope(), expert.scope)
	assert.Equal(t, testPlanID, expert.planID)
	assert.EqualValues(t, 81, response.Msg.ExpertId)
	assert.EqualValues(t, 91, response.Msg.WorkerSpecSnapshotId)
	assert.EqualValues(t, 3, response.Msg.ResourceRevision)
	assert.Equal(t, testResourceID, response.Msg.Resource.Identity.Uid)
}

func TestApplyWorkflowPlanReturnsPinnedDomainProjection(t *testing.T) {
	workflow := &workflowApplyStub{
		result: workerplanner.AppliedWorkflow{
			Head:                 testHead(),
			WorkflowID:           82,
			WorkerSpecSnapshotID: 92,
			ResourceRevision:     4,
		},
	}
	server := NewServer(
		&serviceStub{},
		&bindingApplyStub{},
		&workerTemplateApplyStub{},
		&workerApplyStub{},
		&promptApplyStub{},
		&expertApplyStub{},
		workflow,
		&goalLoopApplyStub{},
		testOrganizations(),
	)

	response, err := server.ApplyWorkflowPlan(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ApplyWorkflowPlanRequest{
			OrgSlug: "acme",
			PlanId:  testPlanID,
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, testScope(), workflow.scope)
	assert.Equal(t, testPlanID, workflow.planID)
	assert.EqualValues(t, 82, response.Msg.WorkflowId)
	assert.EqualValues(t, 92, response.Msg.WorkerSpecSnapshotId)
	assert.EqualValues(t, 4, response.Msg.ResourceRevision)
	assert.Equal(t, testResourceID, response.Msg.Resource.Identity.Uid)
}

func TestCreateGoalLoopFromPlanReturnsPinnedDraftProjection(t *testing.T) {
	goalLoop := &goalLoopApplyStub{
		result: workerplanner.AppliedGoalLoop{
			Head:                 testHead(),
			GoalLoopID:           83,
			WorkerSpecSnapshotID: 93,
			ResourceRevision:     5,
		},
	}
	server := &Server{
		service:       &serviceStub{},
		goalLoopApply: goalLoop,
		orgs:          testOrganizations(),
	}

	response, err := server.CreateGoalLoopFromPlan(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.CreateGoalLoopFromPlanRequest{
			OrgSlug: "acme",
			PlanId:  testPlanID,
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, testScope(), goalLoop.scope)
	assert.Equal(t, testPlanID, goalLoop.planID)
	assert.EqualValues(t, 83, response.Msg.GoalLoopId)
	assert.EqualValues(t, 93, response.Msg.WorkerSpecSnapshotId)
	assert.EqualValues(t, 5, response.Msg.ResourceRevision)
	assert.Equal(t, testResourceID, response.Msg.Resource.Identity.Uid)
}

func TestApplyRejectsInvalidPlanIDBeforeCallingTargetService(t *testing.T) {
	binding := &bindingApplyStub{}
	server := NewServer(
		&serviceStub{},
		binding,
		&workerTemplateApplyStub{},
		&workerApplyStub{},
		&promptApplyStub{},
		&expertApplyStub{},
		&workflowApplyStub{},
		&goalLoopApplyStub{},
		testOrganizations(),
	)

	_, err := server.ApplyBindingResourcePlan(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ApplyBindingResourcePlanRequest{
			OrgSlug: "acme",
			PlanId:  "not-a-uuid",
		}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	assert.Zero(t, binding.calls)
}

func TestApplyRechecksAuthorizationBeforeCallingTargetService(t *testing.T) {
	controlService := &serviceStub{authorizeApplyErr: controlservice.ErrForbidden}
	binding := &bindingApplyStub{}
	server := NewServer(
		controlService,
		binding,
		&workerTemplateApplyStub{},
		&workerApplyStub{},
		&promptApplyStub{},
		&expertApplyStub{},
		&workflowApplyStub{},
		&goalLoopApplyStub{},
		testOrganizations(),
	)

	_, err := server.ApplyBindingResourcePlan(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ApplyBindingResourcePlanRequest{
			OrgSlug: "acme",
			PlanId:  testPlanID,
		}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
	assert.Equal(t, testScope(), controlService.authorizeApplyScope)
	assert.Equal(t, testPlanID, controlService.authorizeApplyPlanID)
	assert.Zero(t, binding.calls)
}

func TestApplySanitizesConsumedPlanDetails(t *testing.T) {
	binding := &bindingApplyStub{
		err: errors.Join(
			control.ErrConsumed,
			errors.New(`artifact {"secret":"value"} SQLSTATE 23505`),
		),
	}
	server := NewServer(
		&serviceStub{},
		binding,
		&workerTemplateApplyStub{},
		&workerApplyStub{},
		&promptApplyStub{},
		&expertApplyStub{},
		&workflowApplyStub{},
		&goalLoopApplyStub{},
		testOrganizations(),
	)

	_, err := server.ApplyBindingResourcePlan(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ApplyBindingResourcePlanRequest{
			OrgSlug: "acme",
			PlanId:  testPlanID,
		}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeAborted, connect.CodeOf(err))
	assert.NotContains(t, err.Error(), "artifact")
	assert.NotContains(t, err.Error(), "secret")
	assert.NotContains(t, err.Error(), "SQLSTATE")
}
