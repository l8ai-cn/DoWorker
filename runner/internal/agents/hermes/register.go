package hermes

import (
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
)

func init() {
	tokenusage.RegisterParserOptOut([]string{"hermes", "hermes-agent", "hermes-acp"})
	agentkit.RegisterProcessNames("hermes", "hermes-agent", "hermes-acp")
	agentkit.RegisterAgentHome(agentkit.AgentHomeSpec{
		EnvVar:      "HERMES_HOME",
		UserDirName: ".hermes",
	})
}
