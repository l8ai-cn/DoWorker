package preview

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/organization"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/user"
)

type userReaderStub struct {
	value *user.User
	err   error
}

func (s userReaderStub) GetByID(context.Context, int64) (*user.User, error) {
	return s.value, s.err
}

type orgRoleReaderStub struct {
	role string
	err  error
}

func (s orgRoleReaderStub) GetUserRole(context.Context, int64, int64) (string, error) {
	return s.role, s.err
}

type podReaderStub struct {
	value *agentpod.Pod
	err   error
}

func (s podReaderStub) GetPod(context.Context, string) (*agentpod.Pod, error) {
	return s.value, s.err
}

type grantReaderStub struct {
	userIDs []int64
	err     error
}

func (s grantReaderStub) GetGrantedUserIDs(context.Context, string, string) ([]int64, error) {
	return s.userIDs, s.err
}

func TestSessionAuthorizerChecksCurrentPodAccess(t *testing.T) {
	store := newSessionStore(t)
	record := SessionRecord{
		ID:        "session-1",
		PodKey:    "pod-1",
		UserID:    42,
		OrgID:     3,
		ExpiresAt: time.Now().Add(time.Minute),
	}
	if err := store.Redeem(context.Background(), "bootstrap-1", record, time.Minute); err != nil {
		t.Fatal(err)
	}
	authorizer := NewSessionAuthorizer(
		store,
		userReaderStub{value: &user.User{ID: 42, IsActive: true}},
		orgRoleReaderStub{role: organization.RoleMember},
		podReaderStub{value: &agentpod.Pod{
			PodKey:         "pod-1",
			OrganizationID: 3,
			CreatedByID:    42,
			Status:         agentpod.StatusRunning,
		}},
		grantReaderStub{},
	)

	if err := authorizer.Authorize(context.Background(), SessionIdentity{
		ID: "session-1", PodKey: "pod-1", UserID: 42, OrgID: 3,
	}); err != nil {
		t.Fatalf("active owner denied: %v", err)
	}
}

func TestSessionAuthorizerRejectsCurrentAccessChanges(t *testing.T) {
	for _, test := range []struct {
		name string
		user userReaderStub
		org  orgRoleReaderStub
		pod  podReaderStub
	}{
		{
			name: "disabled user",
			user: userReaderStub{value: &user.User{ID: 42, IsActive: false}},
			org:  orgRoleReaderStub{role: organization.RoleMember},
			pod:  activePodStub(),
		},
		{
			name: "removed organization member",
			user: userReaderStub{value: &user.User{ID: 42, IsActive: true}},
			org:  orgRoleReaderStub{err: errors.New("member missing")},
			pod:  activePodStub(),
		},
		{
			name: "stopped pod",
			user: userReaderStub{value: &user.User{ID: 42, IsActive: true}},
			org:  orgRoleReaderStub{role: organization.RoleMember},
			pod: podReaderStub{value: &agentpod.Pod{
				PodKey:         "pod-1",
				OrganizationID: 3,
				CreatedByID:    42,
				Status:         agentpod.StatusTerminated,
			}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newSessionStore(t)
			record := SessionRecord{
				ID: "session-1", PodKey: "pod-1", UserID: 42, OrgID: 3,
				ExpiresAt: time.Now().Add(time.Minute),
			}
			if err := store.Redeem(context.Background(), "bootstrap-1", record, time.Minute); err != nil {
				t.Fatal(err)
			}
			authorizer := NewSessionAuthorizer(store, test.user, test.org, test.pod, grantReaderStub{})
			err := authorizer.Authorize(context.Background(), SessionIdentity{
				ID: "session-1", PodKey: "pod-1", UserID: 42, OrgID: 3,
			})
			if !errors.Is(err, ErrSessionUnauthorized) {
				t.Fatalf("error = %v, want unauthorized", err)
			}
		})
	}
}

func activePodStub() podReaderStub {
	return podReaderStub{value: &agentpod.Pod{
		PodKey:         "pod-1",
		OrganizationID: 3,
		CreatedByID:    42,
		Status:         agentpod.StatusRunning,
	}}
}
