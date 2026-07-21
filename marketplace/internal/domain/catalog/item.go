package catalog

import (
	"errors"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type ItemStatus string

const (
	ItemStatusDraft  ItemStatus = "draft"
	ItemStatusActive ItemStatus = "active"
)

var (
	ErrVersionNotPassed    = errors.New("catalog version has not passed validation")
	ErrVersionItemMismatch = errors.New("catalog version belongs to another item")
)

type Item struct {
	id                      int64
	publisherID             int64
	slug                    slugkit.Slug
	resourceType            string
	name                    string
	summary                 string
	platformResourceType    string
	platformResourceID      int64
	createdByPlatformUserID int64
	status                  ItemStatus
	latestVersionID         int64
}

type ItemState struct {
	ID                      int64
	PublisherID             int64
	Slug                    string
	ResourceType            string
	Name                    string
	Summary                 string
	PlatformResourceType    string
	PlatformResourceID      int64
	CreatedByPlatformUserID int64
	Status                  ItemStatus
	LatestVersionID         int64
}

func NewItem(
	publisherID int64,
	rawSlug string,
	resourceType string,
	name string,
	summary string,
	platformResourceType string,
	platformResourceID int64,
	actorUserID int64,
) (*Item, error) {
	slug, err := slugkit.NewFromTrusted(rawSlug)
	if err != nil {
		return nil, err
	}
	if publisherID <= 0 || platformResourceID <= 0 || actorUserID <= 0 {
		return nil, errors.New("catalog item references are required")
	}
	if !validResourceTypes[resourceType] {
		return nil, errors.New("invalid catalog resource type")
	}
	if strings.TrimSpace(name) == "" || strings.TrimSpace(summary) == "" ||
		strings.TrimSpace(platformResourceType) == "" {
		return nil, errors.New("catalog item metadata is required")
	}
	return &Item{
		publisherID:             publisherID,
		slug:                    slug,
		resourceType:            resourceType,
		name:                    strings.TrimSpace(name),
		summary:                 strings.TrimSpace(summary),
		platformResourceType:    strings.TrimSpace(platformResourceType),
		platformResourceID:      platformResourceID,
		createdByPlatformUserID: actorUserID,
		status:                  ItemStatusDraft,
	}, nil
}

func RestoreItem(state ItemState) (*Item, error) {
	item, err := NewItem(
		state.PublisherID,
		state.Slug,
		state.ResourceType,
		state.Name,
		state.Summary,
		state.PlatformResourceType,
		state.PlatformResourceID,
		state.CreatedByPlatformUserID,
	)
	if err != nil {
		return nil, err
	}
	if state.Status != ItemStatusDraft && state.Status != ItemStatusActive {
		return nil, errors.New("invalid catalog item status")
	}
	if state.Status == ItemStatusActive && state.LatestVersionID <= 0 {
		return nil, errors.New("active catalog item requires latest version")
	}
	item.id = state.ID
	item.status = state.Status
	item.latestVersionID = state.LatestVersionID
	return item, nil
}

func (i *Item) ActivateVersion(version *Version) error {
	if version.ValidationStatus() != ValidationPassed {
		return ErrVersionNotPassed
	}
	if version.CatalogItemID() != i.id {
		return ErrVersionItemMismatch
	}
	i.status = ItemStatusActive
	i.latestVersionID = version.ID()
	return nil
}

func (i Item) ID() int64                    { return i.id }
func (i Item) PublisherID() int64           { return i.publisherID }
func (i Item) Slug() slugkit.Slug           { return i.slug }
func (i Item) ResourceType() string         { return i.resourceType }
func (i Item) Name() string                 { return i.name }
func (i Item) Summary() string              { return i.summary }
func (i Item) PlatformResourceType() string { return i.platformResourceType }
func (i Item) PlatformResourceID() int64    { return i.platformResourceID }
func (i Item) Status() ItemStatus           { return i.status }
func (i Item) LatestVersionID() int64       { return i.latestVersionID }
func (i Item) CreatedByPlatformUserID() int64 {
	return i.createdByPlatformUserID
}

var validResourceTypes = map[string]bool{
	"application":   true,
	"skill":         true,
	"mcp_connector": true,
	"resource":      true,
}
