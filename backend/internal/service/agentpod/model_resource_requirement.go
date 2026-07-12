package agentpod

import agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"

func AgentRequiresModelResource(agent *agentDomain.Agent) bool {
	if agent == nil {
		return false
	}
	_, required := modelResourceRequirements(agent.Slug, agent)
	return required
}
