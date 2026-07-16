package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMarketplaceInstallationRouteClonesExpert(t *testing.T) {
	installer := &marketplaceInstallerStub{}
	authorizer := &marketplaceAuthorizerStub{allowed: true}
	router := gin.New()
	RegisterMarketplaceInstallationRoutes(
		router.Group("/api/internal/marketplace/installations"),
		MarketplaceInstallationDeps{
			Installer: installer, Authorizer: authorizer, InternalSecret: "secret",
		},
	)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/internal/marketplace/installations/apply",
		strings.NewReader(`{
		  "installation_id":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		  "platform_resource_type":"expert",
		  "platform_resource_id":101,
		  "source_release_id":201,
		  "target_platform_organization_id":9,
		  "actor_platform_user_id":14,
		  "runtime_snapshot":{"market_application_slug":"software-delivery-expert"},
		  "configuration":{
		    "model_resource_id":301,
		    "tool_model_resource_ids":{"seedance-video":302}
		  }
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Internal-Secret", "secret")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{
	  "runtime_ref":"expert:201",
	  "result":{"expert_id":"201","already_installed":false}
	}`, response.Body.String())
	require.Equal(t, int64(9), installer.request.TargetOrganizationID)
	require.Equal(t, int64(301), installer.request.ModelResourceID)
	require.Equal(
		t,
		map[string]int64{"seedance-video": 302},
		installer.request.ToolModelResourceIDs,
	)
	require.Equal(t, int64(101), installer.request.SourceMarketApplicationID)
	require.Equal(t, int64(201), installer.request.SourceMarketReleaseID)
	require.JSONEq(t, `{"market_application_slug":"software-delivery-expert"}`,
		string(installer.request.RuntimeSnapshot))
}

func TestMarketplaceAuthorizationRouteChecksMembership(t *testing.T) {
	router := gin.New()
	RegisterMarketplaceInstallationRoutes(
		router.Group("/api/internal/marketplace/installations"),
		MarketplaceInstallationDeps{
			Installer:      &marketplaceInstallerStub{},
			Authorizer:     &marketplaceAuthorizerStub{allowed: true},
			InternalSecret: "secret",
		},
	)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/internal/marketplace/installations/authorize",
		strings.NewReader(`{
		  "target_platform_organization_id":9,
		  "actor_platform_user_id":14
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Internal-Secret", "secret")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
}

type marketplaceInstallerStub struct {
	request expertsvc.MarketplaceInstallationRequest
}

func (s *marketplaceInstallerStub) InstallMarketplaceExpert(
	_ context.Context,
	request expertsvc.MarketplaceInstallationRequest,
) (*expertdom.Expert, bool, error) {
	s.request = request
	return &expertdom.Expert{ID: 201}, false, nil
}

type marketplaceAuthorizerStub struct {
	allowed bool
}

func (s *marketplaceAuthorizerStub) IsMember(
	context.Context,
	int64,
	int64,
) (bool, error) {
	return s.allowed, nil
}
