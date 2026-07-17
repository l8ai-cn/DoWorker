package preview

import (
	"context"
	"time"
)

type Service struct {
	store      *SessionStore
	authorizer *SessionAuthorizer
}

func NewService(
	store *SessionStore,
	users previewUserReader,
	orgs previewOrgRoleReader,
	pods previewPodReader,
	grants previewGrantReader,
) *Service {
	return &Service{
		store:      store,
		authorizer: NewSessionAuthorizer(store, users, orgs, pods, grants),
	}
}

func (s *Service) Redeem(ctx context.Context, bootstrapID string, record SessionRecord, bootstrapTTL time.Duration) error {
	if s == nil || s.store == nil {
		return ErrStoreUnavailable
	}
	return s.store.Redeem(ctx, bootstrapID, record, bootstrapTTL)
}

func (s *Service) Authorize(ctx context.Context, identity SessionIdentity) error {
	if s == nil || s.authorizer == nil {
		return ErrAuthorizationUnavailable
	}
	return s.authorizer.Authorize(ctx, identity)
}

func (s *Service) RevokeUser(ctx context.Context, userID int64) error {
	if s == nil || s.store == nil {
		return ErrStoreUnavailable
	}
	return s.store.RevokeUser(ctx, userID)
}
