package agentworkbenchconnect

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/pkg/embedtoken"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
)

type embedAccessContextKey struct{}

func withEmbedAccess(
	ctx context.Context,
	claims *embedtoken.Claims,
) context.Context {
	return context.WithValue(ctx, embedAccessContextKey{}, claims)
}

func embedAccess(ctx context.Context) *embedtoken.Claims {
	claims, _ := ctx.Value(embedAccessContextKey{}).(*embedtoken.Claims)
	return claims
}

func requireEmbedCommandCapability(
	claims *embedtoken.Claims,
	command *agentworkbenchv2.CommandEnvelope,
) error {
	if claims == nil {
		return nil
	}
	required := ""
	switch command.GetCommand().(type) {
	case *agentworkbenchv2.CommandEnvelope_SendPrompt:
		required = "write"
	case *agentworkbenchv2.CommandEnvelope_ResolvePermission:
		required = "approve"
	case *agentworkbenchv2.CommandEnvelope_Interrupt,
		*agentworkbenchv2.CommandEnvelope_ChangeConfiguration,
		*agentworkbenchv2.CommandEnvelope_TerminalOperation:
		required = "control"
	case *agentworkbenchv2.CommandEnvelope_ArtifactAction:
		required = "write"
	default:
		return permissionDenied("embedded command is not permitted")
	}
	if hasEmbedCapability(claims, required) {
		if _, ok := command.GetCommand().(*agentworkbenchv2.CommandEnvelope_TerminalOperation); ok &&
			!hasEmbedCapability(claims, "terminal") {
			return permissionDenied("embedded terminal capability is required")
		}
		return nil
	}
	return permissionDenied("embedded command capability is required")
}

func hasEmbedCapability(claims *embedtoken.Claims, capability string) bool {
	for _, value := range claims.Capabilities {
		if value == capability {
			return true
		}
	}
	return false
}
