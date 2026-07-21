package agent

import "github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"

func (a *Agent) ValidateIdentifiers() error {
	if err := slugkit.ValidateIdentifier("agents.slug", a.Slug); err != nil {
		return err
	}
	return slugkit.ValidateIdentifier("agents.adapter_id", a.AdapterID)
}
