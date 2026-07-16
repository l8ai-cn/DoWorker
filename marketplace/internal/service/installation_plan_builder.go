package service

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

func buildInstallationPlan(
	source InstallSource,
	command CreateInstallationPlanCommand,
	digest string,
	now time.Time,
) (InstallationPlanRecord, InstallationPlanResult, error) {
	installationID := uuid.NewString()
	entitlementID := uuid.NewString()
	operationID := uuid.NewString()
	planID := uuid.NewString()
	expiresAt := now.Add(15 * time.Minute)
	plan, err := json.Marshal(struct {
		PlanID               string          `json:"plan_id"`
		PlanDigest           string          `json:"plan_digest"`
		ExpiresAt            time.Time       `json:"expires_at"`
		ListingVersionID     int64           `json:"listing_version_id"`
		TargetOrganizationID int64           `json:"target_platform_organization_id"`
		PlatformResourceType string          `json:"platform_resource_type"`
		PlatformResourceID   int64           `json:"platform_resource_id"`
		SourceReleaseID      int64           `json:"source_release_id"`
		RuntimeSnapshot      json.RawMessage `json:"runtime_snapshot"`
		RequiredPermissions  json.RawMessage `json:"required_permissions"`
		Configuration        json.RawMessage `json:"requested_configuration"`
		EstimatedCredits     int64           `json:"estimated_credits_micro"`
	}{
		PlanID:               planID,
		PlanDigest:           digest,
		ExpiresAt:            expiresAt,
		ListingVersionID:     source.ListingVersionID,
		TargetOrganizationID: command.TargetOrganizationID,
		PlatformResourceType: source.PlatformResourceType,
		PlatformResourceID:   source.PlatformResourceID,
		SourceReleaseID:      source.SourceReleaseID,
		RuntimeSnapshot:      source.RuntimeSnapshot,
		RequiredPermissions:  source.Permissions,
		Configuration:        command.RequestedConfiguration,
		EstimatedCredits:     source.EstimatedCredits,
	})
	if err != nil {
		return InstallationPlanRecord{}, InstallationPlanResult{}, err
	}
	record := InstallationPlanRecord{
		InstallationID: installationID, EntitlementID: entitlementID,
		OperationID: operationID, PlanID: planID, PlanDigest: digest,
		MarketplaceID: source.MarketplaceID, ListingID: source.ListingID,
		ListingVersionID:     source.ListingVersionID,
		TargetOrganizationID: command.TargetOrganizationID,
		ActorUserID:          command.ActorUserID, QuotaAccountID: source.QuotaAccountID,
		EstimatedCredits: source.EstimatedCredits,
		Configuration:    command.RequestedConfiguration,
		Plan:             plan, ExpiresAt: expiresAt,
	}
	result := InstallationPlanResult{
		InstallationID: installationID, OperationID: operationID,
		PlanID: planID, PlanDigest: digest,
		ListingVersionID: source.ListingVersionID,
		EstimatedCredits: source.EstimatedCredits,
		ExpiresAt:        expiresAt, Permissions: source.Permissions,
	}
	return record, result, nil
}
