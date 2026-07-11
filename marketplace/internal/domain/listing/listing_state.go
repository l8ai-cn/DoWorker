package listing

import (
	"errors"
	"time"
)

type State struct {
	ID               int64
	MarketplaceID    int64
	CatalogItemID    int64
	Slug             string
	Status           Status
	Visibility       Visibility
	AccessMode       AccessMode
	CurrentVersionID int64
	SubmittedBy      int64
	PublishedAt      *time.Time
	HasPrimarySpace  bool
	Revision         int64
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
	if !validAccessModes[state.AccessMode] {
		return nil, ErrInvalidAccessMode
	}
	if state.Revision <= 0 {
		return nil, errors.New("invalid listing revision")
	}
	if state.Status != StatusDraft && state.SubmittedBy <= 0 {
		return nil, ErrNotSubmitted
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
	item.accessMode = state.AccessMode
	item.currentVersionID = state.CurrentVersionID
	item.submittedBy = state.SubmittedBy
	item.hasPrimarySpace = state.HasPrimarySpace
	item.revision = state.Revision
	if state.PublishedAt != nil {
		at := *state.PublishedAt
		item.publishedAt = &at
	}
	return item, nil
}
