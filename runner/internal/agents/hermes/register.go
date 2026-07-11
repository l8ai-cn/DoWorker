package harn

import (
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/tokenusage"
)

func init() {
	tokenusage.RegisterParserOptOut([]string{"harn"})
	agentkit.RegisterProcessNames("harn")
	agentkit.RegisterAgentHome(agentkit.AgentHomeSpec{
		EnvVar:      "HARN_HOME",
		UserDirName: ".harn",
	})
}
