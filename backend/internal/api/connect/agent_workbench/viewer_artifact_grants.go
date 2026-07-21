package agentworkbenchconnect

import (
	"context"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

func (authorization *viewerAuthorization) decorateArtifact(
	artifact *agentworkbenchv2.ArtifactDescriptor,
) {
	if artifact == nil {
		return
	}
	candidates := []string{"artifact.download"}
	if authorization.embed == nil || hasEmbedCapability(authorization.embed, "write") {
		switch artifact.GetManifest().GetManifest().(type) {
		case *agentworkbenchv2.ArtifactManifest_ImageEdit:
			candidates = append(candidates, "image.edit")
		case *agentworkbenchv2.ArtifactManifest_Presentation:
			candidates = append(candidates,
				"presentation.regenerate_slide",
				"presentation.replace_slide",
				"presentation.reorder_slide",
				"presentation.export",
				"presentation.select_version",
			)
		}
	}
	actions := make([]string, 0, len(candidates))
	for _, action := range candidates {
		if supportsArtifactAction(authorization.capabilities, action) {
			actions = append(actions, action)
		}
	}
	representations := make([]string, 0, len(artifact.Representations))
	for _, representation := range artifact.Representations {
		if representation.GetRepresentationId() != "" {
			representations = append(representations, representation.GetRepresentationId())
		}
	}
	artifact.Grants = []*agentworkbenchv2.ArtifactGrant{{
		GrantId:           "backend:" + authorization.subject + ":" + artifact.GetArtifactId(),
		Issuer:            optionalString("agentcloud.backend"),
		Subject:           optionalString(authorization.subject),
		RepresentationIds: representations,
		Actions:           actions,
		MinimumRevision:   optionalUint64(artifact.GetRevision()),
		IssuedAt:          optionalString(authorization.issued),
		ExpiresAt:         authorization.expires,
	}}
}

func (authorization *viewerAuthorization) decorateDelta(
	delta *agentworkbenchv2.SessionDeltaBatch,
) (*agentworkbenchv2.SessionDeltaBatch, error) {
	decorated := proto.Clone(delta).(*agentworkbenchv2.SessionDeltaBatch)
	for _, event := range decorated.Events {
		if changed := event.GetCapabilitiesChanged(); changed != nil {
			authorization.capabilities = changed.GetCapabilities()
		}
		if changed := event.GetArtifactChanged(); changed != nil {
			authorization.decorateArtifact(changed.Artifact)
		}
	}
	digest, err := deltaDigest(decorated)
	if err != nil {
		return nil, err
	}
	decorated.Digest = digest
	return decorated, nil
}

func authorizedDelta(
	ctx context.Context,
	delta *agentworkbenchv2.SessionDeltaBatch,
) (*agentworkbenchv2.SessionDeltaBatch, error) {
	authorization := viewerAuthorizationFrom(ctx)
	if authorization == nil {
		return nil, unauthenticated("agent workbench viewer is unavailable")
	}
	return authorization.decorateDelta(delta)
}

func optionalString(value string) *string {
	return &value
}

func optionalUint64(value uint64) *uint64 {
	return &value
}
