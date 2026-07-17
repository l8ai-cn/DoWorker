package agentworkbenchconnect

import agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"

func requireArtifactAuthorization(
	snapshot *agentworkbenchv2.SessionSnapshot,
	command *agentworkbenchv2.ArtifactActionCommand,
) error {
	if command == nil {
		return permissionDenied("artifact action is not authorized")
	}
	for _, artifact := range snapshot.Artifacts {
		if artifact.GetArtifactId() != command.GetArtifactId() ||
			artifact.GetRevision() != command.GetBaseRevision() {
			continue
		}
		if !supportsArtifactAction(snapshot.GetCapabilities(), command.GetActionType()) {
			return permissionDenied("artifact action is not supported")
		}
		for _, grant := range artifact.GetGrants() {
			if grantAppliesToArtifact(grant, artifact.GetRevision(), command.GetActionType()) {
				return nil
			}
		}
		return permissionDenied("artifact action grant is required")
	}
	return permissionDenied("artifact action grant is required")
}

func supportsArtifactAction(
	capabilities *agentworkbenchv2.SupportCapabilities,
	action string,
) bool {
	if capabilities == nil || action == "" {
		return false
	}
	for _, supported := range capabilities.GetArtifactOperations() {
		if supported == action {
			return true
		}
	}
	return false
}

func grantAppliesToArtifact(
	grant *agentworkbenchv2.ArtifactGrant,
	revision uint64,
	action string,
) bool {
	if grant == nil ||
		(grant.MinimumRevision != nil && revision < grant.GetMinimumRevision()) ||
		(grant.MaximumRevision != nil && revision > grant.GetMaximumRevision()) {
		return false
	}
	for _, allowed := range grant.GetActions() {
		if allowed == action {
			return true
		}
	}
	return false
}
