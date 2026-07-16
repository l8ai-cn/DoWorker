package sessionapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	permissiondomain "github.com/anthropics/agentsmesh/backend/internal/domain/sessionpermission"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	permissionservice "github.com/anthropics/agentsmesh/backend/internal/service/sessionpermission"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadOnlySessionPermissionCannotPostEvent(t *testing.T) {
	deps := readOnlySessionPermissionDeps(t)

	response := sessionPermissionRequest(t, deps, "/v1/sessions/conv_read/events", `{"type":"interrupt"}`)

	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestReadOnlySessionPermissionCannotResolveElicitation(t *testing.T) {
	deps := readOnlySessionPermissionDeps(t)

	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/v1/sessions/conv_read/elicitations/elicit_1/resolve",
		bytes.NewBufferString(`{"action":"accept"}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{
		{Key: "id", Value: "conv_read"},
		{Key: "elicitation_id", Value: "elicit_1"},
	}
	ctx.Set("tenant", &middleware.TenantContext{OrganizationID: 21, UserID: 12})
	deps.handleResolveElicitation(ctx)
	ctx.Writer.WriteHeaderNow()

	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestReadOnlySessionPermissionCannotWriteFilesystem(t *testing.T) {
	deps := readOnlySessionPermissionDeps(t)

	response := sessionFilesystemWriteRequest(
		t,
		deps,
		`{"content":"blocked","encoding":"utf-8"}`,
		12,
	)

	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestSessionFilesystemWriteRejectsOversizedBody(t *testing.T) {
	deps := ownerSessionPermissionDeps(t)
	body := `{"content":"` + strings.Repeat("a", int(maxSessionFilesystemWriteBodyBytes)+1) +
		`","encoding":"utf-8"}`

	response := sessionFilesystemWriteRequest(t, deps, body, 11)

	assert.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
}

func readOnlySessionPermissionDeps(t *testing.T) *Deps {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := setupSessionByPodTestDB(t)
	insertSessionByPodTestRow(t, db, "conv_read", "read-pod", 21, 11)
	require.NoError(t, db.AutoMigrate(&permissiondomain.Grant{}))
	require.NoError(t, db.Create(&permissiondomain.Grant{
		SessionID: "conv_read",
		UserID:    "__public__",
		Level:     levelRead,
	}).Error)
	return &Deps{
		Sessions:           sessionsvc.NewService(db),
		SessionPermissions: permissionservice.NewService(db),
	}
}

func ownerSessionPermissionDeps(t *testing.T) *Deps {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := setupSessionByPodTestDB(t)
	insertSessionByPodTestRow(t, db, "conv_read", "read-pod", 21, 11)
	return &Deps{Sessions: sessionsvc.NewService(db)}
}

func sessionFilesystemWriteRequest(
	t *testing.T,
	deps *Deps,
	body string,
	userID int64,
) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodPut,
		"/v1/sessions/conv_read/resources/environments/workspace/filesystem/result.txt",
		bytes.NewBufferString(body),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{
		{Key: "id", Value: "conv_read"},
		{Key: "env", Value: "workspace"},
		{Key: "filepath", Value: "/result.txt"},
	}
	ctx.Set("tenant", &middleware.TenantContext{OrganizationID: 21, UserID: userID})
	deps.handleSessionFilesystemWrite(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}

func sessionPermissionRequest(t *testing.T, deps *Deps, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: "conv_read"}}
	ctx.Set("tenant", &middleware.TenantContext{OrganizationID: 21, UserID: 12})
	deps.handlePostEvent(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}
