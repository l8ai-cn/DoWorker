package tokenquota

import "context"

// Repository persists token quota rows scoped to an organization.
type Repository interface {
	// Upsert creates or updates the quota for a scope, matching on
	// (organization_id, user_id, model) with NULL treated as a wildcard slot.
	Upsert(ctx context.Context, q *TokenQuota) error
	ListByOrg(ctx context.Context, orgID int64) ([]*TokenQuota, error)
	Delete(ctx context.Context, id, orgID int64) error
}
