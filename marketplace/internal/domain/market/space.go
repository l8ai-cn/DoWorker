package market

import (
	"errors"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type SpaceStatus string

const (
	SpaceStatusDraft     SpaceStatus = "draft"
	SpaceStatusPublished SpaceStatus = "published"
	SpaceStatusHidden    SpaceStatus = "hidden"
	SpaceStatusArchived  SpaceStatus = "archived"
)

var ErrInvalidSpaceTransition = errors.New("invalid space status transition")

type Space struct {
	ID                      int64
	MarketplaceID           int64
	Name                    string
	Summary                 string
	CreatedByPlatformUserID int64
	revision                int64
	slug                    slugkit.Slug
	status                  SpaceStatus
	publishedAt             *time.Time
}

type SpaceState struct {
	ID                      int64
	MarketplaceID           int64
	Slug                    string
	Name                    string
	Summary                 string
	Status                  SpaceStatus
	Revision                int64
	CreatedByPlatformUserID int64
	PublishedAt             *time.Time
}

func NewSpace(
	marketplaceID int64,
	rawSlug string,
	name string,
	summary string,
	actorUserID int64,
) (*Space, error) {
	slug, err := slugkit.NewFromTrusted(rawSlug)
	if err != nil {
		return nil, err
	}
	if marketplaceID <= 0 || actorUserID <= 0 {
		return nil, errors.New("space marketplace and creator are required")
	}
	if strings.TrimSpace(name) == "" || strings.TrimSpace(summary) == "" {
		return nil, errors.New("space name and summary are required")
	}
	return &Space{
		MarketplaceID:           marketplaceID,
		Name:                    strings.TrimSpace(name),
		Summary:                 strings.TrimSpace(summary),
		CreatedByPlatformUserID: actorUserID,
		revision:                1,
		slug:                    slug,
		status:                  SpaceStatusDraft,
	}, nil
}

func RestoreSpace(state SpaceState) (*Space, error) {
	space, err := NewSpace(
		state.MarketplaceID,
		state.Slug,
		state.Name,
		state.Summary,
		state.CreatedByPlatformUserID,
	)
	if err != nil {
		return nil, err
	}
	if !validSpaceStatuses[state.Status] || state.Revision <= 0 {
		return nil, ErrInvalidSpaceTransition
	}
	space.ID = state.ID
	space.status = state.Status
	space.revision = state.Revision
	if state.PublishedAt != nil {
		at := *state.PublishedAt
		space.publishedAt = &at
	}
	return space, nil
}

func (s *Space) Publish(at time.Time) error {
	if s.status != SpaceStatusDraft {
		return ErrInvalidSpaceTransition
	}
	s.status = SpaceStatusPublished
	s.publishedAt = &at
	return nil
}

func (s Space) Revision() int64     { return s.revision }
func (s Space) Slug() slugkit.Slug  { return s.slug }
func (s Space) Status() SpaceStatus { return s.status }
func (s Space) PublishedAt() *time.Time {
	if s.publishedAt == nil {
		return nil
	}
	at := *s.publishedAt
	return &at
}

var validSpaceStatuses = map[SpaceStatus]bool{
	SpaceStatusDraft:     true,
	SpaceStatusPublished: true,
	SpaceStatusHidden:    true,
	SpaceStatusArchived:  true,
}
