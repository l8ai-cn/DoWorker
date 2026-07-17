package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/anthropics/agentsmesh/marketplace/internal/domain/catalog"
)

var (
	ErrCatalogItemNotFound     = errors.New("catalog item not found")
	ErrCatalogVersionNotFound  = errors.New("catalog version not found")
	ErrCatalogVersionNotPassed = errors.New("catalog version has not passed validation")
)

type CatalogRegistrationRepository interface {
	CreateCatalogItem(context.Context, *catalog.Item) (int64, error)
	CreateCatalogVersion(context.Context, *catalog.Version) (int64, error)
	GetCatalogItem(context.Context, int64) (*catalog.Item, error)
	GetCatalogVersion(context.Context, int64) (*catalog.Version, error)
	ActivateCatalogVersion(context.Context, *catalog.Item, *catalog.Version) error
}

type CatalogRegistrationService struct {
	repository CatalogRegistrationRepository
}

type RegisterCatalogItemCommand struct {
	PublisherID          int64
	Slug                 string
	ResourceType         string
	Name                 string
	Summary              string
	PlatformResourceType string
	PlatformResourceID   int64
	ActorUserID          int64
}

type RegisterCatalogVersionCommand struct {
	CatalogItemID  int64
	Version        string
	SourceRevision string
	ContentDigest  string
	Manifest       json.RawMessage
	Compatibility  json.RawMessage
	ActorUserID    int64
}

type CatalogItemResult struct {
	CatalogItemID   int64
	Slug            string
	ItemStatus      catalog.ItemStatus
	LatestVersionID int64
}

type CatalogVersionResult struct {
	CatalogItemVersionID int64
	CatalogItemID        int64
	Version              string
	ValidationStatus     catalog.ValidationStatus
}

func NewCatalogRegistrationService(
	repository CatalogRegistrationRepository,
) *CatalogRegistrationService {
	return &CatalogRegistrationService{repository: repository}
}

func (s *CatalogRegistrationService) RegisterItem(
	ctx context.Context,
	command RegisterCatalogItemCommand,
) (CatalogItemResult, error) {
	item, err := catalog.NewItem(
		command.PublisherID,
		command.Slug,
		command.ResourceType,
		command.Name,
		command.Summary,
		command.PlatformResourceType,
		command.PlatformResourceID,
		command.ActorUserID,
	)
	if err != nil {
		return CatalogItemResult{}, err
	}
	id, err := s.repository.CreateCatalogItem(ctx, item)
	if err != nil {
		return CatalogItemResult{}, err
	}
	return CatalogItemResult{
		CatalogItemID: id,
		Slug:          item.Slug().String(),
		ItemStatus:    item.Status(),
	}, nil
}

func (s *CatalogRegistrationService) RegisterVersion(
	ctx context.Context,
	command RegisterCatalogVersionCommand,
) (CatalogVersionResult, error) {
	version, err := catalog.NewVersion(
		command.CatalogItemID,
		command.Version,
		command.SourceRevision,
		command.ContentDigest,
		command.Manifest,
		command.Compatibility,
		command.ActorUserID,
	)
	if err != nil {
		return CatalogVersionResult{}, err
	}
	id, err := s.repository.CreateCatalogVersion(ctx, version)
	if err != nil {
		return CatalogVersionResult{}, err
	}
	return CatalogVersionResult{
		CatalogItemVersionID: id,
		CatalogItemID:        version.CatalogItemID(),
		Version:              version.Version(),
		ValidationStatus:     version.ValidationStatus(),
	}, nil
}

func (s *CatalogRegistrationService) MarkVersionPassed(
	ctx context.Context,
	versionID int64,
) (CatalogItemResult, error) {
	version, err := s.repository.GetCatalogVersion(ctx, versionID)
	if err != nil {
		return CatalogItemResult{}, err
	}
	item, err := s.repository.GetCatalogItem(ctx, version.CatalogItemID())
	if err != nil {
		return CatalogItemResult{}, err
	}
	version.MarkValidationPassed()
	if err := item.ActivateVersion(version); err != nil {
		return CatalogItemResult{}, err
	}
	if err := s.repository.ActivateCatalogVersion(ctx, item, version); err != nil {
		return CatalogItemResult{}, err
	}
	return CatalogItemResult{
		CatalogItemID:   item.ID(),
		Slug:            item.Slug().String(),
		ItemStatus:      item.Status(),
		LatestVersionID: item.LatestVersionID(),
	}, nil
}
