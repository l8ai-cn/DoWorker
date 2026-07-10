package virtualkey

import (
	"context"

	aimodeldomain "github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	virtualkeydomain "github.com/anthropics/agentsmesh/backend/internal/domain/virtualkey"
)

type callTrace struct{ calls []string }

func (t *callTrace) add(call string) { t.calls = append(t.calls, call) }

type scopedKeyCall struct {
	id     int64
	orgID  int64
	userID int64
}

type fakeVirtualKeyRepository struct {
	keys        map[int64]*virtualkeydomain.VirtualAPIKey
	createErr   error
	scopedErr   error
	touchErr    error
	created     []*virtualkeydomain.VirtualAPIKey
	scopedCalls []scopedKeyCall
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

func (r *fakeVirtualKeyRepository) UpdateStatus(context.Context, int64, string) error { return nil }

func (r *fakeVirtualKeyRepository) TouchLastUsed(_ context.Context, id int64) error {
	r.trace.add("key.touch")
	r.touchCalls = append(r.touchCalls, id)
	return r.touchErr
}

type visibleModelCall struct {
	id     int64
	userID int64
	orgID  int64
}

type fakeAIModelRepository struct {
	models       map[int64]*aimodeldomain.AIModel
	visibleErr   error
	visibleCalls []visibleModelCall
	trace        *callTrace
}

func (r *fakeAIModelRepository) GetByID(_ context.Context, id int64) (*aimodeldomain.AIModel, error) {
	return r.models[id], nil
}

func (r *fakeAIModelRepository) GetVisibleByID(
	_ context.Context, id, userID, orgID int64,
) (*aimodeldomain.AIModel, error) {
	r.trace.add("model.get-visible")
	r.visibleCalls = append(r.visibleCalls, visibleModelCall{id: id, userID: userID, orgID: orgID})
	if r.visibleErr != nil {
		return nil, r.visibleErr
	}
	model := r.models[id]
	if model == nil || !model.IsEnabled {
		return nil, nil
	}
	if model.OrganizationID != nil && *model.OrganizationID == orgID {
		return model, nil
	}
	if model.UserID != nil && *model.UserID == userID {
		return model, nil
	}
	return nil, nil
}

func (r *fakeAIModelRepository) Create(context.Context, *aimodeldomain.AIModel) error { return nil }
func (r *fakeAIModelRepository) Save(context.Context, *aimodeldomain.AIModel) error   { return nil }
func (r *fakeAIModelRepository) Delete(context.Context, int64) error                  { return nil }
func (r *fakeAIModelRepository) ListVisible(context.Context, int64, int64) ([]*aimodeldomain.AIModel, error) {
	return nil, nil
}
func (r *fakeAIModelRepository) DefaultVisible(context.Context, int64, int64) (*aimodeldomain.AIModel, error) {
	return nil, nil
}
func (r *fakeAIModelRepository) ClearDefaults(context.Context, int64, int64) error { return nil }
func (r *fakeAIModelRepository) CountOrg(context.Context, int64) (int64, error)    { return 0, nil }
func (r *fakeAIModelRepository) FirstVisibleByProvider(
	context.Context, int64, int64, string,
) (*aimodeldomain.AIModel, error) {
	return nil, nil
}

func int64Pointer(value int64) *int64 { return &value }
