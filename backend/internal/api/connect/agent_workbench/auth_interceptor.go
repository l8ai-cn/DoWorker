package agentworkbenchconnect

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	authpkg "github.com/l8ai-cn/agentcloud/backend/pkg/auth"
	"github.com/l8ai-cn/agentcloud/backend/pkg/embedtoken"
)

type EmbedTokenValidator interface {
	ValidateSession(string) (*embedtoken.Claims, error)
}

type authInterceptor struct {
	embedTokens EmbedTokenValidator
	manager     *authpkg.AccessTokenManager
	audience    string
}

func NewAuthInterceptor(
	manager *authpkg.AccessTokenManager,
	audience string,
	embedTokens EmbedTokenValidator,
) connect.Interceptor {
	return &authInterceptor{
		manager: manager, audience: audience, embedTokens: embedTokens,
	}
}

func (auth *authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(
		ctx context.Context,
		request connect.AnyRequest,
	) (connect.AnyResponse, error) {
		if request.Spec().IsClient {
			return next(ctx, request)
		}
		scoped, err := auth.authenticate(ctx, request.Header())
		if err != nil {
			return nil, err
		}
		return next(scoped, request)
	}
}

func (auth *authInterceptor) WrapStreamingClient(
	next connect.StreamingClientFunc,
) connect.StreamingClientFunc {
	return next
}

func (auth *authInterceptor) WrapStreamingHandler(
	next connect.StreamingHandlerFunc,
) connect.StreamingHandlerFunc {
	return func(ctx context.Context, connection connect.StreamingHandlerConn) error {
		scoped, err := auth.authenticate(ctx, connection.RequestHeader())
		if err != nil {
			return err
		}
		return next(scoped, connection)
	}
}

func (auth *authInterceptor) authenticate(
	ctx context.Context,
	header http.Header,
) (context.Context, error) {
	tokenValue, err := bearerValue(header.Get("Authorization"))
	if err != nil {
		return ctx, err
	}
	if auth.manager != nil {
		if claims, validateErr := auth.manager.ValidateToken(
			tokenValue,
			auth.audience,
		); validateErr == nil && claims.UserID > 0 {
			return middleware.SetTenant(
				ctx,
				&middleware.TenantContext{UserID: claims.UserID},
			), nil
		}
	}
	if auth.embedTokens == nil {
		return ctx, unauthenticated("invalid or expired token")
	}
	embedClaims, err := auth.embedTokens.ValidateSession(tokenValue)
	if err != nil || embedClaims.TokenUse != embedtoken.SessionTokenUse {
		return ctx, unauthenticated("invalid or expired token")
	}
	tenant := &middleware.TenantContext{
		OrganizationID:   embedClaims.OrganizationID,
		OrganizationSlug: embedClaims.OrganizationSlug,
		UserID:           embedClaims.UserID,
		UserRole:         "embed",
	}
	if tenant.UserID <= 0 {
		return ctx, unauthenticated("invalid or expired token")
	}
	ctx = middleware.SetTenant(ctx, tenant)
	return withEmbedAccess(ctx, embedClaims), nil
}

func bearerValue(value string) (string, error) {
	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", unauthenticated("authorization bearer token is required")
	}
	return parts[1], nil
}
