package airesourceconnect

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rejectingInterceptor struct {
	mu    sync.Mutex
	calls map[string]int
}

func (i *rejectingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(_ context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		i.mu.Lock()
		i.calls[request.Spec().Procedure]++
		i.mu.Unlock()
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
}

func (*rejectingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (*rejectingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

func TestMountAppliesAuthenticationOptionsToEveryRPC(t *testing.T) {
	procedures := []string{
		GetCatalogProcedure,
		ListPersonalConnectionsProcedure,
		ListOrganizationConnectionsProcedure,
		ListPersonalEffectiveResourcesProcedure,
		ListOrganizationEffectiveResourcesProcedure,
		CreatePersonalConnectionProcedure,
		CreateOrganizationConnectionProcedure,
		UpdateConnectionProcedure,
		RotateConnectionCredentialsProcedure,
		SetConnectionEnabledProcedure,
		ValidateConnectionProcedure,
		DeleteConnectionProcedure,
		CreateResourceProcedure,
		UpdateResourceProcedure,
		SetResourceEnabledProcedure,
		DeleteResourceProcedure,
		SetDefaultProcedure,
	}
	interceptor := &rejectingInterceptor{calls: map[string]int{}}
	mux := http.NewServeMux()
	Mount(mux, NewServer(&fakeService{}, fakeOrgService{}), connect.WithInterceptors(interceptor))
	server := httptest.NewServer(mux)
	defer server.Close()

	for _, procedure := range procedures {
		request, err := http.NewRequest(http.MethodPost, server.URL+procedure, bytes.NewBufferString("{}"))
		require.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")
		response, err := server.Client().Do(request)
		require.NoError(t, err)
		require.NoError(t, response.Body.Close())
	}
	for _, procedure := range procedures {
		assert.Equal(t, 1, interceptor.calls[procedure], procedure)
	}
}
