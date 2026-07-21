package sessionapi

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	sessionfiledomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/sessionfile"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/storage"
	filesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/file"
	sessionfilesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessionfile"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbedAttachmentUploadRequiresWriteCapability(t *testing.T) {
	deps := embedContextDeps(t)
	router := gin.New()
	registerEmbedRoutes(router.Group("/v1"), *deps)

	response := embedAttachmentRequest(
		t,
		router,
		"conv_embed",
		embedSessionToken(t, deps, []string{"read"}),
	)

	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestEmbedAttachmentUploadIsBoundToTokenSession(t *testing.T) {
	deps := embedContextDeps(t)
	router := gin.New()
	registerEmbedRoutes(router.Group("/v1"), *deps)

	response := embedAttachmentRequest(
		t,
		router,
		"conv_other",
		embedSessionToken(t, deps, []string{"read", "write"}),
	)

	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.JSONEq(
		t,
		`{"error":"session not found","code":"session_not_found"}`,
		response.Body.String(),
	)
}

func TestEmbedAttachmentUploadUsesSessionFileProtocol(t *testing.T) {
	deps, db := embedContextDepsWithDB(t)
	require.NoError(t, db.AutoMigrate(&sessionfiledomain.File{}))
	store := storage.NewMockStorage()
	deps.SessionFiles = sessionfilesvc.NewService(
		db,
		filesvc.NewService(store, config.StorageConfig{
			MaxFileSize:  10,
			AllowedTypes: []string{"text/plain"},
		}),
	)
	router := gin.New()
	registerEmbedRoutes(router.Group("/v1"), *deps)

	response := embedAttachmentRequest(
		t,
		router,
		"conv_embed",
		embedSessionToken(t, deps, []string{"read", "write"}),
	)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var body struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Object    string `json:"object"`
		SessionID string `json:"session_id"`
		Type      string `json:"type"`
		Metadata  struct {
			Bytes int64 `json:"bytes"`
		} `json:"metadata"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body.ID)
	assert.Equal(t, "notes.txt", body.Name)
	assert.Equal(t, "file", body.Object)
	assert.Equal(t, "conv_embed", body.SessionID)
	assert.Equal(t, "file", body.Type)
	assert.Equal(t, int64(len("attachment body")), body.Metadata.Bytes)
	assert.Equal(t, 1, store.FileCount())
}

func embedAttachmentRequest(
	t *testing.T,
	router *gin.Engine,
	sessionID string,
	token string,
) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "notes.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("attachment body"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/embed/sessions/"+sessionID+"/resources/files",
		&body,
	)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
