package embedtoken

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestContextRedeemsToRestrictedSession(t *testing.T) {
	service := testService(t)
	grant, err := service.IssueContext(t.Context(), ContextInput{
		SessionID:            "session-123",
		OrganizationID:       42,
		OrganizationSlug:     "acme",
		UserID:               7,
		Capabilities:         []string{"read", "write"},
		AllowedParentOrigins: []string{"https://portal.example"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, grant.RedemptionProof)
	require.WithinDuration(t, time.Now().Add(5*time.Minute), grant.ExpiresAt, 2*time.Second)

	contextClaims, err := service.InspectContext(t.Context(), grant.Token)
	require.NoError(t, err)
	require.Equal(t, "session-123", contextClaims.SessionID)

	sessionToken, _, err := service.RedeemContext(
		t.Context(),
		grant.Token,
		grant.RedemptionProof,
	)
	require.NoError(t, err)

	sessionClaims, err := service.ValidateSession(sessionToken)
	require.NoError(t, err)
	require.Equal(t, contextClaims.SessionID, sessionClaims.SessionID)
	require.Equal(t, contextClaims.OrganizationID, sessionClaims.OrganizationID)
	require.Equal(t, contextClaims.Capabilities, sessionClaims.Capabilities)
	require.Equal(t, contextClaims.AllowedParentOrigins, sessionClaims.AllowedParentOrigins)
}

func TestContextCanOnlyBeRedeemedOnce(t *testing.T) {
	service := testService(t)
	grant, err := service.IssueContext(t.Context(), ContextInput{
		SessionID:            "session-123",
		OrganizationID:       42,
		OrganizationSlug:     "acme",
		UserID:               7,
		Capabilities:         []string{"read"},
		AllowedParentOrigins: []string{"https://portal.example"},
	})
	require.NoError(t, err)

	_, _, err = service.RedeemContext(t.Context(), grant.Token, grant.RedemptionProof)
	require.NoError(t, err)
	_, _, err = service.RedeemContext(t.Context(), grant.Token, grant.RedemptionProof)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestWrongProofDoesNotConsumeContext(t *testing.T) {
	service := testService(t)
	grant, err := service.IssueContext(t.Context(), ContextInput{
		SessionID:            "session-123",
		OrganizationID:       42,
		OrganizationSlug:     "acme",
		UserID:               7,
		Capabilities:         []string{"read"},
		AllowedParentOrigins: []string{"https://portal.example"},
	})
	require.NoError(t, err)

	_, _, err = service.RedeemContext(t.Context(), grant.Token, "wrong-proof")
	require.ErrorIs(t, err, ErrInvalidToken)
	_, err = service.InspectContext(t.Context(), grant.Token)
	require.NoError(t, err)

	_, _, err = service.RedeemContext(t.Context(), grant.Token, grant.RedemptionProof)
	require.NoError(t, err)
	_, err = service.InspectContext(t.Context(), grant.Token)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func testService(t *testing.T) *Service {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return NewService("test-secret", client)
}
