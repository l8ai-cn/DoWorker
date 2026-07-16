package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateInstallationPlanIsStableAndDirect(t *testing.T) {
	repository := &installationRepositoryStub{
		source: InstallSource{
			MarketplaceID: 42, ListingID: 108, ListingVersionID: 301,
			AccessMode: "direct", ContentDigest: "content-digest",
			Permissions:          json.RawMessage(`["repository.write"]`),
			Manifest:             json.RawMessage(`{"single_instance":true}`),
			PlatformResourceType: "expert",
			RuntimeSnapshot:      json.RawMessage(`{"market_application_slug":"software-delivery-expert"}`),
			QuotaPlanID:          71,
			QuotaAccountID:       "quota-1", EstimatedCredits: 20_000_000,
		},
	}
	clock := func() time.Time { return time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC) }
	orchestration := NewInstallationOrchestrationService(repository, &runtimeBridgeStub{}, clock)
	command := CreateInstallationPlanCommand{
		MarketSlug: "commerce-market", ListingSlug: "listing-optimizer",
		ListingVersionID: 301, TargetOrganizationID: 9,
		ActorUserID:            14,
		RequestedConfiguration: json.RawMessage(`{"model_resource_id":"18"}`),
	}

	first, err := orchestration.CreatePlan(context.Background(), command)
	require.NoError(t, err)
	second, err := orchestration.CreatePlan(context.Background(), command)
	require.NoError(t, err)

	require.Equal(t, first.PlanDigest, second.PlanDigest)
	require.NotEqual(t, first.InstallationID, second.InstallationID)
	require.Equal(t, int64(20_000_000), first.EstimatedCredits)
	require.Equal(t, clock().Add(15*time.Minute), first.ExpiresAt)
}

func TestCreateInstallationPlanDoesNotFakeApproval(t *testing.T) {
	repository := &installationRepositoryStub{
		source: InstallSource{AccessMode: "approval"},
	}
	orchestration := NewInstallationOrchestrationService(
		repository,
		&runtimeBridgeStub{},
		time.Now,
	)

	_, err := orchestration.CreatePlan(context.Background(), CreateInstallationPlanCommand{
		MarketSlug: "commerce-market", ListingSlug: "listing-optimizer",
		ListingVersionID: 301, TargetOrganizationID: 9,
		ActorUserID: 14,
	})
	require.ErrorIs(t, err, ErrApprovalRequired)
	require.Zero(t, repository.createdPlans)
}

func TestCreateInstallationPlanRequiresTargetOrganizationMembership(t *testing.T) {
	repository := &installationRepositoryStub{}
	runtime := &runtimeBridgeStub{authorizeErr: ErrTargetOrganizationForbidden}
	orchestration := NewInstallationOrchestrationService(repository, runtime, time.Now)

	_, err := orchestration.CreatePlan(context.Background(), CreateInstallationPlanCommand{
		MarketSlug: "commerce-market", ListingSlug: "listing-optimizer",
		ListingVersionID: 301, TargetOrganizationID: 9, ActorUserID: 14,
	})
	require.ErrorIs(t, err, ErrTargetOrganizationForbidden)
	require.Zero(t, repository.createdPlans)
}

func TestApplyPlanIsIdempotent(t *testing.T) {
	repository := &installationRepositoryStub{}
	runtime := &runtimeBridgeStub{result: RuntimeInstallResult{RuntimeRef: "expert-18"}}
	orchestration := NewInstallationOrchestrationService(repository, runtime, time.Now)
	command := ApplyInstallationCommand{
		OperationID:    "11111111-1111-4111-8111-111111111111",
		PlanID:         "22222222-2222-4222-8222-222222222222",
		PlanDigest:     strings.Repeat("a", 64),
		IdempotencyKey: "33333333-3333-4333-8333-333333333333", ActorUserID: 14,
	}

	first, err := orchestration.Apply(context.Background(), command)
	require.NoError(t, err)
	second, err := orchestration.Apply(context.Background(), command)
	require.NoError(t, err)

	require.Equal(t, ApplySucceeded, first.Status)
	require.Equal(t, first, second)
	require.Equal(t, 1, runtime.calls)
	require.Equal(t, 1, repository.completed)
	require.JSONEq(t, `{"market_application_slug":"software-delivery-expert"}`,
		string(runtime.lastRequest.RuntimeSnapshot))
	require.Equal(t, int64(9), runtime.lastRequest.TargetOrganizationID)
	require.Equal(t, int64(14), runtime.lastRequest.ActorUserID)
	require.Equal(t, int64(101), runtime.lastRequest.PlatformResourceID)
	require.Equal(t, int64(201), runtime.lastRequest.SourceReleaseID)
	require.Equal(t, int64(14), repository.resultActorUserID)
	require.Equal(t, 1, runtime.authorizationCalls)
}

func TestApplyPlanRechecksTargetOrganizationMembership(t *testing.T) {
	repository := &installationRepositoryStub{}
	runtime := &runtimeBridgeStub{authorizeErr: ErrTargetOrganizationForbidden}
	orchestration := NewInstallationOrchestrationService(repository, runtime, time.Now)

	result, err := orchestration.Apply(context.Background(), ApplyInstallationCommand{
		OperationID:    "11111111-1111-4111-8111-111111111111",
		PlanID:         "22222222-2222-4222-8222-222222222222",
		PlanDigest:     strings.Repeat("a", 64),
		IdempotencyKey: "33333333-3333-4333-8333-333333333333",
		ActorUserID:    14,
	})

	require.ErrorIs(t, err, ErrTargetOrganizationForbidden)
	require.Equal(t, ApplyFailed, result.Status)
	require.Zero(t, runtime.calls)
	require.Equal(t, 1, repository.failed)
}

func TestApplyPlanStopsBeforeRuntimeWhenQuotaIsInsufficient(t *testing.T) {
	repository := &installationRepositoryStub{beginErr: ErrQuotaInsufficient}
	runtime := &runtimeBridgeStub{}
	orchestration := NewInstallationOrchestrationService(repository, runtime, time.Now)

	_, err := orchestration.Apply(context.Background(), ApplyInstallationCommand{
		OperationID:    "11111111-1111-4111-8111-111111111111",
		PlanID:         "22222222-2222-4222-8222-222222222222",
		PlanDigest:     strings.Repeat("a", 64),
		IdempotencyKey: "33333333-3333-4333-8333-333333333333", ActorUserID: 14,
	})
	require.ErrorIs(t, err, ErrQuotaInsufficient)
	require.Zero(t, runtime.calls)
}

func TestApplyPlanKeepsIndeterminateRuntimeOutcomeRecoverable(t *testing.T) {
	repository := &installationRepositoryStub{}
	runtime := &runtimeBridgeStub{err: ErrRuntimeInstallationUnknown}
	orchestration := NewInstallationOrchestrationService(repository, runtime, time.Now)

	result, err := orchestration.Apply(context.Background(), ApplyInstallationCommand{
		OperationID:    "11111111-1111-4111-8111-111111111111",
		PlanID:         "22222222-2222-4222-8222-222222222222",
		PlanDigest:     strings.Repeat("a", 64),
		IdempotencyKey: "33333333-3333-4333-8333-333333333333",
		ActorUserID:    14,
	})

	require.ErrorIs(t, err, ErrRuntimeInstallationUnknown)
	require.Equal(t, ApplyRunning, result.Status)
	require.Zero(t, repository.failed)
}

func TestApplyPlanReturnsConcurrentSuccessTerminalState(t *testing.T) {
	completeErr := errors.New("reservation already settled")
	repository := &installationRepositoryStub{
		completeErr: completeErr,
		result: ApplyResult{
			InstallationID: "installation-1",
			OperationID:    "11111111-1111-4111-8111-111111111111",
			Status:         ApplySucceeded,
			Stage:          "settle",
			RuntimeRef:     "expert-18",
		},
	}
	runtime := &runtimeBridgeStub{result: RuntimeInstallResult{RuntimeRef: "expert-18"}}
	orchestration := NewInstallationOrchestrationService(repository, runtime, time.Now)

	result, err := orchestration.Apply(context.Background(), validApplyCommand())

	require.NoError(t, err)
	require.Equal(t, ApplySucceeded, result.Status)
	require.Equal(t, "expert-18", result.RuntimeRef)
}

func TestApplyPlanReturnsConcurrentFailureTerminalState(t *testing.T) {
	repository := &installationRepositoryStub{
		failErr: errors.New("reservation already released"),
		result: ApplyResult{
			InstallationID: "installation-1",
			OperationID:    "11111111-1111-4111-8111-111111111111",
			Status:         ApplyFailed,
			Stage:          "runtime",
			ErrorCode:      "RUNTIME_INSTALL_FAILED",
		},
	}
	runtime := &runtimeBridgeStub{err: ErrRuntimeInstallationRejected}
	orchestration := NewInstallationOrchestrationService(repository, runtime, time.Now)

	result, err := orchestration.Apply(context.Background(), validApplyCommand())

	require.ErrorIs(t, err, ErrRuntimeInstallationRejected)
	require.Equal(t, ApplyFailed, result.Status)
	require.Equal(t, "RUNTIME_INSTALL_FAILED", result.ErrorCode)
}

func validApplyCommand() ApplyInstallationCommand {
	return ApplyInstallationCommand{
		OperationID:    "11111111-1111-4111-8111-111111111111",
		PlanID:         "22222222-2222-4222-8222-222222222222",
		PlanDigest:     strings.Repeat("a", 64),
		IdempotencyKey: "33333333-3333-4333-8333-333333333333",
		ActorUserID:    14,
	}
}

type installationRepositoryStub struct {
	source            InstallSource
	beginErr          error
	completeErr       error
	failErr           error
	beginExisting     bool
	createdPlans      int
	completed         int
	failed            int
	result            ApplyResult
	resultActorUserID int64
}

func (r *installationRepositoryStub) ResolveInstallSource(
	context.Context,
	string,
	string,
	int64,
	int64,
) (InstallSource, error) {
	return r.source, nil
}

func (r *installationRepositoryStub) CreateDirectPlan(
	_ context.Context,
	record InstallationPlanRecord,
) error {
	r.createdPlans++
	r.result = ApplyResult{
		InstallationID: record.InstallationID,
		OperationID:    record.OperationID,
		Status:         ApplyPlanned,
		Stage:          "entitlement",
	}
	return nil
}

func (r *installationRepositoryStub) BeginApply(
	_ context.Context,
	command ApplyInstallationCommand,
) (ApplyExecution, bool, error) {
	if r.beginErr != nil {
		return ApplyExecution{}, false, r.beginErr
	}
	if r.beginExisting {
		return ApplyExecution{}, true, nil
	}
	return ApplyExecution{
		InstallationID: "installation-1", OperationID: command.OperationID,
		ListingVersionID: 301, TargetOrganizationID: 9,
		PlatformResourceType: "expert",
		PlatformResourceID:   101,
		SourceReleaseID:      201,
		RuntimeSnapshot:      json.RawMessage(`{"market_application_slug":"software-delivery-expert"}`),
		ActorUserID:          command.ActorUserID,
		Configuration:        json.RawMessage(`{}`), ReservedCredits: 20_000_000,
	}, false, nil
}

func (r *installationRepositoryStub) CompleteApply(
	_ context.Context,
	execution ApplyExecution,
	result RuntimeInstallResult,
) (ApplyResult, error) {
	r.completed++
	if r.completeErr != nil {
		return ApplyResult{}, r.completeErr
	}
	r.result = ApplyResult{
		InstallationID: execution.InstallationID,
		OperationID:    execution.OperationID,
		Status:         ApplySucceeded, Stage: "settle", RuntimeRef: result.RuntimeRef,
	}
	r.beginExisting = true
	return r.result, nil
}

func (r *installationRepositoryStub) FailApply(
	context.Context,
	ApplyExecution,
	error,
) (ApplyResult, error) {
	r.failed++
	if r.failErr != nil {
		return ApplyResult{}, r.failErr
	}
	return ApplyResult{Status: ApplyFailed}, nil
}

func (r *installationRepositoryStub) GetApplyResult(
	_ context.Context,
	_ string,
	actorUserID int64,
) (ApplyResult, error) {
	r.resultActorUserID = actorUserID
	return r.result, nil
}

type runtimeBridgeStub struct {
	result             RuntimeInstallResult
	err                error
	calls              int
	lastRequest        RuntimeInstallRequest
	authorizeErr       error
	authorizationCalls int
}

func (r *runtimeBridgeStub) Authorize(context.Context, int64, int64) error {
	r.authorizationCalls++
	return r.authorizeErr
}

func (r *runtimeBridgeStub) Install(
	_ context.Context,
	request RuntimeInstallRequest,
) (RuntimeInstallResult, error) {
	r.calls++
	r.lastRequest = request
	return r.result, r.err
}
