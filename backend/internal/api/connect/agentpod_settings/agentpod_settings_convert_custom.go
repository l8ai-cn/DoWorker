package agentpodsettingsconnect

import (
	poddom "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	podv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/pod/v1"
)

// toProtoSettings preserves the REST contract: nil input → empty (not nil)
// AgentPodSettings so first-read callers see a default-shape response. The
// codegen `ToProtoAgentPodSettings` returns nil on nil; this thin wrapper
// reproduces the empty-shape branch.
func toProtoSettings(s *poddom.UserAgentPodSettings) *podv1.AgentPodSettings {
	if s == nil {
		return &podv1.AgentPodSettings{}
	}
	return ToProtoAgentPodSettings(s)
}
