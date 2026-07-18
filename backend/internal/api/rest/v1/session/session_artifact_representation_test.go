package sessionapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestArtifactRepresentationRequiresSessionOwner(t *testing.T) {
	deps := readOnlySessionPermissionDeps(t)
	response := artifactRepresentationRequest(t, deps, 12)

	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestArtifactRepresentationOwnerReachesArtifactService(t *testing.T) {
	deps := ownerSessionPermissionDeps(t)
	response := artifactRepresentationRequest(t, deps, 11)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestArtifactDownloadGrantScope(t *testing.T) {
	now := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)
	minimum := uint64(3)
	maximum := uint64(3)
	expires := now.Add(time.Minute).Format(time.RFC3339Nano)
	artifact := &agentworkbenchv2.ArtifactDescriptor{
		Revision: 3,
		Grants: []*agentworkbenchv2.ArtifactGrant{{
			GrantId:           "grant-1",
			Actions:           []string{"artifact.download"},
			RepresentationIds: []string{"preview"},
			MinimumRevision:   &minimum,
			MaximumRevision:   &maximum,
			ExpiresAt:         &expires,
		}},
	}

	assert.True(t, artifactDownloadGranted(artifact, "preview", now))
	assert.False(t, artifactDownloadGranted(artifact, "original", now))
	artifact.Grants[0].Actions = []string{"image.edit"}
	assert.False(t, artifactDownloadGranted(artifact, "preview", now))
	artifact.Grants[0].Actions = []string{"artifact.download"}
	assert.False(t, artifactDownloadGranted(
		artifact,
		"preview",
		now.Add(2*time.Minute),
	))
}

func artifactRepresentationRequest(
	t *testing.T,
	deps *Deps,
	userID int64,
) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/v1/sessions/conv_read/artifacts/representation",
		nil,
	)
	ctx.Params = gin.Params{{Key: "id", Value: "conv_read"}}
	ctx.Set("tenant", &middleware.TenantContext{
		OrganizationID: 21,
		UserID:         userID,
	})
	deps.handleGetSessionArtifactRepresentation(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}
