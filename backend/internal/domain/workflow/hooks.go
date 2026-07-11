package workflow

import "github.com/anthropics/agentsmesh/backend/pkg/slugkit"

func (l *Workflow) ValidateIdentifiers() error {
	if err := slugkit.ValidateIdentifier("workflows.slug", l.Slug); err != nil {
		return err
	}
	return slugkit.ValidateIdentifier("workflows.agent_slug", l.AgentSlug)
}
