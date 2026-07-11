package listing

import (
	"errors"
	"time"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type Status string
type Visibility string

const (
	StatusDraft     Status = "draft"
	StatusApproved  Status = "approved"
	StatusPublished Status = "published"
)

const (
	VisibilityPublic  Visibility = "public"
	VisibilityMembers Visibility = "members"
	VisibilityHidden  Visibility = "hidden"
)

var (
	ErrNotApproved          = errors.New("listing is not approved")
	ErrVersionRequired      = errors.New("listing version is required")
	ErrPrimarySpaceRequired = errors.New("primary space is required")
	ErrInvalidVisibility    = errors.New("invalid listing visibility")
)

type Listing struct {
	ID               int64
	MarketplaceID    int64
	CatalogItemID    int64
	slug             slugkit.Slug
	status           Status
	visibility       Visibility
	currentVersionID int64
	publishedAt      *time.Time
	hasPrimarySpace  bool
}

type State struct {
	ID               int64
	MarketplaceID    int64
	CatalogItemID    int64
	Slug             string
	Status           Status
	Visibility       Visibility
	CurrentVersionID int64
	PublishedAt      *time.Time
	HasPrimarySpace  bool
}

func New(marketplaceID, catalogItemID int64, rawSlug string) (*Listing, error) {
	slug, err := slugkit.NewFromTrusted(rawSlug)
	if err != nil {
		return nil, err
	}
	if marketplaceID <= 0 || catalogItemID <= 0 {
		return nil, errors.New("marketplace and catalog item are required")
	}
	return &Listing{
		MarketplaceID: marketplaceID,
		CatalogItemID: catalogItemID,
		slug:          slug,
		status:        StatusDraft,
		visibility:    VisibilityHidden,
	}, nil
}

func Restore(state State) (*Listing, error) {
	item, err := New(state.MarketplaceID, state.CatalogItemID, state.Slug)
	if err != nil {
		return nil, err
	}
	if !validStatuses[state.Status] {
		return nil, errors.New("invalid listing status")
	}
	if !validVisibilities[state.Visibility] {
		return nil, ErrInvalidVisibility
	}
	if state.Status == StatusPublished &&
		(state.CurrentVersionID <= 0 || state.PublishedAt == nil) {
		return nil, ErrVersionRequired
	}
	if state.Status == StatusPublished && !state.HasPrimarySpace {
		return nil, ErrPrimarySpaceRequired
	}
	item.ID = state.ID
	item.status = state.Status
	item.visibility = state.Visibility
	item.currentVersionID = state.CurrentVersionID
	item.hasPrimarySpace = state.HasPrimarySpace
	if state.PublishedAt != nil {
		at := *state.PublishedAt
		item.publishedAt = &at
	}
	return item, nil
}

func (l *Listing) Approve() error {
	if l.status != StatusDraft {
		return ErrNotApproved
	}
	l.status = StatusApproved
	return nil
}

func (l *Listing) SetVisibility(visibility Visibility) error {
	switch visibility {
	case VisibilityPublic, VisibilityMembers, VisibilityHidden:
		l.visibility = visibility
		return nil
	default:
		return ErrInvalidVisibility
	}
}

func (l *Listing) Publish(versionID int64, hasPrimarySpace bool, at time.Time) error {
	if l.status != StatusApproved {
		return ErrNotApproved
	}
	if versionID <= 0 {
		return ErrVersionRequired
	}
	if !hasPrimarySpace {
		return ErrPrimarySpaceRequired
	}
	l.currentVersionID = versionID
	l.status = StatusPublished
	l.publishedAt = &at
	l.hasPrimarySpace = true
	return nil
}

func (l Listing) IsPublic() bool {
	return l.status == StatusPublished &&
		l.visibility == VisibilityPublic &&
		l.currentVersionID > 0 &&
		l.publishedAt != nil &&
		l.hasPrimarySpace
}

func (l Listing) Status() Status          { return l.status }
func (l Listing) Visibility() Visibility  { return l.visibility }
func (l Listing) Slug() slugkit.Slug      { return l.slug }
func (l Listing) CurrentVersionID() int64 { return l.currentVersionID }
func (l Listing) PublishedAt() *time.Time {
	if l.publishedAt == nil {
		return nil
	}
	at := *l.publishedAt
	return &at
}

var validStatuses = map[Status]bool{
	StatusDraft:     true,
	StatusApproved:  true,
	StatusPublished: true,
}

var validVisibilities = map[Visibility]bool{
	VisibilityPublic:  true,
	VisibilityMembers: true,
	VisibilityHidden:  true,
}
