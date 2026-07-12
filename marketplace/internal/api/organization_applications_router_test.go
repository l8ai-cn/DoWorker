package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authpkg "github.com/anthropics/agentsmesh/backend/pkg/auth"
	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"github.com/stretchr/testify/require"
)

func TestRouterRegistersOrganizationApplications(t *testing.T) {
	router := NewRouter(Dependencies{
		Ready:         func(context.Context) error { return nil },
		Storefront:    testStorefront(),
		Identity:      organizationApplicationsTokenVerifier{},
		Installations: healthInstallations{},
		Applications: organizationApplicationsStub{
			items: []service.OrganizationApplication{{
				InstallationID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
				Status:         "active",
			}},
		},
	})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/marketplace/v1/organizations/9/applications",
		nil,
	)
	request.Header.Set("Authorization", "Bearer token")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{
	  "applications":[{
	    "installation_id":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	    "market_slug":"",
	    "listing_slug":"",
	    "display_name":"",
	    "tagline":"",
	    "resource_type":"",
	    "outcomes":null,
	    "runtime_ref":"",
	    "status":"active",
	    "installed_at":"0001-01-01T00:00:00Z"
	  }]
	}`, response.Body.String())
}

type organizationApplicationsTokenVerifier struct{}

func (organizationApplicationsTokenVerifier) Verify(
	context.Context,
	string,
) (*authpkg.Claims, error) {
	return &authpkg.Claims{UserID: 14}, nil
}

type organizationApplicationsStub struct {
	items []service.OrganizationApplication
}

func (s organizationApplicationsStub) ListOrganizationApplications(
	context.Context,
	int64,
	int64,
) ([]service.OrganizationApplication, error) {
	return s.items, nil
}
