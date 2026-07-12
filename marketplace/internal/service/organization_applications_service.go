package service

import (
	"context"
	"time"
)

type OrganizationApplication struct {
	InstallationID string    `json:"installation_id"`
	MarketSlug     string    `json:"market_slug"`
	ListingSlug    string    `json:"listing_slug"`
	DisplayName    string    `json:"display_name"`
	Tagline        string    `json:"tagline"`
	ResourceType   string    `json:"resource_type"`
	Outcomes       []string  `json:"outcomes"`
	RuntimeRef     string    `json:"runtime_ref"`
	Status         string    `json:"status"`
	InstalledAt    time.Time `json:"installed_at"`
}

type OrganizationApplicationReader interface {
	ListOrganizationApplications(
		context.Context,
		int64,
	) ([]OrganizationApplication, error)
}

type OrganizationApplicationsAuthorizer interface {
	Authorize(context.Context, int64, int64) error
}

type OrganizationApplicationsService struct {
	reader     OrganizationApplicationReader
	authorizer OrganizationApplicationsAuthorizer
}

func NewOrganizationApplicationsService(
	reader OrganizationApplicationReader,
	authorizer OrganizationApplicationsAuthorizer,
) *OrganizationApplicationsService {
	return &OrganizationApplicationsService{
		reader: reader, authorizer: authorizer,
	}
}

func (s *OrganizationApplicationsService) ListOrganizationApplications(
	ctx context.Context,
	organizationID int64,
	actorUserID int64,
) ([]OrganizationApplication, error) {
	if organizationID <= 0 || actorUserID <= 0 {
		return nil, ErrInvalidInstallationRequest
	}
	if err := s.authorizer.Authorize(ctx, organizationID, actorUserID); err != nil {
		return nil, err
	}
	return s.reader.ListOrganizationApplications(ctx, organizationID)
}
