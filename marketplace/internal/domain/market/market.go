package market

import (
	"errors"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type Status string

const (
	StatusDraft       Status = "draft"
	StatusConfiguring Status = "configuring"
	StatusReview      Status = "review"
	StatusPublished   Status = "published"
	StatusSuspended   Status = "suspended"
	StatusArchived    Status = "archived"
)

var ErrInvalidTransition = errors.New("invalid marketplace status transition")

type Market struct {
	ID                      int64
	Name                    string
	Summary                 string
	Visibility              string
	OwnerPlatformOrgID      int64
	CreatedByPlatformUserID int64
	revision                int64
	slug                    slugkit.Slug
	status                  Status
}

type State struct {
	ID                      int64
	Slug                    string
	Name                    string
	Summary                 string
	Status                  Status
	Visibility              string
	OwnerPlatformOrgID      int64
	CreatedByPlatformUserID int64
	Revision                int64
}

func New(rawSlug, name, summary string, ownerOrgID, actorUserID int64) (*Market, error) {
	slug, err := slugkit.NewFromTrusted(rawSlug)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(name) == "" || strings.TrimSpace(summary) == "" {
		return nil, errors.New("marketplace name and summary are required")
	}
	if ownerOrgID <= 0 || actorUserID <= 0 {
		return nil, errors.New("marketplace owner and creator are required")
	}
	return &Market{
		Name:                    strings.TrimSpace(name),
		Summary:                 strings.TrimSpace(summary),
		Visibility:              "private",
		OwnerPlatformOrgID:      ownerOrgID,
		CreatedByPlatformUserID: actorUserID,
		revision:                1,
		slug:                    slug,
		status:                  StatusDraft,
	}, nil
}

func Restore(state State) (*Market, error) {
	item, err := New(
		state.Slug,
		state.Name,
		state.Summary,
		state.OwnerPlatformOrgID,
		state.CreatedByPlatformUserID,
	)
	if err != nil {
		return nil, err
	}
	if !validStatuses[state.Status] {
		return nil, ErrInvalidTransition
	}
	if state.Visibility != "public" && state.Visibility != "private" {
		return nil, errors.New("invalid marketplace visibility")
	}
	if state.Revision <= 0 {
		return nil, errors.New("invalid marketplace revision")
	}
	item.ID = state.ID
	item.status = state.Status
	item.Visibility = state.Visibility
	item.revision = state.Revision
	return item, nil
}

func (m *Market) Transition(next Status) error {
	if !allowedTransitions[m.status][next] {
		return ErrInvalidTransition
	}
	m.status = next
	return nil
}

func (m Market) Status() Status     { return m.status }
func (m Market) Slug() slugkit.Slug { return m.slug }
func (m Market) Revision() int64    { return m.revision }

var allowedTransitions = map[Status]map[Status]bool{
	StatusDraft:       {StatusConfiguring: true},
	StatusConfiguring: {StatusReview: true},
	StatusReview:      {StatusPublished: true},
	StatusPublished:   {StatusSuspended: true},
	StatusSuspended:   {StatusArchived: true},
}

var validStatuses = map[Status]bool{
	StatusDraft:       true,
	StatusConfiguring: true,
	StatusReview:      true,
	StatusPublished:   true,
	StatusSuspended:   true,
	StatusArchived:    true,
}
