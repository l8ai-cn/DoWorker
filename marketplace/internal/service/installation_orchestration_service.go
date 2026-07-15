package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type InstallationOrchestrationService struct {
	repository InstallationRepository
	runtime    RuntimeBridge
	now        func() time.Time
}

func NewInstallationOrchestrationService(
	repository InstallationRepository,
	runtime RuntimeBridge,
	now func() time.Time,
) *InstallationOrchestrationService {
	return &InstallationOrchestrationService{
		repository: repository,
		runtime:    runtime,
		now:        now,
	}
}

func (s *InstallationOrchestrationService) CreatePlan(
	ctx context.Context,
	command CreateInstallationPlanCommand,
) (InstallationPlanResult, error) {
	if len(command.RequestedConfiguration) == 0 {
		command.RequestedConfiguration = json.RawMessage(`{}`)
	}
	if err := validatePlanCommand(command); err != nil {
		return InstallationPlanResult{}, err
	}
	if err := s.runtime.Authorize(
		ctx,
		command.TargetOrganizationID,
		command.ActorUserID,
	); err != nil {
		return InstallationPlanResult{}, err
	}
	source, err := s.repository.ResolveInstallSource(
		ctx,
		command.MarketSlug,
		command.ListingSlug,
		command.ListingVersionID,
		command.TargetOrganizationID,
	)
	if err != nil {
		return InstallationPlanResult{}, err
	}
	switch source.AccessMode {
	case "direct":
	case "approval":
		return InstallationPlanResult{}, ErrApprovalRequired
	case "grant_only":
		return InstallationPlanResult{}, ErrGrantRequired
	default:
		return InstallationPlanResult{}, ErrInvalidInstallationRequest
	}
	if source.PlatformResourceType != "expert" ||
		(source.PlatformResourceID == 0) != (source.SourceReleaseID == 0) ||
		source.PlatformResourceID < 0 ||
		source.SourceReleaseID < 0 ||
		!json.Valid(source.RuntimeSnapshot) ||
		source.QuotaPlanID <= 0 ||
		source.EstimatedCredits <= 0 ||
		strings.TrimSpace(source.QuotaAccountID) == "" {
		return InstallationPlanResult{}, ErrInvalidInstallationRequest
	}
	digest, err := installationPlanDigest(source, command)
	if err != nil {
		return InstallationPlanResult{}, err
	}
	now := s.now().UTC()
	record, result, err := buildInstallationPlan(source, command, digest, now)
	if err != nil {
		return InstallationPlanResult{}, err
	}
	if err := s.repository.CreateDirectPlan(ctx, record); err != nil {
		return InstallationPlanResult{}, err
	}
	return result, nil
}

func (s *InstallationOrchestrationService) Apply(
	ctx context.Context,
	command ApplyInstallationCommand,
) (ApplyResult, error) {
	if uuid.Validate(command.OperationID) != nil ||
		uuid.Validate(command.PlanID) != nil ||
		uuid.Validate(command.IdempotencyKey) != nil ||
		len(command.PlanDigest) != 64 ||
		!isLowerHex(command.PlanDigest) ||
		command.ActorUserID <= 0 {
		return ApplyResult{}, ErrInvalidInstallationRequest
	}
	execution, existing, err := s.repository.BeginApply(ctx, command)
	if err != nil {
		return ApplyResult{}, err
	}
	if existing {
		return s.repository.GetApplyResult(ctx, command.OperationID, command.ActorUserID)
	}
	if authorizeErr := s.runtime.Authorize(
		ctx,
		execution.TargetOrganizationID,
		execution.ActorUserID,
	); authorizeErr != nil {
		return s.failApply(ctx, execution, authorizeErr)
	}
	runtimeResult, runtimeErr := s.runtime.Install(ctx, RuntimeInstallRequest{
		InstallationID:       execution.InstallationID,
		ListingVersionID:     execution.ListingVersionID,
		TargetOrganizationID: execution.TargetOrganizationID,
		PlatformResourceType: execution.PlatformResourceType,
		PlatformResourceID:   execution.PlatformResourceID,
		SourceReleaseID:      execution.SourceReleaseID,
		RuntimeSnapshot:      execution.RuntimeSnapshot,
		ActorUserID:          execution.ActorUserID,
		Configuration:        execution.Configuration,
	})
	if runtimeErr != nil {
		if errors.Is(runtimeErr, ErrRuntimeInstallationUnknown) {
			return ApplyResult{
				InstallationID: execution.InstallationID,
				OperationID:    execution.OperationID,
				Status:         ApplyRunning,
				Stage:          "runtime",
			}, runtimeErr
		}
		return s.failApply(ctx, execution, runtimeErr)
	}
	result, completeErr := s.repository.CompleteApply(ctx, execution, runtimeResult)
	if completeErr == nil {
		return result, nil
	}
	terminal, queryErr := s.repository.GetApplyResult(
		ctx,
		execution.OperationID,
		execution.ActorUserID,
	)
	if queryErr == nil && terminal.Status == ApplySucceeded {
		return terminal, nil
	}
	return ApplyResult{}, completeErr
}

func (s *InstallationOrchestrationService) GetOperation(
	ctx context.Context,
	operationID string,
	actorUserID int64,
) (ApplyResult, error) {
	if uuid.Validate(operationID) != nil || actorUserID <= 0 {
		return ApplyResult{}, ErrInvalidInstallationRequest
	}
	return s.repository.GetApplyResult(ctx, operationID, actorUserID)
}

func (s *InstallationOrchestrationService) failApply(
	ctx context.Context,
	execution ApplyExecution,
	cause error,
) (ApplyResult, error) {
	result, failErr := s.repository.FailApply(ctx, execution, cause)
	if failErr == nil {
		return result, cause
	}
	terminal, queryErr := s.repository.GetApplyResult(
		ctx,
		execution.OperationID,
		execution.ActorUserID,
	)
	if queryErr == nil && terminal.Status == ApplyFailed {
		return terminal, cause
	}
	return ApplyResult{}, failErr
}

func isLowerHex(value string) bool {
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func validatePlanCommand(command CreateInstallationPlanCommand) error {
	if strings.TrimSpace(command.MarketSlug) == "" ||
		strings.TrimSpace(command.ListingSlug) == "" ||
		command.ListingVersionID <= 0 ||
		command.TargetOrganizationID <= 0 ||
		command.ActorUserID <= 0 {
		return ErrInvalidInstallationRequest
	}
	if !json.Valid(command.RequestedConfiguration) {
		return ErrInvalidInstallationRequest
	}
	return nil
}
