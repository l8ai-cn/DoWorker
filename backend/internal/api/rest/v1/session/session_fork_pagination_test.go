package sessionapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	itemDomain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	sessionService "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	itemService "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCopyConversationItemsCopiesEveryPage(t *testing.T) {
	db, deps := setupForkPaginationTest(t)
	seedForkPaginationSessions(t, deps)
	for position := int64(1); position <= 125; position++ {
		require.NoError(t, deps.Items.Append(context.Background(), &itemDomain.Item{
			ID: fmt.Sprintf("item_%03d", position), SessionID: "conv_source",
			ItemType: "message", ResponseID: fmt.Sprintf("resp_%03d", position),
			Status: "completed", Position: position,
			Payload: []byte(`{"type":"message"}`), CreatedAt: time.Now(),
		}))
	}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("POST", "/v1/sessions/conv_source/fork", nil)

	err := deps.copyConversationItems(ctx, deps.Items, "conv_source", "conv_dest", nil)

	require.NoError(t, err)
	var count int64
	require.NoError(t, db.Model(&itemDomain.Item{}).
		Where("session_id = ?", "conv_dest").
		Count(&count).Error)
	assert.Equal(t, int64(125), count)
}

func TestCopyConversationItemsRewritesPayloadItemID(t *testing.T) {
	db, deps := setupForkPaginationTest(t)
	seedForkPaginationSessions(t, deps)
	require.NoError(t, deps.Items.Append(context.Background(), &itemDomain.Item{
		ID: "item_source", SessionID: "conv_source", ItemType: "message",
		ResponseID: "resp_source", Status: "completed", Position: 1,
		Payload:   []byte(`{"id":"item_source","type":"message","exact":9007199254740993}`),
		CreatedAt: time.Now(),
	}))
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("POST", "/v1/sessions/conv_source/fork", nil)

	require.NoError(t, deps.copyConversationItems(ctx, deps.Items, "conv_source", "conv_dest", nil))

	var copied itemDomain.Item
	require.NoError(t, db.Where("session_id = ?", "conv_dest").First(&copied).Error)
	var payload map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(copied.Payload, &payload))
	assert.JSONEq(t, `"`+copied.ID+`"`, string(payload["id"]))
	assert.Equal(t, "9007199254740993", string(payload["exact"]))
}

func TestCopyConversationItemsIncludesWholeTargetResponse(t *testing.T) {
	db, deps := setupForkPaginationTest(t)
	seedForkPaginationSessions(t, deps)
	for position, responseID := range []string{"resp_before", "resp_target", "resp_target", "resp_after"} {
		appendForkPaginationItem(t, deps, int64(position+1), responseID)
	}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("POST", "/v1/sessions/conv_source/fork", nil)
	target := "resp_target"

	require.NoError(t, deps.copyConversationItems(ctx, deps.Items, "conv_source", "conv_dest", &target))

	var count int64
	require.NoError(t, db.Model(&itemDomain.Item{}).
		Where("session_id = ?", "conv_dest").
		Count(&count).Error)
	assert.Equal(t, int64(3), count)
}

func TestCopyConversationItemsRejectsMissingTargetResponse(t *testing.T) {
	_, deps := setupForkPaginationTest(t)
	seedForkPaginationSessions(t, deps)
	appendForkPaginationItem(t, deps, 1, "resp_existing")
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("POST", "/v1/sessions/conv_source/fork", nil)
	target := "resp_missing"

	err := deps.copyConversationItems(ctx, deps.Items, "conv_source", "conv_dest", &target)

	require.Error(t, err)
}

func setupForkPaginationTest(t *testing.T) (*gorm.DB, *Deps) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/fork-pagination.db"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&sessionDomain.Session{}, &itemDomain.Item{}))
	return db, &Deps{
		Sessions: sessionService.NewService(db),
		Items:    itemService.NewService(db),
	}
}

func seedForkPaginationSessions(t *testing.T, deps *Deps) {
	t.Helper()
	for _, row := range []*sessionDomain.Session{
		{ID: "conv_source", OrganizationID: 21, UserID: 11, PodKey: "source-pod", AgentSlug: "codex-cli"},
		{ID: "conv_dest", OrganizationID: 21, UserID: 11, PodKey: "dest-pod", AgentSlug: "codex-cli"},
	} {
		require.NoError(t, deps.Sessions.Create(context.Background(), row))
	}
}

func appendForkPaginationItem(t *testing.T, deps *Deps, position int64, responseID string) {
	t.Helper()
	require.NoError(t, deps.Items.Append(context.Background(), &itemDomain.Item{
		ID: fmt.Sprintf("item_%03d", position), SessionID: "conv_source",
		ItemType: "message", ResponseID: responseID, Status: "completed",
		Position: position, Payload: []byte(`{"type":"message"}`), CreatedAt: time.Now(),
	}))
}
