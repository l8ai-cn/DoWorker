package coordinator

import "github.com/anthropics/agentsmesh/backend/pkg/slugkit"

func (p *Project) ValidateIdentifiers() error {
	if err := slugkit.ValidateIdentifier("coordinator_projects.slug", p.Slug); err != nil {
		return err
	}
	return slugkit.ValidateIdentifier("coordinator_projects.agent_slug", p.AgentSlug)
}
