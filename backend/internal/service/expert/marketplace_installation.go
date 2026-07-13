package expert

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/google/uuid"
)

var ErrMarketplaceInstallationInvalid = errors.New("invalid marketplace installation")

type MarketplaceInstallationRequest struct {
	InstallationID       string
	TargetOrganizationID int64
	ActorUserID          int64
	RuntimeSnapshot      json.RawMessage
}

type marketplaceExpertSnapshot struct {
	MarketApplicationSlug string `json:"market_application_slug"`
}

func (s *Service) InstallMarketplaceExpert(
	ctx context.Context,
	request MarketplaceInstallationRequest,
) (*expertdom.Expert, bool, error) {
	slug, err := marketplaceInstallationSlug(request)
	if err != nil {
		return nil, false, err
	}
	existing, err := s.store.GetBySlug(ctx, request.TargetOrganizationID, slug)
	if err == nil {
		return existing, true, nil
	}
	if !errors.Is(err, expertdom.ErrNotFound) {
		return nil, false, err
	}
	var snapshot marketplaceExpertSnapshot
	if json.Unmarshal(request.RuntimeSnapshot, &snapshot) != nil {
		return nil, false, ErrMarketplaceInstallationInvalid
	}
	app, ok := findMarketApplication(snapshot.MarketApplicationSlug)
	if !ok {
		return nil, false, ErrMarketApplicationNotFound
	}
	row, _, err := s.installMarketApplication(
		ctx,
		request.TargetOrganizationID,
		request.ActorUserID,
		app,
		slug,
	)
	if err != nil {
		existing, lookupErr := s.store.GetBySlug(
			ctx,
			request.TargetOrganizationID,
			slug,
		)
		if lookupErr == nil {
			return existing, true, nil
		}
	}
	return row, false, err
}

func marketplaceInstallationSlug(
	request MarketplaceInstallationRequest,
) (string, error) {
	if uuid.Validate(request.InstallationID) != nil ||
		request.TargetOrganizationID <= 0 ||
		request.ActorUserID <= 0 ||
		!json.Valid(request.RuntimeSnapshot) {
		return "", ErrMarketplaceInstallationInvalid
	}
	compact := strings.ReplaceAll(request.InstallationID, "-", "")
	return "market-" + compact, nil
}
