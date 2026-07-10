package workerspec

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidOrganizationID = errors.New("workerspec organization id must be positive")
	ErrNotFound              = errors.New("workerspec snapshot not found")
	ErrSummaryMismatch       = errors.New("workerspec snapshot summary mismatch")
)

type Snapshot struct {
	ID             int64
	OrganizationID int64
	Spec           Spec
	Summary        Summary
	CreatedAt      time.Time
}

func NewSnapshot(organizationID int64, spec Spec) (*Snapshot, error) {
	if organizationID <= 0 {
		return nil, ErrInvalidOrganizationID
	}
	normalized, err := NormalizeAndValidate(spec)
	if err != nil {
		return nil, err
	}
	summary, err := Summarize(normalized)
	if err != nil {
		return nil, fmt.Errorf("summarize workerspec snapshot: %w", err)
	}
	return &Snapshot{
		OrganizationID: organizationID,
		Spec:           normalized,
		Summary:        summary,
	}, nil
}
