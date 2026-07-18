package sessionapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	permissiondomain "github.com/anthropics/agentsmesh/backend/internal/domain/sessionpermission"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	permissionservice "github.com/anthropics/agentsmesh/backend/internal/service/sessionpermission"
	"github.com/anthropics/agentsmesh/backend/pkg/embedtoken"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCreateEmbedContextIssuesSessionBoundToken(t *testing.T) {
	deps := embedContextDeps(t)
	response := embedContextRequest(t, deps, 11, `{
		"parent_origins":["https://portal.example"],
		"capabilities":["read","write"]
	}`)

	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	var body struct {
		EmbedContext    string `json:"embed_context"`
		RedemptionProof string `json:"redemption_proof"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.NotEmpty(t, body.RedemptionProof)
	claims, err := deps.EmbedTokens.ValidateContext(body.EmbedContext)
	require.NoError(t, err)
	assert.Equal(t, "conv_embed", claims.SessionID)
	assert.Equal(t, int64(21), claims.OrganizationID)
	assert.Equal(t, []string{"https://portal.example"}, claims.AllowedParentOrigins)
	assert.Equal(t, []string{"read", "write"}, claims.Capabilities)
}

func TestCreateEmbedContextAcceptsAgentWorkspaceCapabilities(t *testing.T) {
	deps := embedContextDeps(t)
	response := embedContextRequest(t, deps, 11, `{
		"parent_origins":["https://portal.example"],
		"capabilities":["read","write","approve","terminal","control"]
	}`)

	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	var body struct {
		EmbedContext string `json:"embed_context"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	claims, err := deps.EmbedTokens.ValidateContext(body.EmbedContext)
	require.NoError(t, err)
	assert.Equal(
		t,
		[]string{"read", "write", "approve", "terminal", "control"},
		claims.Capabilities,
	)
}

func TestCreateEmbedContextRejectsInvalidOrigins(t *testing.T) {
	for _, body := range []string{
		`{"parent_origins":["*"],"capabilities":["read"]}`,
		`{"parent_origins":["https://portal.example/path"],"capabilities":["read"]}`,
		`{"parent_origins":["https://portal.example","https://portal.example"],"capabilities":["read"]}`,
	} {
		t.Run(body, func(t *testing.T) {
			response := embedContextRequest(t, embedContextDeps(t), 11, body)
			assert.Equal(t, http.StatusBadRequest, response.Code)
		})
	}
}

func TestCreateEmbedContextRequiresManagePermission(t *testing.T) {
	deps := embedContextDeps(t)
	_, err := deps.SessionPermissions.Upsert(t.Context(), "conv_embed", "__public__", levelEdit)
	require.NoError(t, err)

	response := embedContextRequest(t, deps, 12, `{
		"parent_origins":["https://portal.example"],
		"capabilities":["read"]
	}`)

	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestRedeemEmbedContextReturnsSessionToken(t *testing.T) {
	deps := embedContextDeps(t)
	grant, err := deps.EmbedTokens.IssueContext(t.Context(), embedtoken.ContextInput{
		SessionID:            "conv_embed",
		OrganizationID:       21,
		OrganizationSlug:     "acme",
		UserID:               11,
		Capabilities:         []string{"read"},
		AllowedParentOrigins: []string{"https://portal.example"},
	})
	require.NoError(t, err)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/v1/embed-contexts/redeem",
		bytes.NewBufferString(`{"redemption_proof":"`+grant.RedemptionProof+`"}`),
	)
	ctx.Request.Header.Set("Authorization", "Bearer "+grant.Token)
	ctx.Request.Header.Set("Content-Type", "application/json")

	deps.handleRedeemEmbedContext(ctx)
	ctx.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var body struct {
		AccessToken   string   `json:"access_token"`
		SessionID     string   `json:"session_id"`
		Capabilities  []string `json:"capabilities"`
		ParentOrigins []string `json:"parent_origins"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, "conv_embed", body.SessionID)
	assert.Equal(t, []string{"read"}, body.Capabilities)
	assert.Equal(t, []string{"https://portal.example"}, body.ParentOrigins)
	claims, err := deps.EmbedTokens.ValidateSession(body.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "conv_embed", claims.SessionID)
}

func TestInspectEmbedContextReturnsOriginsWithoutRedeeming(t *testing.T) {
	deps := embedContextDeps(t)
	grant, err := deps.EmbedTokens.IssueContext(t.Context(), embedtoken.ContextInput{
		SessionID:            "conv_embed",
		OrganizationID:       21,
		OrganizationSlug:     "acme",
		UserID:               11,
		Capabilities:         []string{"read"},
		AllowedParentOrigins: []string{"https://portal.example"},
	})
	require.NoError(t, err)
	router := gin.New()
	registerEmbedContextRoutes(router.Group("/v1"), *deps)

	response := embedSessionRequest(
		router,
		http.MethodPost,
		"/v1/embed-contexts/inspect",
		grant.Token,
	)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var body struct {
		ParentOrigins []string `json:"parent_origins"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, []string{"https://portal.example"}, body.ParentOrigins)
	_, _, err = deps.EmbedTokens.RedeemContext(
		t.Context(),
		grant.Token,
		grant.RedemptionProof,
	)
	require.NoError(t, err)
}

func TestWrongRedemptionProofDoesNotConsumeContext(t *testing.T) {
	deps := embedContextDeps(t)
	grant, err := deps.EmbedTokens.IssueContext(t.Context(), embedtoken.ContextInput{
		SessionID:            "conv_embed",
		OrganizationID:       21,
		OrganizationSlug:     "acme",
		UserID:               11,
		Capabilities:         []string{"read"},
		AllowedParentOrigins: []string{"https://portal.example"},
	})
	require.NoError(t, err)
	router := gin.New()
	registerEmbedContextRoutes(router.Group("/v1"), *deps)

	wrong := embedContextRedemptionRequest(router, grant.Token, "wrong-proof")
	assert.Equal(t, http.StatusUnauthorized, wrong.Code)
	correct := embedContextRedemptionRequest(
		router,
		grant.Token,
		grant.RedemptionProof,
	)
	assert.Equal(t, http.StatusOK, correct.Code, correct.Body.String())
}

func TestEmbedSessionTokenCanReadOnlyItsSession(t *testing.T) {
	deps, db := embedContextDepsWithDB(t)
	insertSessionByPodTestRow(t, db, "conv_other", "other-pod", 21, 11)
	router := gin.New()
	router.GET("/v1/embed/sessions/:id", deps.embedSessionAuth(), func(c *gin.Context) {
		_, _, ok := deps.authorizeSession(c, c.Param("id"))
		if ok {
			c.Status(http.StatusNoContent)
		}
	})
	token := embedSessionToken(t, deps, []string{"read"})

	allowed := embedSessionRequest(router, http.MethodGet, "/v1/embed/sessions/conv_embed", token)
	assert.Equal(t, http.StatusNoContent, allowed.Code)

	denied := embedSessionRequest(router, http.MethodGet, "/v1/embed/sessions/conv_other", token)
	assert.Equal(t, http.StatusNotFound, denied.Code)
}

func TestEmbedSessionTokenWithoutWriteCannotEnterEventRoute(t *testing.T) {
	deps := embedContextDeps(t)
	router := gin.New()
	router.POST(
		"/v1/embed/sessions/:id/events",
		deps.embedSessionAuth(),
		requireEmbedCapability("write"),
		func(c *gin.Context) { c.Status(http.StatusNoContent) },
	)

	response := embedSessionRequest(
		router,
		http.MethodPost,
		"/v1/embed/sessions/conv_embed/events",
		embedSessionToken(t, deps, []string{"read"}),
	)

	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestRegisterEmbedRoutesServesOnlyTheTokenSession(t *testing.T) {
	deps := embedContextDeps(t)
	router := gin.New()
	registerEmbedRoutes(router.Group("/v1"), *deps)
	token := embedSessionToken(t, deps, []string{"read"})

	response := embedSessionRequest(
		router,
		http.MethodGet,
		"/v1/embed/sessions/conv_embed",
		token,
	)

	assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
}

func TestEmbedReadCapabilityReachesSessionArtifactRoute(t *testing.T) {
	deps := embedContextDeps(t)
	router := gin.New()
	registerEmbedRoutes(router.Group("/v1"), *deps)

	response := embedSessionRequest(
		router,
		http.MethodGet,
		"/v1/embed/sessions/conv_embed/resources/files/file-1/content",
		embedSessionToken(t, deps, []string{"read"}),
	)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code, response.Body.String())
}

func TestEmbedReadCapabilityReachesWorkspaceArtifactContentRoute(t *testing.T) {
	deps := embedContextDeps(t)
	router := gin.New()
	registerEmbedRoutes(router.Group("/v1"), *deps)

	response := embedSessionRequest(
		router,
		http.MethodGet,
		"/v1/embed/sessions/conv_embed/resources/environments/workspace/artifacts/content/output/demo.mp4",
		embedSessionToken(t, deps, []string{"read"}),
	)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code, response.Body.String())
}

func TestEmbedWorkspaceRoutesRequireTheirExplicitCapabilities(t *testing.T) {
	deps := embedContextDeps(t)
	router := gin.New()
	registerEmbedRoutes(router.Group("/v1"), *deps)

	approval := embedSessionRequest(
		router,
		http.MethodPost,
		"/v1/embed/sessions/conv_embed/elicitations/elicit_1/resolve",
		embedSessionToken(t, deps, []string{"read", "write"}),
	)
	assert.Equal(t, http.StatusForbidden, approval.Code)

	terminal := embedSessionRequest(
		router,
		http.MethodGet,
		"/v1/embed/sessions/conv_embed/resources/terminals",
		embedSessionToken(t, deps, []string{"read"}),
	)
	assert.Equal(t, http.StatusForbidden, terminal.Code)

	control := embedSessionRequest(
		router,
		http.MethodGet,
		"/v1/embed/sessions/conv_embed/relay-connection",
		embedSessionToken(t, deps, []string{"read", "terminal"}),
	)
	assert.Equal(t, http.StatusForbidden, control.Code)

	acpControl := embedSessionRequest(
		router,
		http.MethodGet,
		"/v1/embed/sessions/conv_embed/acp-relay-connection",
		embedSessionToken(t, deps, []string{"read", "terminal"}),
	)
	assert.Equal(t, http.StatusForbidden, acpControl.Code)
}

func embedContextDeps(t *testing.T) *Deps {
	t.Helper()
	deps, _ := embedContextDepsWithDB(t)
	return deps
}

func embedContextDepsWithDB(t *testing.T) (*Deps, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := setupSessionByPodTestDB(t)
	insertSessionByPodTestRow(t, db, "conv_embed", "embed-pod", 21, 11)
	require.NoError(t, db.AutoMigrate(&permissiondomain.Grant{}))
	return &Deps{
		Sessions:           sessionsvc.NewService(db),
		SessionPermissions: permissionservice.NewService(db),
		EmbedTokens:        embedtoken.NewService("test-secret", embedContextRedis(t)),
	}, db
}

func embedSessionToken(t *testing.T, deps *Deps, capabilities []string) string {
	t.Helper()
	grant, err := deps.EmbedTokens.IssueContext(t.Context(), embedtoken.ContextInput{
		SessionID:            "conv_embed",
		OrganizationID:       21,
		OrganizationSlug:     "acme",
		UserID:               11,
		Capabilities:         capabilities,
		AllowedParentOrigins: []string{"https://portal.example"},
	})
	require.NoError(t, err)
	accessToken, _, err := deps.EmbedTokens.RedeemContext(
		t.Context(),
		grant.Token,
		grant.RedemptionProof,
	)
	require.NoError(t, err)
	return accessToken
}

func embedContextRedis(t *testing.T) *redis.Client {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func embedSessionRequest(router *gin.Engine, method, path, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func embedContextRedemptionRequest(
	router *gin.Engine,
	token string,
	proof string,
) *httptest.ResponseRecorder {
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/embed-contexts/redeem",
		bytes.NewBufferString(`{"redemption_proof":"`+proof+`"}`),
	)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func embedContextRequest(t *testing.T, deps *Deps, userID int64, body string) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/v1/sessions/conv_embed/embed-context",
		bytes.NewBufferString(body),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: "conv_embed"}}
	ctx.Set("tenant", &middleware.TenantContext{
		OrganizationID: 21, OrganizationSlug: "acme", UserID: userID, UserRole: "member",
	})
	deps.handleCreateEmbedContext(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}
