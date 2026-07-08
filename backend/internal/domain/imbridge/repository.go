package imbridge

import "context"

type Repository interface {
	ListConnections(ctx context.Context, orgID int64) ([]*Connection, error)
	ListActiveByProvider(ctx context.Context, provider string) ([]*Connection, error)
	GetConnection(ctx context.Context, orgID, id int64) (*Connection, error)
	GetConnectionByToken(ctx context.Context, provider, token string) (*Connection, error)
	CreateConnection(ctx context.Context, conn *Connection) error
	UpdateConnection(ctx context.Context, conn *Connection) error
	DeleteConnection(ctx context.Context, orgID, id int64) error

	GetThreadMapping(ctx context.Context, connectionID int64, externalThreadID string) (*ThreadMapping, error)
	GetThreadMappingByChannel(ctx context.Context, connectionID, channelID int64) (*ThreadMapping, error)
	UpsertThreadMapping(ctx context.Context, mapping *ThreadMapping) error
}
