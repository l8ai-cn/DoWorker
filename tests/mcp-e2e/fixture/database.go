package fixture

import (
	"testing"

	"github.com/anthropics/agentsmesh/tests/mcp-e2e/client"
)

func OpenDB(t *testing.T, env *Env) *client.DB {
	t.Helper()
	db, err := client.OpenDB(env.PostgresDSN)
	if err != nil {
		t.Fatalf("open MCP E2E database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
