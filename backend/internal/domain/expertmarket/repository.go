package expertmarket

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound                        = errors.New("expert market resource not found")
	ErrConflict                        = errors.New("expert market resource conflicts with persisted state")
	ErrInvalidLatestReleaseStatus      = errors.New("latest expert market release must be published")
	ErrPublicationRequiresLatestUpdate = errors.New(
		"publishing expert market release requires latest update",
	)
	ErrLatestReleaseStatusConflict = errors.New(
		"latest expert market release cannot be demoted",
	)
	ErrInvalidWithdrawalStatus = errors.New(
		"expert market release withdrawal must use withdrawn status",
	)
	ErrPendingReleaseExists = errors.New(
		"expert market application already has a pending release",
	)
	ErrLifecycleStatusConflict = errors.New(
		"expert market release status changed before lifecycle update",
	)
)

type ApplicationListFilter struct {
	PublisherOrganizationID *int64
	Limit                   int
	Offset                  int
}

type ReleaseListFilter struct {
	ApplicationID           *int64
	PublisherOrganizationID *int64
	Status                  *ReleaseStatus
	Limit                   int
	Offset                  int
}

type LifecycleUpdate struct {
	Status          ReleaseStatus
	ExpectedStatus  *ReleaseStatus
	ReviewerUserID  *int64
	RejectionReason *string
	SubmittedAt     *time.Time
	ReviewedAt      *time.Time
	PublishedAt     *time.Time
	RejectedAt      *time.Time
	WithdrawnAt     *time.Time
}

func (update LifecycleUpdate) Validate() error {
	if !update.Status.Valid() {
		return ErrInvalidStatus
	}
	if update.ExpectedStatus != nil && !update.ExpectedStatus.Valid() {
		return ErrInvalidStatus
	}
	return nil
}

type Repository interface {
	CreateApplication(ctx context.Context, application *Application) error
	GetApplicationByID(ctx context.Context, id int64) (*Application, error)
	GetApplicationBySlug(ctx context.Context, slug string) (*Application, error)
	ListApplications(ctx context.Context, filter ApplicationListFilter) ([]Application, int64, error)

	CreateRelease(ctx context.Context, release *Release) error
	CreateSubmission(
		ctx context.Context,
		application *Application,
		release *Release,
	) error
	GetReleaseByID(ctx context.Context, id int64) (*Release, error)
	ListReleases(ctx context.Context, filter ReleaseListFilter) ([]Release, int64, error)
	UpdateReleaseLifecycle(ctx context.Context, releaseID int64, update LifecycleUpdate) error
	UpdateReleaseLifecycleAndLatest(
		ctx context.Context,
		applicationID, releaseID int64,
		update LifecycleUpdate,
	) error
	WithdrawReleaseAndRefreshLatest(
		ctx context.Context,
		applicationID, releaseID int64,
		update LifecycleUpdate,
	) error
}
