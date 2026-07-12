package executioncluster

import "context"

type Repository interface {
	ListByOrganization(ctx context.Context, organizationID int64) ([]*Cluster, error)
	GetByIDAndOrganization(ctx context.Context, id, organizationID int64) (*Cluster, error)
	EnsureDefaults(ctx context.Context, organizationID int64) ([]*Cluster, error)
}
