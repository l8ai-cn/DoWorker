package agentsession_test

import (
	"context"
	"testing"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	itemDomain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	svc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.Session{}))
	require.NoError(t, db.AutoMigrate(&itemDomain.Item{}))
	return db
}

func TestNewID(t *testing.T) {
	id, err := svc.NewID()
	require.NoError(t, err)
	assert.Regexp(t, `^conv_[a-f0-9]{16}$`, id)
}

func TestCreateAndGet(t *testing.T) {
	s := svc.NewService(testDB(t))
	ctx := context.Background()
	id, err := svc.NewID()
	require.NoError(t, err)
	row := &domain.Session{
		ID: id, OrganizationID: 1, UserID: 2,
		PodKey: "2-standalone-abc", AgentSlug: "do-agent", Status: "idle",
	}
	require.NoError(t, s.Create(ctx, row))
	got, err := s.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "do-agent", got.AgentSlug)
}
