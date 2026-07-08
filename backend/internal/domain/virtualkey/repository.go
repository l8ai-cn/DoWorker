package virtualkey

import "context"

// Repository persists virtual API key rows. Scope queries union org-shared
// and user-private ownership for the caller.
type Repository interface {
	Create(ctx context.Context, k *VirtualAPIKey) error
	GetByID(ctx context.Context, id int64) (*VirtualAPIKey, error)
	GetByHash(ctx context.Context, hash string) (*VirtualAPIKey, error)
	ListByScope(ctx context.Context, orgID, userID int64) ([]*VirtualAPIKey, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	TouchLastUsed(ctx context.Context, id int64) error
}
