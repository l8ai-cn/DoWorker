package sessionapi

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/virtualkey"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	virtualkeysvc "github.com/l8ai-cn/agentcloud/backend/internal/service/virtualkey"
	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRevokeVirtualKeyIsScopedToTheTenantUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE virtual_api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			organization_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			model_resource_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			key_prefix TEXT NOT NULL,
			key_hash TEXT NOT NULL UNIQUE,
			token_budget INTEGER,
			status TEXT NOT NULL,
			last_used_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error)
	key := &virtualkey.VirtualAPIKey{
		OrganizationID: 21, UserID: 11, ModelResourceID: 31,
		Name: "Worker key", KeyPrefix: "dwk_12345678",
		KeyHash: "key-hash", Status: virtualkey.StatusActive,
	}
	require.NoError(t, db.Create(key).Error)
	deps := &Deps{
		VirtualKeys: virtualkeysvc.NewService(infra.NewVirtualAPIKeyRepository(db), nil),
	}

	wrongUser := revokeVirtualKey(t, deps, key.ID, 21, 12)
	assert.Equal(t, http.StatusNotFound, wrongUser.Code)
	require.NoError(t, db.First(key, key.ID).Error)
	assert.Equal(t, virtualkey.StatusActive, key.Status)

	owner := revokeVirtualKey(t, deps, key.ID, 21, 11)
	assert.Equal(t, http.StatusNoContent, owner.Code)
	require.NoError(t, db.First(key, key.ID).Error)
	assert.Equal(t, virtualkey.StatusRevoked, key.Status)
}

func revokeVirtualKey(t *testing.T, deps *Deps, keyID, orgID, userID int64) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
	ctx.Params = gin.Params{{Key: "id", Value: stringIDForTest(keyID)}}
	ctx.Set("tenant", &middleware.TenantContext{OrganizationID: orgID, UserID: userID})
	deps.handleRevokeVirtualKey(ctx)
	ctx.Writer.WriteHeaderNow()
	return recorder
}

func stringIDForTest(id int64) string {
	return strconv.FormatInt(id, 10)
}
