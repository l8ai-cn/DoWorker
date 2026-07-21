package runner

import (
	"github.com/l8ai-cn/agentcloud/runner/internal/terminal"
	"github.com/l8ai-cn/agentcloud/runner/internal/terminal/aggregator"
	"github.com/l8ai-cn/agentcloud/runner/internal/terminal/vt"
)

// PTYComponents holds PTY-specific infrastructure shared between PTYPodIO and PTYPodRelay.
// Both abstractions receive a pointer to the same instance, eliminating field duplication.
type PTYComponents struct {
	Terminal        *terminal.Terminal
	VirtualTerminal *vt.VirtualTerminal
	Aggregator      *aggregator.SmartAggregator
	PTYLogger       *aggregator.PTYLogger
}
