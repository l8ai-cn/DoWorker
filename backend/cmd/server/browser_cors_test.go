package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBrowserCORSHandlesConnectPreflightBeforeRouting(t *testing.T) {
	connectCalled := false
	restCalled := false
	handler := withBrowserCORS(
		[]string{"http://localhost:29957"},
		routeConnectOrREST(
			http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				connectCalled = true
			}),
			http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				restCalled = true
			}),
		),
	)
	request := httptest.NewRequest(
		http.MethodOptions,
		"/proto.agent_workbench.v2.AgentWorkbenchService/StreamSessionDeltas",
		nil,
	)
	request.Header.Set("Origin", "http://localhost:29957")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set(
		"Access-Control-Request-Headers",
		"authorization,connect-protocol-version,connect-timeout-ms,content-type",
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
	require.False(t, connectCalled)
	require.False(t, restCalled)
	require.Equal(
		t,
		"http://localhost:29957",
		response.Header().Get("Access-Control-Allow-Origin"),
	)
	allowedHeaders := strings.ToLower(response.Header().Get("Access-Control-Allow-Headers"))
	require.Contains(t, allowedHeaders, "connect-protocol-version")
	require.Contains(t, allowedHeaders, "connect-timeout-ms")
}

func TestBrowserCORSRejectsUnlistedOrigin(t *testing.T) {
	called := false
	handler := withBrowserCORS(
		[]string{"https://app.example.com"},
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			called = true
		}),
	)
	request := httptest.NewRequest(http.MethodOptions, "/proto.example.Service/Stream", nil)
	request.Header.Set("Origin", "https://untrusted.example.com")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
	require.False(t, called)
	require.Empty(t, response.Header().Get("Access-Control-Allow-Origin"))
}
