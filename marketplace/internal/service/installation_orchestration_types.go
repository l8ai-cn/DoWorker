package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrInvalidInstallationRequest  = errors.New("invalid installation request")
	ErrApprovalRequired            = errors.New("listing requires approval")
	ErrGrantRequired               = errors.New("listing requires a grant")
	ErrQuotaAccountNotFound        = errors.New("marketplace quota account not found")
	ErrQuotaInsufficient           = errors.New("marketplace quota insufficient")
	ErrPlanExpired                 = errors.New("installation plan expired")
	ErrPlanMismatch                = errors.New("installation plan mismatch")
	ErrOperationNotFound           = errors.New("installation operation not found")
	ErrTargetOrganizationForbidden = errors.New("target organization forbidden")
	ErrRuntimeAuthorizationFailed  = errors.New("runtime authorization failed")
	ErrRuntimeInstallationRejected = errors.New("runtime installation rejected")
	ErrRuntimeInstallationUnknown  = errors.New("runtime installation outcome unknown")
)

type InstallSource struct {
	MarketplaceID        int64
	ListingID            int64
	ListingVersionID     int64
	AccessMode           string
	ContentDigest        string
	Permissions          json.RawMessage
	Manifest             json.RawMessage
	PlatformResourceType string
	PlatformResourceID   int64
	RuntimeSnapshot      json.RawMessage
	QuotaPlanID          int64
	QuotaChargeScope     string
	QuotaAccountID       string
	EstimatedCredits     int64
}

type CreateInstallationPlanCommand struct {
	MarketSlug             string
	ListingSlug            string
	ListingVersionID       int64
	TargetOrganizationID   int64
	ActorUserID            int64
	RequestedConfiguration json.RawMessage
}

type InstallationPlanRecord struct {
	InstallationID       string
	EntitlementID        string
	OperationID          string
	PlanID               string
	PlanDigest           string
	MarketplaceID        int64
	ListingID            int64
	ListingVersionID     int64
	TargetOrganizationID int64
	ActorUserID          int64
	QuotaAccountID       string
	EstimatedCredits     int64
	Configuration        json.RawMessage
	Plan                 json.RawMessage
	ExpiresAt            time.Time
}

type InstallationPlanResult struct {
	InstallationID   string
	OperationID      string
	PlanID           string
	PlanDigest       string
	ListingVersionID int64
	EstimatedCredits int64
	ExpiresAt        time.Time
	Permissions      json.RawMessage
}

type ApplyInstallationCommand struct {
	OperationID    string
	PlanID         string
	PlanDigest     string
	IdempotencyKey string
	ActorUserID    int64
}

type ApplyStatus string

const (
	ApplyPlanned   ApplyStatus = "planned"
	ApplyRunning   ApplyStatus = "running"
	ApplySucceeded ApplyStatus = "succeeded"
	ApplyFailed    ApplyStatus = "failed"
)

type ApplyExecution struct {
	InstallationID       string
	OperationID          string
	ListingVersionID     int64
	TargetOrganizationID int64
	PlatformResourceType string
	RuntimeSnapshot      json.RawMessage
	ActorUserID          int64
	Configuration        json.RawMessage
	ReservedCredits      int64
}

type ApplyResult struct {
	InstallationID string
	OperationID    string
	Status         ApplyStatus
	Stage          string
	RuntimeRef     string
	ErrorCode      string
	ErrorMessage   string
}

type RuntimeInstallRequest struct {
	InstallationID       string
	ListingVersionID     int64
	TargetOrganizationID int64
	PlatformResourceType string
	RuntimeSnapshot      json.RawMessage
	ActorUserID          int64
	Configuration        json.RawMessage
}

type RuntimeInstallResult struct {
	RuntimeRef string
	Result     json.RawMessage
}

type RuntimeBridge interface {
	Authorize(context.Context, int64, int64) error
	Install(context.Context, RuntimeInstallRequest) (RuntimeInstallResult, error)
}

type InstallationRepository interface {
	ResolveInstallSource(context.Context, string, string, int64, int64) (InstallSource, error)
	CreateDirectPlan(context.Context, InstallationPlanRecord) error
	BeginApply(context.Context, ApplyInstallationCommand) (ApplyExecution, bool, error)
	CompleteApply(context.Context, ApplyExecution, RuntimeInstallResult) (ApplyResult, error)
	FailApply(context.Context, ApplyExecution, error) (ApplyResult, error)
	GetApplyResult(context.Context, string, int64) (ApplyResult, error)
}
