package agentconnect

import (
	"github.com/l8ai-cn/agentcloud/agentfile/capability"
	agentv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent/v1"
)

func enrichCapabilities(proto *agentv1.Agent, agentfileSource *string) {
	if proto == nil {
		return
	}
	src := ""
	if agentfileSource != nil {
		src = *agentfileSource
	}
	if src == "" {
		src = proto.GetAgentfileSource()
	}
	if src == "" {
		return
	}
	caps := capability.ScanDeclarations(src)
	if len(caps) == 0 {
		return
	}
	proto.Capabilities = caps
}
