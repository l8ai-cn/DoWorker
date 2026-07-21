package aider

import (
	"github.com/l8ai-cn/agentcloud/runner/internal/agentkit"
	"github.com/l8ai-cn/agentcloud/runner/internal/tokenusage"
)

func init() {
	tokenusage.RegisterParser([]string{"aider"}, &aiderParser{})
	agentkit.RegisterProcessNames("aider")
}
