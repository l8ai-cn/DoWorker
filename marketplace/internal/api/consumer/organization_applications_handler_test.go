package consumer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	actorapi "github.com/l8ai-cn/agentcloud/marketplace/internal/api/actor"
	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestListOrganizationApplicationsUsesAuthenticatedActor(t *testing.T) {
	applications := &organizationApplicationsStub{
		items: []service.OrganizationApplication{
			{
				InstallationID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
				MarketSlug:     "agent-cloud-market",
				ListingSlug:    "software-delivery-expert",
				DisplayName:    "软件交付专家",
				Tagline:        "把需求变成可验证的代码交付",
				ResourceType:   "application",
				Outcomes:       []string{"执行关键路径验证"},
				RuntimeRef:     "expert:12",
				Status:         "active",
				InstalledAt:    time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC),
			},
		},
	}
	router := authenticatedOrganizationApplicationsRouter(applications)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/marketplace/v1/organizations/9/applications",
		nil,
	)
	request.Header.Set("Authorization", "Bearer token")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, int64(9), applications.organizationID)
	require.Equal(t, int64(14), applications.actorUserID)
	require.JSONEq(t, `{
	  "applications":[{
	    "installation_id":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	    "market_slug":"agent-cloud-market",
	    "listing_slug":"software-delivery-expert",
	    "display_name":"软件交付专家",
	    "tagline":"把需求变成可验证的代码交付",
	    "resource_type":"application",
	    "outcomes":["执行关键路径验证"],
	    "runtime_ref":"expert:12",
	    "status":"active",
	    "installed_at":"2026-07-12T08:00:00Z"
	  }]
	}`, response.Body.String())
}

func TestListOrganizationApplicationsReturnsStableForbiddenError(t *testing.T) {
	applications := &organizationApplicationsStub{
		err: service.ErrTargetOrganizationForbidden,
	}
	router := authenticatedOrganizationApplicationsRouter(applications)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/marketplace/v1/organizations/9/applications",
		nil,
	)
	request.Header.Set("Authorization", "Bearer token")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusForbidden, response.Code)
	require.JSONEq(t, `{
	  "error":{
	    "code":"TARGET_ORGANIZATION_FORBIDDEN",
	    "message":"你无权在这个组织中启用应用"
	  }
	}`, response.Body.String())
}

func authenticatedOrganizationApplicationsRouter(
	applications OrganizationApplicationsReader,
) *gin.Engine {
	router := gin.New()
	group := router.Group("/api/marketplace/v1")
	group.Use(actorapi.Middleware(tokenVerifierStub{}))
	NewOrganizationApplicationsHandler(applications).RegisterRoutes(group)
	return router
}

type organizationApplicationsStub struct {
	items          []service.OrganizationApplication
	err            error
	organizationID int64
	actorUserID    int64
}

func (s *organizationApplicationsStub) ListOrganizationApplications(
	_ context.Context,
	organizationID int64,
	actorUserID int64,
) ([]service.OrganizationApplication, error) {
	s.organizationID = organizationID
	s.actorUserID = actorUserID
	return s.items, s.err
}
