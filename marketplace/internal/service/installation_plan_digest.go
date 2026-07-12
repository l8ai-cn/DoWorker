package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func installationPlanDigest(
	source InstallSource,
	command CreateInstallationPlanCommand,
) (string, error) {
	payload := struct {
		MarketplaceID        int64           `json:"marketplace_id"`
		ListingID            int64           `json:"listing_id"`
		ListingVersionID     int64           `json:"listing_version_id"`
		TargetOrganizationID int64           `json:"target_organization_id"`
		ContentDigest        string          `json:"content_digest"`
		PlatformResourceType string          `json:"platform_resource_type"`
		PlatformResourceID   int64           `json:"platform_resource_id"`
		RuntimeSnapshot      json.RawMessage `json:"runtime_snapshot"`
		QuotaPlanID          int64           `json:"quota_plan_id"`
		QuotaAccountID       string          `json:"quota_account_id"`
		EstimatedCredits     int64           `json:"estimated_credits"`
		Permissions          json.RawMessage `json:"permissions"`
		Manifest             json.RawMessage `json:"manifest"`
		Configuration        json.RawMessage `json:"configuration"`
	}{
		MarketplaceID:        source.MarketplaceID,
		ListingID:            source.ListingID,
		ListingVersionID:     source.ListingVersionID,
		TargetOrganizationID: command.TargetOrganizationID,
		ContentDigest:        source.ContentDigest,
		PlatformResourceType: source.PlatformResourceType,
		PlatformResourceID:   source.PlatformResourceID,
		RuntimeSnapshot:      source.RuntimeSnapshot,
		QuotaPlanID:          source.QuotaPlanID,
		QuotaAccountID:       source.QuotaAccountID,
		EstimatedCredits:     source.EstimatedCredits,
		Permissions:          source.Permissions,
		Manifest:             source.Manifest,
		Configuration:        command.RequestedConfiguration,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", ErrInvalidInstallationRequest
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}
