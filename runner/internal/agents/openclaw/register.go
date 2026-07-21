package openclaw

import (
	"github.com/l8ai-cn/agentcloud/runner/internal/agentkit"
	"github.com/l8ai-cn/agentcloud/runner/internal/tokenusage"
)

func init() {
	tokenusage.RegisterParserOptOut([]string{"openclaw"})
	agentkit.RegisterProcessNames("openclaw")
	agentkit.RegisterAgentHome(agentkit.AgentHomeSpec{
		EnvVar:      "OPENCLAW_HOME",
		UserDirName: ".openclaw",
		MergeConfig: MergeConfig,
	})
}
