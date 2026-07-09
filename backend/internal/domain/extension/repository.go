package extension

import (
	"context"
	"errors"
)

// Domain-level repository errors
var (
	// ErrDuplicateInstall is returned when attempting to install a skill or MCP server
	// that already exists with the same unique key (org + repo + scope + user + slug).
	ErrDuplicateInstall = errors.New("already installed with the same slug in this scope")
)

// McpMarketRepository owns the MCP server catalog populated by the
// upstream registry syncer.
type McpMarketRepository interface {
	ListMcpMarketItems(ctx context.Context, query string, category string, limit, offset int) ([]*McpMarketItem, int64, error)
	GetMcpMarketItem(ctx context.Context, id int64) (*McpMarketItem, error)
	FindMcpMarketItemByRegistryName(ctx context.Context, registryName string) (*McpMarketItem, error)
	UpsertMcpMarketItem(ctx context.Context, item *McpMarketItem) error
	BatchUpsertMcpMarketItems(ctx context.Context, items []*McpMarketItem) error
	DeactivateMcpMarketItemsNotIn(ctx context.Context, sourceType string, registryNames []string) (int64, error)
}

// InstalledMcpRepository owns per-repository installations of MCP servers.
type InstalledMcpRepository interface {
	ListInstalledMcpServers(ctx context.Context, orgID, repoID, userID int64, scope string) ([]*InstalledMcpServer, error)
	GetInstalledMcpServer(ctx context.Context, id int64) (*InstalledMcpServer, error)
	CreateInstalledMcpServer(ctx context.Context, server *InstalledMcpServer) error
	UpdateInstalledMcpServer(ctx context.Context, server *InstalledMcpServer) error
	DeleteInstalledMcpServer(ctx context.Context, id int64) error
	GetEffectiveMcpServers(ctx context.Context, orgID, userID, repoID int64) ([]*InstalledMcpServer, error)
}

// InstalledSkillRepository owns per-repository installations of skills.
type InstalledSkillRepository interface {
	ListInstalledSkills(ctx context.Context, orgID, repoID, userID int64, scope string) ([]*InstalledSkill, error)
	GetInstalledSkill(ctx context.Context, id int64) (*InstalledSkill, error)
	CreateInstalledSkill(ctx context.Context, skill *InstalledSkill) error
	UpdateInstalledSkill(ctx context.Context, skill *InstalledSkill) error
	DeleteInstalledSkill(ctx context.Context, id int64) error
	GetEffectiveSkills(ctx context.Context, orgID, userID, repoID int64) ([]*InstalledSkill, error)
}

// Repository is the union of the focused sub-interfaces. Callers should
// prefer to depend on the narrowest sub-interface they actually need —
// this aggregate exists for wiring (single concrete impl) and for tests
// that need to mock the full surface.
type Repository interface {
	McpMarketRepository
	InstalledMcpRepository
	InstalledSkillRepository
}
