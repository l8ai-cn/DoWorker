package knowledgebase

import "github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"

func (kb *KnowledgeBase) ValidateIdentifiers() error {
	return slugkit.ValidateIdentifier("knowledge_bases.slug", kb.Slug)
}

func (m *AgentMount) ValidateIdentifiers() error {
	return slugkit.ValidateIdentifier("knowledge_base_agent_mounts.agent_slug", m.AgentSlug)
}
