package preview

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/domain/grant"
	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
	"github.com/anthropics/agentsmesh/backend/pkg/policy"
)

var (
	ErrSessionUnauthorized      = errors.New("preview session unauthorized")
	ErrAuthorizationUnavailable = errors.New("preview authorization unavailable")
)

type SessionIdentity struct {
	ID     string
	PodKey string
	UserID int64
	OrgID  int64
}

type previewUserReader interface {
	GetByID(context.Context, int64) (*user.User, error)
}

type previewOrgRoleReader interface {
	GetUserRole(context.Context, int64, int64) (string, error)
}

type previewPodReader interface {
	GetPod(context.Context, string) (*agentpod.Pod, error)
}

type previewGrantReader interface {
	GetGrantedUserIDs(context.Context, string, string) ([]int64, error)
}

type SessionAuthorizer struct {
	store  *SessionStore
	users  previewUserReader
	orgs   previewOrgRoleReader
	pods   previewPodReader
	grants previewGrantReader
}

func NewSessionAuthorizer(
	store *SessionStore,
	users previewUserReader,
	orgs previewOrgRoleReader,
	pods previewPodReader,
	grants previewGrantReader,
) *SessionAuthorizer {
	return &SessionAuthorizer{store: store, users: users, orgs: orgs, pods: pods, grants: grants}
}

func (a *SessionAuthorizer) Authorize(ctx context.Context, identity SessionIdentity) error {
	if a == nil || a.store == nil || a.users == nil || a.orgs == nil || a.pods == nil {
		return ErrAuthorizationUnavailable
	}
	record, err := a.store.Get(ctx, identity.ID)
	if err != nil {
		if errors.Is(err, ErrSessionInactive) {
			return ErrSessionUnauthorized
		}
		return ErrAuthorizationUnavailable
	}
	if record.ID != identity.ID ||
		record.PodKey != identity.PodKey ||
		record.UserID != identity.UserID ||
		record.OrgID != identity.OrgID {
		return ErrSessionUnauthorized
	}
	currentUser, err := a.users.GetByID(ctx, identity.UserID)
	if err != nil || currentUser == nil || !currentUser.IsActive {
		return ErrSessionUnauthorized
	}
	role, err := a.orgs.GetUserRole(ctx, identity.OrgID, identity.UserID)
	if err != nil {
		return ErrSessionUnauthorized
	}
	pod, err := a.pods.GetPod(ctx, identity.PodKey)
	if err != nil || pod == nil || !pod.IsActive() || pod.OrganizationID != identity.OrgID {
		return ErrSessionUnauthorized
	}
	resource := policy.PodResource(pod.OrganizationID, pod.CreatedByID)
	if a.grants != nil {
		granted, grantErr := a.grants.GetGrantedUserIDs(ctx, grant.TypePod, identity.PodKey)
		if grantErr != nil {
			return ErrAuthorizationUnavailable
		}
		resource = resource.WithGrants(granted)
	}
	if !policy.PodPolicy.AllowRead(policy.NewSubject(identity.OrgID, identity.UserID, role), resource) {
		return ErrSessionUnauthorized
	}
	return nil
}
