package orchestrationresourceconnect

import (
	"context"
	"errors"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	service "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
)

type serviceStub struct {
	validateRequest service.ValidateRequest
	validateResult  service.ValidationResult
	validateErr     error
	validateCalls   int

	planRequest service.PlanRequest
	planResult  service.PlanResult
	planErr     error

	getScope  control.Scope
	getTarget control.ResourceTarget
	getResult control.ResourceHead
	getErr    error

	capabilitiesScope  control.Scope
	capabilitiesTarget control.ResourceTarget
	capabilitiesResult service.ResourceCapabilities
	capabilitiesErr    error

	listScope  control.Scope
	listFilter service.ResourceListFilter
	listResult service.ResourceListPage
	listErr    error

	exportRequest service.ExportResourceRequest
	exportResult  service.ResourceExport
	exportErr     error

	getPlanScope  control.Scope
	getPlanID     string
	getPlanResult control.Plan
	getPlanErr    error

	authorizeApplyScope  control.Scope
	authorizeApplyPlanID string
	authorizeApplyErr    error
	authorizeApplyCalls  int
}

type bindingApplyStub struct {
	scope  control.Scope
	planID string
	result control.ResourceHead
	err    error
	calls  int
}

func (stub *bindingApplyStub) Apply(
	_ context.Context,
	scope control.Scope,
	planID string,
) (control.ResourceHead, error) {
	stub.scope = scope
	stub.planID = planID
	stub.calls++
	return stub.result, stub.err
}

type workerTemplateApplyStub struct {
	scope  control.Scope
	planID string
	result workerplanner.AppliedWorkerTemplate
	err    error
	calls  int
}

func (stub *workerTemplateApplyStub) Apply(
	_ context.Context,
	scope control.Scope,
	planID string,
) (workerplanner.AppliedWorkerTemplate, error) {
	stub.scope = scope
	stub.planID = planID
	stub.calls++
	return stub.result, stub.err
}

type workerApplyStub struct {
	scope  control.Scope
	planID string
	result workerplanner.AppliedWorker
	err    error
	calls  int
}

func (stub *workerApplyStub) Apply(
	_ context.Context,
	scope control.Scope,
	planID string,
) (workerplanner.AppliedWorker, error) {
	stub.scope = scope
	stub.planID = planID
	stub.calls++
	return stub.result, stub.err
}

type promptApplyStub struct {
	scope  control.Scope
	planID string
	result control.ResourceHead
	err    error
	calls  int
}

func (stub *promptApplyStub) Apply(
	_ context.Context,
	scope control.Scope,
	planID string,
) (control.ResourceHead, error) {
	stub.scope = scope
	stub.planID = planID
	stub.calls++
	return stub.result, stub.err
}

type expertApplyStub struct {
	scope  control.Scope
	planID string
	result workerplanner.AppliedExpert
	err    error
	calls  int
}

func (stub *expertApplyStub) Apply(
	_ context.Context,
	scope control.Scope,
	planID string,
) (workerplanner.AppliedExpert, error) {
	stub.scope = scope
	stub.planID = planID
	stub.calls++
	return stub.result, stub.err
}

type workflowApplyStub struct {
	scope  control.Scope
	planID string
	result workerplanner.AppliedWorkflow
	err    error
	calls  int
}

func (stub *workflowApplyStub) Apply(
	_ context.Context,
	scope control.Scope,
	planID string,
) (workerplanner.AppliedWorkflow, error) {
	stub.scope = scope
	stub.planID = planID
	stub.calls++
	return stub.result, stub.err
}

type goalLoopApplyStub struct {
	scope  control.Scope
	planID string
	result workerplanner.AppliedGoalLoop
	err    error
	calls  int
}

func (stub *goalLoopApplyStub) Apply(
	_ context.Context,
	scope control.Scope,
	planID string,
) (workerplanner.AppliedGoalLoop, error) {
	stub.scope = scope
	stub.planID = planID
	stub.calls++
	return stub.result, stub.err
}

func newTestServer(
	service *serviceStub,
	organizations middleware.OrganizationService,
) *Server {
	return NewServer(
		service,
		&bindingApplyStub{},
		&workerTemplateApplyStub{},
		&workerApplyStub{},
		&promptApplyStub{},
		&expertApplyStub{},
		&workflowApplyStub{},
		&goalLoopApplyStub{},
		organizations,
	)
}

func (stub *serviceStub) Validate(
	_ context.Context,
	request service.ValidateRequest,
) (service.ValidationResult, error) {
	stub.validateCalls++
	stub.validateRequest = request
	return stub.validateResult, stub.validateErr
}

func (stub *serviceStub) Plan(
	_ context.Context,
	request service.PlanRequest,
) (service.PlanResult, error) {
	stub.planRequest = request
	return stub.planResult, stub.planErr
}

func (stub *serviceStub) GetResource(
	_ context.Context,
	scope control.Scope,
	target control.ResourceTarget,
) (control.ResourceHead, error) {
	stub.getScope = scope
	stub.getTarget = target
	return stub.getResult, stub.getErr
}

func (stub *serviceStub) GetResourceCapabilities(
	_ context.Context,
	scope control.Scope,
	target control.ResourceTarget,
) (service.ResourceCapabilities, error) {
	stub.capabilitiesScope = scope
	stub.capabilitiesTarget = target
	return stub.capabilitiesResult, stub.capabilitiesErr
}

func (stub *serviceStub) ListResources(
	_ context.Context,
	scope control.Scope,
	filter service.ResourceListFilter,
) (service.ResourceListPage, error) {
	stub.listScope = scope
	stub.listFilter = filter
	result := stub.listResult
	if result.AppliedFilter.ModelBinding == nil &&
		result.AppliedFilter.EnvironmentBundle == nil {
		result.AppliedFilter = filter
	}
	return result, stub.listErr
}

func (stub *serviceStub) ExportResource(
	_ context.Context,
	request service.ExportResourceRequest,
) (service.ResourceExport, error) {
	stub.exportRequest = request
	return stub.exportResult, stub.exportErr
}

func (stub *serviceStub) GetResourcePlan(
	_ context.Context,
	scope control.Scope,
	planID string,
) (control.Plan, error) {
	stub.getPlanScope = scope
	stub.getPlanID = planID
	return stub.getPlanResult, stub.getPlanErr
}

func (stub *serviceStub) AuthorizeApply(
	_ context.Context,
	scope control.Scope,
	planID string,
) error {
	stub.authorizeApplyCalls++
	stub.authorizeApplyScope = scope
	stub.authorizeApplyPlanID = planID
	return stub.authorizeApplyErr
}

type organizationStub struct {
	id        int64
	slug      string
	role      string
	member    bool
	getErr    error
	roleErr   error
	memberErr error
}

func (stub organizationStub) GetBySlug(
	_ context.Context,
	_ string,
) (middleware.OrganizationGetter, error) {
	if stub.getErr != nil {
		return nil, stub.getErr
	}
	return organizationRecord{id: stub.id, slug: stub.slug}, nil
}

func (stub organizationStub) IsMember(context.Context, int64, int64) (bool, error) {
	return stub.member, stub.memberErr
}

func (stub organizationStub) GetMemberRole(context.Context, int64, int64) (string, error) {
	if stub.roleErr != nil {
		return "", stub.roleErr
	}
	if !stub.member {
		return "", errors.New("not a member")
	}
	return stub.role, nil
}

type organizationRecord struct {
	id   int64
	slug string
}

func (record organizationRecord) GetID() int64    { return record.id }
func (record organizationRecord) GetSlug() string { return record.slug }
func (record organizationRecord) GetName() string { return record.slug }

func testOrganizations() organizationStub {
	return organizationStub{id: 81, slug: "acme", role: "member", member: true}
}

func authenticatedContext(userID int64) context.Context {
	return middleware.SetTenant(
		context.Background(),
		&middleware.TenantContext{UserID: userID},
	)
}
