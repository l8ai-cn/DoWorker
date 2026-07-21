package virtualkey

import (
	"context"

	virtualkeydomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/virtualkey"
	airesourcesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
)

type callTrace struct{ calls []string }

func (t *callTrace) add(call string) { t.calls = append(t.calls, call) }

type scopedKeyCall struct {
	id     int64
	orgID  int64
	userID int64
}

type scopedStatusCall struct {
	id     int64
	orgID  int64
	userID int64
	status string
}

type fakeVirtualKeyRepository struct {
	keys        map[int64]*virtualkeydomain.VirtualAPIKey
	createErr   error
	scopedErr   error
	touchErr    error
	created     []*virtualkeydomain.VirtualAPIKey
	scopedCalls []scopedKeyCall
	statusCalls []scopedStatusCall
	touchCalls  []int64
	trace       *callTrace
}

func (r *fakeVirtualKeyRepository) Create(_ context.Context, key *virtualkeydomain.VirtualAPIKey) error {
	r.trace.add("key.create")
	r.created = append(r.created, key)
	return r.createErr
}

func (r *fakeVirtualKeyRepository) GetByID(_ context.Context, id int64) (*virtualkeydomain.VirtualAPIKey, error) {
	return r.keys[id], nil
}

func (r *fakeVirtualKeyRepository) GetByIDForScope(
	_ context.Context, id, orgID, userID int64,
) (*virtualkeydomain.VirtualAPIKey, error) {
	r.trace.add("key.get-scoped")
	r.scopedCalls = append(r.scopedCalls, scopedKeyCall{id: id, orgID: orgID, userID: userID})
	if r.scopedErr != nil {
		return nil, r.scopedErr
	}
	key := r.keys[id]
	if key == nil || key.OrganizationID != orgID || key.UserID != userID {
		return nil, nil
	}
	return key, nil
}

func (r *fakeVirtualKeyRepository) GetByHash(context.Context, string) (*virtualkeydomain.VirtualAPIKey, error) {
	return nil, nil
}

func (r *fakeVirtualKeyRepository) ListByScope(context.Context, int64, int64) ([]*virtualkeydomain.VirtualAPIKey, error) {
	return nil, nil
}

func (r *fakeVirtualKeyRepository) UpdateStatusForScope(
	_ context.Context,
	id, orgID, userID int64,
	status string,
) (bool, error) {
	r.statusCalls = append(r.statusCalls, scopedStatusCall{
		id: id, orgID: orgID, userID: userID, status: status,
	})
	key := r.keys[id]
	if key == nil || key.OrganizationID != orgID || key.UserID != userID {
		return false, nil
	}
	key.Status = status
	return true, nil
}

func (r *fakeVirtualKeyRepository) TouchLastUsed(_ context.Context, id int64) error {
	r.trace.add("key.touch")
	r.touchCalls = append(r.touchCalls, id)
	return r.touchErr
}

type visibleModelCall struct {
	id    int64
	actor airesourcesvc.Actor
	orgID int64
}

type fakeModelResourceValidator struct {
	err          error
	visibleCalls []visibleModelCall
	trace        *callTrace
}

func (r *fakeModelResourceValidator) EnsureSelectable(
	_ context.Context, actor airesourcesvc.Actor, orgID, resourceID int64,
) error {
	r.trace.add("resource.ensure-selectable")
	r.visibleCalls = append(r.visibleCalls, visibleModelCall{id: resourceID, actor: actor, orgID: orgID})
	return r.err
}

func int64Pointer(value int64) *int64 { return &value }
