package listing

import (
	"errors"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type Status string
type Visibility string
type AccessMode string

const (
	StatusDraft        Status = "draft"
	StatusSubmitted    Status = "submitted"
	StatusValidating   Status = "validating"
	StatusNeedsChanges Status = "needs_changes"
	StatusApproved     Status = "approved"
	StatusPublished    Status = "published"
	StatusSuspended    Status = "suspended"
	StatusDeprecated   Status = "deprecated"
	StatusRemoved      Status = "removed"
)

const (
	VisibilityPublic  Visibility = "public"
	VisibilityMembers Visibility = "members"
	VisibilityHidden  Visibility = "hidden"
)

const (
	AccessModeDirect    AccessMode = "direct"
	AccessModeApproval  AccessMode = "approval"
	AccessModeGrantOnly AccessMode = "grant_only"
)

var (
	ErrNotApproved           = errors.New("listing is not approved")
	ErrNotSubmitted          = errors.New("listing is not submitted")
	ErrSubmitterCannotReview = errors.New("listing submitter cannot review")
	ErrVersionRequired       = errors.New("listing version is required")
	ErrPrimarySpaceRequired  = errors.New("primary space is required")
	ErrInvalidVisibility     = errors.New("invalid listing visibility")
	ErrInvalidAccessMode     = errors.New("invalid listing access mode")
)

type Listing struct {
	ID               int64
	MarketplaceID    int64
	CatalogItemID    int64
	slug             slugkit.Slug
	status           Status
	visibility       Visibility
	accessMode       AccessMode
	currentVersionID int64
	submittedBy      int64
	publishedAt      *time.Time
	hasPrimarySpace  bool
	revision         int64
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
		accessMode:    AccessModeDirect,
		revision:      1,
	}, nil
}

func (l *Listing) Submit(actorUserID int64) error {
	if l.status != StatusDraft || actorUserID <= 0 {
		return ErrNotSubmitted
	}
	l.status = StatusSubmitted
	l.submittedBy = actorUserID
	return nil
}

func (l *Listing) Approve(actorUserID int64) error {
	if l.status != StatusSubmitted {
		return ErrNotSubmitted
	}
	if actorUserID == l.submittedBy {
		return ErrSubmitterCannotReview
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

func (l *Listing) SetAccessMode(accessMode AccessMode) error {
	if !validAccessModes[accessMode] {
		return ErrInvalidAccessMode
	}
	l.accessMode = accessMode
	return nil
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
func (l Listing) AccessMode() AccessMode  { return l.accessMode }
func (l Listing) Slug() slugkit.Slug      { return l.slug }
func (l Listing) CurrentVersionID() int64 { return l.currentVersionID }
func (l Listing) SubmittedBy() int64      { return l.submittedBy }
func (l Listing) Revision() int64         { return l.revision }
func (l *Listing) AdvanceRevision(expected int64) error {
	if l.revision != expected {
		return errors.New("listing revision conflict")
	}
	l.revision++
	return nil
}
func (l Listing) PublishedAt() *time.Time {
	if l.publishedAt == nil {
		return nil
	}
	at := *l.publishedAt
	return &at
}

var validStatuses = map[Status]bool{
	StatusDraft:        true,
	StatusSubmitted:    true,
	StatusValidating:   true,
	StatusNeedsChanges: true,
	StatusApproved:     true,
	StatusPublished:    true,
	StatusSuspended:    true,
	StatusDeprecated:   true,
	StatusRemoved:      true,
}

var validVisibilities = map[Visibility]bool{
	VisibilityPublic:  true,
	VisibilityMembers: true,
	VisibilityHidden:  true,
}

var validAccessModes = map[AccessMode]bool{
	AccessModeDirect:    true,
	AccessModeApproval:  true,
	AccessModeGrantOnly: true,
}
