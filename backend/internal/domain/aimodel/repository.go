package aimodel

import "context"

// Repository persists model-pool rows. Visibility queries union org-shared and
// user-private rows for the caller.
type Repository interface {
	GetByID(ctx context.Context, id int64) (*AIModel, error)
	Create(ctx context.Context, m *AIModel) error
	Save(ctx context.Context, m *AIModel) error
	Delete(ctx context.Context, id int64) error

	// ListVisible returns enabled rows the (userID, orgID) pair can see:
	// org-shared rows for orgID plus user-private rows for userID.
	ListVisible(ctx context.Context, userID, orgID int64) ([]*AIModel, error)

	// DefaultVisible returns the default model for the pair, preferring a
	// user-private default over the org default. Nil when none configured.
	DefaultVisible(ctx context.Context, userID, orgID int64) (*AIModel, error)

	// ClearDefaults unsets is_default across the scope a new default belongs to
	// (org scope when orgID>0 & userID==0, else user scope).
	ClearDefaults(ctx context.Context, userID, orgID int64) error

	// CountOrg reports how many org-scoped rows exist (dev seed idempotency).
	CountOrg(ctx context.Context, orgID int64) (int64, error)

	// FirstVisibleByProvider returns the best visible row for a provider
	// (default first). Nil when none match.
	FirstVisibleByProvider(ctx context.Context, userID, orgID int64, providerType string) (*AIModel, error)
}
