package sessionapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGetSessionByPodKeyUsesNoContentForAbsentAssociation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSessionByPodTestDB(t)
	deps := &Deps{Sessions: sessionsvc.NewService(db)}

	missing := getSessionByPodKey(t, deps, "missing-pod", 21, 11)

	assert.Equal(t, http.StatusNoContent, missing.Code)
	assert.Empty(t, missing.Body.String())
}

func TestGetSessionByPodKeyDoesNotExposeAnotherUsersAssociation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSessionByPodTestDB(t)
	insertSessionByPodTestRow(t, db, "conv_private", "private-pod", 21, 11)
	deps := &Deps{Sessions: sessionsvc.NewService(db)}

	response := getSessionByPodKey(t, deps, "private-pod", 21, 12)

	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Empty(t, response.Body.String())
}

func TestGetSessionByPodKeyReturnsVisibleAssociation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSessionByPodTestDB(t)
	insertSessionByPodTestRow(t, db, "conv_owner", "owner-pod", 21, 11)
	deps := &Deps{Sessions: sessionsvc.NewService(db)}

	response := getSessionByPodKey(t, deps, "owner-pod", 21, 11)

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `"conv_owner"`, responseBodyField(t, response, "id"))
}

func TestGetSessionByPodKeySurfacesLookupFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSessionByPodTestDB(t)
	require.NoError(t, db.Exec("DROP TABLE agent_sessions").Error)
	deps := &Deps{Sessions: sessionsvc.NewService(db)}

	response := getSessionByPodKey(t, deps, "owner-pod", 21, 11)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func setupSessionByPodTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE agent_sessions (
			id TEXT PRIMARY KEY,
			organization_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			pod_key TEXT NOT NULL UNIQUE,
			agent_slug TEXT NOT NULL,
			runner_node_id TEXT,
			title TEXT,
			status TEXT NOT NULL DEFAULT 'idle',
			parent_session_id TEXT,
			project TEXT,
			archived BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at DATETIME,
			mcp_servers TEXT NOT NULL DEFAULT '[]',
			codex_goal TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error)
	return db
}

func insertSessionByPodTestRow(
	t *testing.T,
	db *gorm.DB,
	id string,
	podKey string,
	orgID int64,
	userID int64,
) {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO agent_sessions (id, organization_id, user_id, pod_key, agent_slug) VALUES (?, ?, ?, ?, ?)`,
		id,
		orgID,
		userID,
		podKey,
		"codex-cli",
	).Error)
}

func getSessionByPodKey(
	t *testing.T,
	deps *Deps,
	podKey string,
	orgID int64,
	userID int64,
) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/sessions/by-pod/"+podKey, nil)
	ctx.Params = gin.Params{{Key: "pod_key", Value: podKey}}
	ctx.Set("tenant", &middleware.TenantContext{OrganizationID: orgID, UserID: userID})
	deps.handleGetSessionByPodKey(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}

func responseBodyField(t *testing.T, response *httptest.ResponseRecorder, field string) string {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	value, ok := body[field]
	require.True(t, ok)
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	return string(encoded)
}
