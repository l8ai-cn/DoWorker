package interceptors

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	authpkg "github.com/l8ai-cn/agentcloud/backend/pkg/auth"
)

func NewAuthInterceptor(
	manager *authpkg.AccessTokenManager,
	audience string,
) connect.Interceptor {
	return &authInterceptor{manager: manager, audience: audience}
}

type authInterceptor struct {
	manager  *authpkg.AccessTokenManager
	audience string
}

func (a *authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			return next(ctx, req)
		}
		ctx, err := a.injectTenant(ctx, req.Header())
		if err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

func (a *authInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (a *authInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, err := a.injectTenant(ctx, conn.RequestHeader())
		if err != nil {
			return err
		}
		return next(ctx, conn)
	}
}

func (a *authInterceptor) injectTenant(ctx context.Context, header http.Header) (context.Context, error) {
	claims, err := parseBearerToken(header.Get("Authorization"), a.manager, a.audience)
	if err != nil {
		return ctx, err
	}
	ctx = middleware.SetTenant(ctx, &middleware.TenantContext{UserID: claims.UserID})
	return withClaims(ctx, claims), nil
}

func parseBearerToken(
	header string,
	manager *authpkg.AccessTokenManager,
	audience string,
) (*middleware.JWTClaims, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return nil, unauthenticated("authorization bearer token is required")
	}
	if manager == nil {
		return nil, unauthenticated("access token verifier is not configured")
	}
	claims, err := manager.ValidateToken(parts[1], audience)
	if err != nil {
		return nil, unauthenticated("invalid or expired token")
	}
	return claims, nil
}

func unauthenticated(message string) error {
	return connect.NewError(connect.CodeUnauthenticated, errors.New(message))
}
