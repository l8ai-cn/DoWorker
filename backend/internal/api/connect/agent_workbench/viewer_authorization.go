package agentworkbenchconnect

import (
	"context"
	"fmt"
	"time"

	sessiondomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	"github.com/l8ai-cn/agentcloud/backend/pkg/embedtoken"
	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
)

type viewerAuthorization struct {
	capabilities *agentworkbenchv2.SupportCapabilities
	embed        *embedtoken.Claims
	issued       string
	expires      *string
	subject      string
}

type viewerAuthorizationContextKey struct{}

func withViewerAuthorization(
	ctx context.Context,
	authorization *viewerAuthorization,
) context.Context {
	return context.WithValue(ctx, viewerAuthorizationContextKey{}, authorization)
}

func viewerAuthorizationFrom(ctx context.Context) *viewerAuthorization {
	value, _ := ctx.Value(viewerAuthorizationContextKey{}).(*viewerAuthorization)
	return value
}

func viewerAuthorizationFor(
	ctx context.Context,
	session *sessiondomain.Session,
) (*viewerAuthorization, error) {
	issued := session.CreatedAt
	if issued.IsZero() {
		issued = time.Unix(0, 0)
	}
	authorization := &viewerAuthorization{
		issued: issued.UTC().Format(time.RFC3339Nano),
	}
	if claims := embedAccess(ctx); claims != nil {
		authorization.embed = claims
		authorization.subject = fmt.Sprintf("embed:user:%d", claims.UserID)
		if claims.ID != "" {
			authorization.subject = "embed:" + claims.ID
		}
		if claims.ExpiresAt != nil {
			value := claims.ExpiresAt.Time.UTC().Format(time.RFC3339Nano)
			authorization.expires = &value
		}
		return authorization, nil
	}
	tenant := middleware.GetTenant(ctx)
	if tenant == nil || tenant.UserID <= 0 {
		return nil, unauthenticated("agent workbench viewer is unavailable")
	}
	authorization.subject = fmt.Sprintf("user:%d", tenant.UserID)
	return authorization, nil
}

func (authorization *viewerAuthorization) decorateSnapshot(
	snapshot *agentworkbenchv2.SessionSnapshot,
) error {
	authorization.capabilities = snapshot.GetCapabilities()
	snapshot.Grants = []*agentworkbenchv2.AuthorizationGrant{{
		GrantId:   "backend:" + authorization.subject + ":" + snapshot.GetSessionId(),
		Issuer:    "agentcloud.backend",
		Subject:   authorization.subject,
		SessionId: snapshot.GetSessionId(),
		Actions:   authorization.sessionActions(snapshot.GetCapabilities()),
		IssuedAt:  authorization.issued,
		ExpiresAt: authorization.expires,
	}}
	for _, artifact := range snapshot.Artifacts {
		authorization.decorateArtifact(artifact)
	}
	digest, err := snapshotDigest(snapshot)
	if err != nil {
		return err
	}
	snapshot.Digest = &digest
	return nil
}

func (authorization *viewerAuthorization) sessionActions(
	capabilities *agentworkbenchv2.SupportCapabilities,
) []string {
	if capabilities == nil {
		return nil
	}
	actions := make([]string, 0)
	for _, descriptor := range capabilities.CommandSchemas {
		for _, action := range descriptor.GetActions() {
			if authorization.allowsSessionAction(action) {
				actions = appendUnique(actions, action)
			}
		}
	}
	for _, action := range capabilities.TerminalOperations {
		if authorization.allowsTerminal() {
			actions = appendUnique(actions, action)
		}
	}
	return actions
}

func (authorization *viewerAuthorization) allowsSessionAction(action string) bool {
	if authorization.embed == nil {
		return true
	}
	switch action {
	case "session.send":
		return hasEmbedCapability(authorization.embed, "write")
	case "session.permission.resolve":
		return hasEmbedCapability(authorization.embed, "approve")
	case "session.interrupt", "session.configure":
		return hasEmbedCapability(authorization.embed, "control")
	default:
		return false
	}
}

func (authorization *viewerAuthorization) allowsTerminal() bool {
	return authorization.embed == nil ||
		(hasEmbedCapability(authorization.embed, "terminal") &&
			hasEmbedCapability(authorization.embed, "control"))
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
