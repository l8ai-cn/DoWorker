package agentworkbenchconnect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/pkg/embedtoken"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

type fakeEmbedTokenValidator struct {
	claims *embedtoken.Claims
	err    error
}

func (f fakeEmbedTokenValidator) ValidateSession(string) (*embedtoken.Claims, error) {
	return f.claims, f.err
}

func TestAgentWorkbenchAuthAcceptsScopedEmbedSessionToken(t *testing.T) {
	repository := &fakeRepository{state: statePointer(testSnapshotState(t, 1, 1))}
	claims := testEmbedClaims([]string{"read"})
	mux := http.NewServeMux()
	Mount(
		mux,
		testServer(repository, nil),
		connect.WithInterceptors(NewAuthInterceptor(
			workbenchTestAccessTokenManager(t),
			testAudience,
			fakeEmbedTokenValidator{claims: claims},
		)),
	)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client := connect.NewClient[
		agentworkbenchv2.GetSessionSnapshotRequest,
		agentworkbenchv2.SessionSnapshot,
	](server.Client(), server.URL+GetSessionSnapshotProcedure)
	request := connect.NewRequest(&agentworkbenchv2.GetSessionSnapshotRequest{
		OrgSlug: testOrgSlug, SessionId: testSessionID,
	})
	request.Header().Set("Authorization", "Bearer "+signedEmbedMarker(t))

	response, err := client.CallUnary(context.Background(), request)

	require.NoError(t, err)
	require.Equal(t, testSessionID, response.Msg.GetSessionId())
}

func TestAgentWorkbenchAuthRejectsEmbedTokenWithoutValidator(t *testing.T) {
	interceptor := NewAuthInterceptor(
		workbenchTestAccessTokenManager(t),
		testAudience,
		nil,
	)
	handler := interceptor.WrapUnary(func(
		context.Context,
		connect.AnyRequest,
	) (connect.AnyResponse, error) {
		return nil, nil
	})
	request := connect.NewRequest(&agentworkbenchv2.GetSessionSnapshotRequest{})
	request.Header().Set("Authorization", "Bearer "+signedEmbedMarker(t))

	_, err := handler(context.Background(), request)

	requireConnectCode(t, err, connect.CodeUnauthenticated)
}

func testEmbedClaims(capabilities []string) *embedtoken.Claims {
	return &embedtoken.Claims{
		SessionID:            testSessionID,
		OrganizationID:       7,
		OrganizationSlug:     testOrgSlug,
		UserID:               testUserID,
		Capabilities:         capabilities,
		AllowedParentOrigins: []string{"https://portal.example"},
		TokenUse:             embedtoken.SessionTokenUse,
	}
}

func signedEmbedMarker(t *testing.T) string {
	t.Helper()
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.JWTClaims{
		UserID:   testUserID,
		TokenUse: embedtoken.SessionTokenUse,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	})
	signed, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)
	return signed
}
