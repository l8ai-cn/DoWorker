package runtime

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"github.com/stretchr/testify/require"
)

func TestClientInstallsExpertThroughInternalBridge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		require.Equal(t, "secret", request.Header.Get("X-Internal-Secret"))
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), `"runtime_snapshot":{"market_application_slug":"software-delivery-expert"}`)
		require.Contains(t, string(body), `"actor_platform_user_id":14`)
		require.Contains(t, string(body), `"platform_resource_id":101`)
		require.Contains(t, string(body), `"source_release_id":201`)
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
		  "runtime_ref":"expert:201",
		  "result":{"expert_id":"201","already_installed":false}
		}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "secret", server.Client())
	require.NoError(t, err)

	result, err := client.Install(context.Background(), service.RuntimeInstallRequest{
		InstallationID:       "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		PlatformResourceType: "expert",
		PlatformResourceID:   101,
		SourceReleaseID:      201,
		RuntimeSnapshot:      []byte(`{"market_application_slug":"software-delivery-expert"}`),
		TargetOrganizationID: 9, ActorUserID: 14,
	})
	require.NoError(t, err)
	require.Equal(t, "expert:201", result.RuntimeRef)
}

func TestClientAuthorizesTargetOrganizationThroughInternalBridge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		require.Equal(t, "/authorize", request.URL.Path)
		require.Equal(t, "secret", request.Header.Get("X-Internal-Secret"))
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), `"target_platform_organization_id":9`)
		require.Contains(t, string(body), `"actor_platform_user_id":14`)
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "secret", server.Client())
	require.NoError(t, err)

	require.NoError(t, client.Authorize(context.Background(), 9, 14))
}

func TestClientReturnsForbiddenOnlyForAuthorizationRejection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writer.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "secret", server.Client())
	require.NoError(t, err)

	require.ErrorIs(t, client.Authorize(context.Background(), 9, 14),
		service.ErrTargetOrganizationForbidden)
}

func TestClientDoesNotMaskAuthorizationInfrastructureFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "secret", server.Client())
	require.NoError(t, err)

	err = client.Authorize(context.Background(), 9, 14)
	require.Error(t, err)
	require.NotErrorIs(t, err, service.ErrTargetOrganizationForbidden)
}

func TestClientFailsClosedOnRuntimeRejection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writer.WriteHeader(http.StatusUnprocessableEntity)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "secret", server.Client())
	require.NoError(t, err)

	_, err = client.Install(context.Background(), service.RuntimeInstallRequest{})
	require.ErrorIs(t, err, service.ErrRuntimeInstallationRejected)
}

func TestClientKeepsServerFailureRecoverable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "secret", server.Client())
	require.NoError(t, err)

	_, err = client.Install(context.Background(), service.RuntimeInstallRequest{})
	require.ErrorIs(t, err, service.ErrRuntimeInstallationUnknown)
}
