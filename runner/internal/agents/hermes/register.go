package hermes

import (
	"github.com/l8ai-cn/agentcloud/runner/internal/agentkit"
	"github.com/l8ai-cn/agentcloud/runner/internal/tokenusage"
)

func init() {
	tokenusage.RegisterParserOptOut([]string{"hermes", "hermes-agent", "hermes-acp"})
	agentkit.RegisterProcessNames("hermes", "hermes-agent", "hermes-acp")
	agentkit.RegisterAgentHome(agentkit.AgentHomeSpec{
		EnvVar:      "HERMES_HOME",
		UserDirName: ".hermes",
	})
}
