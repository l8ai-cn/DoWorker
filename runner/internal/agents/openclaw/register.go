package openclaw

import (
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
)

func init() {
	tokenusage.RegisterParserOptOut([]string{"openclaw"})
	agentkit.RegisterProcessNames("openclaw")
	agentkit.RegisterAgentHome(agentkit.AgentHomeSpec{
		EnvVar:      "OPENCLAW_HOME",
		UserDirName: ".openclaw",
	})
}
