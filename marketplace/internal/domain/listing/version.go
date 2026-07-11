package listing

import (
	"encoding/json"
	"errors"
	"strings"
)

type ReviewStatus string

const (
	ReviewDraft     ReviewStatus = "draft"
	ReviewSubmitted ReviewStatus = "submitted"
	ReviewApproved  ReviewStatus = "approved"
	ReviewRejected  ReviewStatus = "rejected"
)

var (
	ErrInvalidPresentation     = errors.New("invalid listing presentation")
	ErrInvalidReviewTransition = errors.New("invalid listing review transition")
)

type Version struct {
	id                   int64
	listingID            int64
	catalogItemVersionID int64
	revision             int
	displayName          string
	tagline              string
	description          string
	outcomes             json.RawMessage
	useCases             json.RawMessage
	targetAudience       json.RawMessage
	requirements         json.RawMessage
	tags                 []string
	releaseNotes         string
	reviewStatus         ReviewStatus
}

func NewVersion(
	listingID, catalogItemVersionID int64,
	revision int,
	displayName, tagline, description string,
	outcomes, useCases, targetAudience, requirements json.RawMessage,
	tags []string,
	releaseNotes string,
) (*Version, error) {
	if listingID < 0 || catalogItemVersionID <= 0 || revision <= 0 ||
		strings.TrimSpace(displayName) == "" ||
		strings.TrimSpace(tagline) == "" ||
		strings.TrimSpace(description) == "" ||
		len(tags) > 12 ||
		!jsonArray(outcomes) || !jsonArray(useCases) ||
		!jsonArray(targetAudience) || !jsonArray(requirements) {
		return nil, ErrInvalidPresentation
	}
	return &Version{
		listingID:            listingID,
		catalogItemVersionID: catalogItemVersionID,
		revision:             revision,
		displayName:          strings.TrimSpace(displayName),
		tagline:              strings.TrimSpace(tagline),
		description:          strings.TrimSpace(description),
		outcomes:             cloneJSON(outcomes),
		useCases:             cloneJSON(useCases),
		targetAudience:       cloneJSON(targetAudience),
		requirements:         cloneJSON(requirements),
		tags:                 append([]string(nil), tags...),
		releaseNotes:         strings.TrimSpace(releaseNotes),
		reviewStatus:         ReviewDraft,
	}, nil
}

func (v *Version) Submit() error {
	if v.reviewStatus != ReviewDraft {
		return ErrInvalidReviewTransition
	}
	v.reviewStatus = ReviewSubmitted
	return nil
}

func (v *Version) Approve() error {
	if v.reviewStatus != ReviewSubmitted {
		return ErrInvalidReviewTransition
	}
	v.reviewStatus = ReviewApproved
	return nil
}

func (v Version) ID() int64                       { return v.id }
func (v Version) ListingID() int64                { return v.listingID }
func (v Version) CatalogItemVersionID() int64     { return v.catalogItemVersionID }
func (v Version) Revision() int                   { return v.revision }
func (v Version) DisplayName() string             { return v.displayName }
func (v Version) Tagline() string                 { return v.tagline }
func (v Version) Description() string             { return v.description }
func (v Version) Tags() []string                  { return append([]string(nil), v.tags...) }
func (v Version) ReleaseNotes() string            { return v.releaseNotes }
func (v Version) ReviewStatus() ReviewStatus      { return v.reviewStatus }
func (v Version) Outcomes() json.RawMessage       { return cloneJSON(v.outcomes) }
func (v Version) UseCases() json.RawMessage       { return cloneJSON(v.useCases) }
func (v Version) TargetAudience() json.RawMessage { return cloneJSON(v.targetAudience) }
func (v Version) Requirements() json.RawMessage   { return cloneJSON(v.requirements) }
func (v *Version) AssignID(id int64)              { v.id = id }
func (v *Version) BindListingID(id int64) error {
	if v.listingID != 0 || id <= 0 {
		return ErrInvalidPresentation
	}
	v.listingID = id
	return nil
}

func jsonArray(value json.RawMessage) bool {
	var items []any
	return json.Unmarshal(value, &items) == nil
}

func cloneJSON(value json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), value...)
}
