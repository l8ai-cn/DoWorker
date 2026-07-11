package service

import (
	"context"
	"errors"
	"time"

	marketdomain "github.com/anthropics/agentsmesh/marketplace/internal/domain/market"
)

var (
	ErrMarketNotConfigurable = errors.New("marketplace is not configurable")
	ErrRevisionConflict      = errors.New("revision conflict")
	ErrSpaceNotFound         = errors.New("space not found")
)

type MarketConsoleRepository interface {
	CreateMarketWithDomain(context.Context, *marketdomain.Market, string) error
	GetMarketBySlug(context.Context, string) (*marketdomain.Market, error)
	CreateSpace(context.Context, *marketdomain.Space) error
	GetSpace(context.Context, int64, string) (*marketdomain.Space, error)
	SaveSpace(context.Context, *marketdomain.Space, int64) error
}

type MarketLifecycleService struct {
	repository       MarketConsoleRepository
	publicBaseDomain string
}

type CreateMarketCommand struct {
	Slug        string
	Name        string
	Summary     string
	OwnerOrgID  int64
	ActorUserID int64
}

type CreateSpaceCommand struct {
	MarketSlug  string
	Slug        string
	Name        string
	Summary     string
	ActorUserID int64
}

type PublishSpaceCommand struct {
	MarketSlug       string
	SpaceSlug        string
	ExpectedRevision int64
	PublishedAt      time.Time
}

type MarketResult struct {
	MarketplaceID int64
	Slug          string
	Status        marketdomain.Status
	Revision      int64
}

type SpaceResult struct {
	SpaceID     int64
	Slug        string
	Status      marketdomain.SpaceStatus
	Revision    int64
	PublishedAt *time.Time
}

func NewMarketLifecycleService(
	repository MarketConsoleRepository,
	publicBaseDomain string,
) (*MarketLifecycleService, error) {
	baseDomain, err := normalizePlatformBaseDomain(publicBaseDomain)
	if err != nil {
		return nil, err
	}
	return &MarketLifecycleService{
		repository:       repository,
		publicBaseDomain: baseDomain,
	}, nil
}

func (s *MarketLifecycleService) CreateMarket(
	ctx context.Context,
	command CreateMarketCommand,
) (MarketResult, error) {
	item, err := marketdomain.New(
		command.Slug,
		command.Name,
		command.Summary,
		command.OwnerOrgID,
		command.ActorUserID,
	)
	if err != nil {
		return MarketResult{}, err
	}
	if err := item.Transition(marketdomain.StatusConfiguring); err != nil {
		return MarketResult{}, err
	}
	host := item.Slug().String() + "." + s.publicBaseDomain
	if err := s.repository.CreateMarketWithDomain(ctx, item, host); err != nil {
		return MarketResult{}, err
	}
	return mapMarketResult(item), nil
}

func (s *MarketLifecycleService) CreateSpace(
	ctx context.Context,
	command CreateSpaceCommand,
) (SpaceResult, error) {
	market, err := s.repository.GetMarketBySlug(ctx, command.MarketSlug)
	if err != nil {
		return SpaceResult{}, err
	}
	if !marketConfigurable(market) {
		return SpaceResult{}, ErrMarketNotConfigurable
	}
	space, err := marketdomain.NewSpace(
		market.ID,
		command.Slug,
		command.Name,
		command.Summary,
		command.ActorUserID,
	)
	if err != nil {
		return SpaceResult{}, err
	}
	if err := s.repository.CreateSpace(ctx, space); err != nil {
		return SpaceResult{}, err
	}
	return mapSpaceResult(space, space.Revision()), nil
}

func (s *MarketLifecycleService) PublishSpace(
	ctx context.Context,
	command PublishSpaceCommand,
) (SpaceResult, error) {
	market, err := s.repository.GetMarketBySlug(ctx, command.MarketSlug)
	if err != nil {
		return SpaceResult{}, err
	}
	if !marketConfigurable(market) {
		return SpaceResult{}, ErrMarketNotConfigurable
	}
	space, err := s.repository.GetSpace(ctx, market.ID, command.SpaceSlug)
	if err != nil {
		return SpaceResult{}, err
	}
	if space.Revision() != command.ExpectedRevision {
		return SpaceResult{}, ErrRevisionConflict
	}
	if err := space.Publish(command.PublishedAt); err != nil {
		return SpaceResult{}, err
	}
	if err := s.repository.SaveSpace(ctx, space, command.ExpectedRevision); err != nil {
		return SpaceResult{}, err
	}
	return mapSpaceResult(space, command.ExpectedRevision+1), nil
}

func marketConfigurable(item *marketdomain.Market) bool {
	return item.Status() == marketdomain.StatusConfiguring ||
		item.Status() == marketdomain.StatusReview
}

func mapMarketResult(item *marketdomain.Market) MarketResult {
	return MarketResult{
		MarketplaceID: item.ID,
		Slug:          item.Slug().String(),
		Status:        item.Status(),
		Revision:      item.Revision(),
	}
}

func mapSpaceResult(item *marketdomain.Space, revision int64) SpaceResult {
	return SpaceResult{
		SpaceID:     item.ID,
		Slug:        item.Slug().String(),
		Status:      item.Status(),
		Revision:    revision,
		PublishedAt: item.PublishedAt(),
	}
}
