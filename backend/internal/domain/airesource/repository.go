package airesource

import (
	"context"
	"errors"
	"time"
)

var ErrConflict = errors.New("AI resource identifier already exists in this scope")

type Repository interface {
	GetConnectionByID(ctx context.Context, id int64) (*Connection, error)
	CreateConnection(ctx context.Context, connection *Connection) error
	SaveConnection(ctx context.Context, connection *Connection) error
	DeleteConnection(ctx context.Context, id, expectedRevision int64) error
	ListConnectionsByOwner(ctx context.Context, scope OwnerScope, ownerID int64) ([]*Connection, error)

	GetResourceByID(ctx context.Context, id int64) (*ModelResource, error)
	CreateResource(ctx context.Context, resource *ModelResource) error
	SaveResource(ctx context.Context, resource *ModelResource) error
	DeleteResource(ctx context.Context, id, expectedRevision int64) error
	ListResourcesByConnection(ctx context.Context, connectionID int64) ([]*ModelResource, error)
	ListResourcesByOwner(ctx context.Context, scope OwnerScope, ownerID int64) ([]*ModelResource, error)
	ListEffective(ctx context.Context, userID, orgID int64, modalities []Modality) ([]*ModelResource, error)
	SetDefault(ctx context.Context, resourceID int64, modality Modality) error
	SetValidationState(ctx context.Context, connectionID, expectedRevision int64, status ConnectionStatus, at time.Time, validationError string) (int64, error)
}
