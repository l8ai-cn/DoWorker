package sessionapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	airesourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	airesourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubModelResourceLister struct {
	resources []airesourcesvc.EffectiveResourceView
	err       error
}

func (s stubModelResourceLister) ListEffective(
	context.Context,
	airesourcesvc.Actor,
	int64,
	[]airesourcedomain.Modality,
) ([]airesourcesvc.EffectiveResourceView, error) {
	return s.resources, s.err
}

func TestListModelResourcesFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("requires tenant", func(t *testing.T) {
		response := listModelResources(t, &Deps{}, nil)
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	})

	t.Run("requires service", func(t *testing.T) {
		response := listModelResources(t, &Deps{}, tenantForModelResources())
		assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	})

	t.Run("surfaces query failure", func(t *testing.T) {
		deps := &Deps{AIResources: stubModelResourceLister{err: errors.New("database unavailable")}}
		response := listModelResources(t, deps, tenantForModelResources())
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})
}

func TestListModelResourcesReturnsOnlySelectableChatResources(t *testing.T) {
	deps := &Deps{
		AIResources: stubModelResourceLister{resources: []airesourcesvc.EffectiveResourceView{
			{
				Connection: airesourcesvc.ConnectionView{ProviderKey: slugkit.Slug("openai")},
				Resource: airesourcesvc.ResourceView{
					ID: 42, DisplayName: "Codex", ModelID: "gpt-5.5",
					DefaultModalities: []airesourcedomain.Modality{airesourcedomain.ModalityChat},
				},
				Selectable: true,
			},
			{
				Connection: airesourcesvc.ConnectionView{ProviderKey: slugkit.Slug("anthropic")},
				Resource:   airesourcesvc.ResourceView{ID: 43, DisplayName: "Disabled", ModelID: "claude"},
				Selectable: false,
			},
			{
				Connection: airesourcesvc.ConnectionView{ProviderKey: slugkit.Slug("openai")},
				Resource: airesourcesvc.ResourceView{
					ID: 44, DisplayName: "Embedding default", ModelID: "gpt-5",
					DefaultModalities: []airesourcedomain.Modality{airesourcedomain.ModalityEmbedding},
				},
				Selectable: true,
			},
		}},
	}

	response := listModelResources(t, deps, tenantForModelResources())

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{
		"object": "list",
		"data": [{
			"id": 42,
			"name": "Codex",
			"provider_key": "openai",
			"model": "gpt-5.5",
			"is_default": true
		}, {
			"id": 44,
			"name": "Embedding default",
			"provider_key": "openai",
			"model": "gpt-5",
			"is_default": false
		}]
	}`, response.Body.String())
}

func listModelResources(
	t *testing.T,
	deps *Deps,
	tenant *middleware.TenantContext,
) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/model-resources", nil)
	if tenant != nil {
		ctx.Set("tenant", tenant)
	}
	deps.handleListModelResources(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}

func tenantForModelResources() *middleware.TenantContext {
	return &middleware.TenantContext{UserID: 11, OrganizationID: 21}
}
