package knowledgebase

import "context"

type ListFilter struct {
	OrganizationID int64
	SourceType     string
}

type Repository interface {
	Create(ctx context.Context, kb *KnowledgeBase) error
	Get(ctx context.Context, orgID, id int64) (*KnowledgeBase, error)
	GetBySlug(ctx context.Context, orgID int64, slug string) (*KnowledgeBase, error)
	List(ctx context.Context, filter *ListFilter) ([]*KnowledgeBase, error)
	// ListExternal returns every KB with a non-git source type across all
	// orgs — the sync worker's work queue.
	ListExternal(ctx context.Context) ([]*KnowledgeBase, error)
	ListBySlugs(ctx context.Context, orgID int64, slugs []string) ([]*KnowledgeBase, error)
	Update(ctx context.Context, orgID, id int64, updates map[string]any) error
	Delete(ctx context.Context, orgID, id int64) error
	SlugExists(ctx context.Context, orgID int64, slug string) (bool, error)

	ReplaceAgentMounts(ctx context.Context, orgID, kbID int64, mounts []*AgentMount) error
	ListAgentMounts(ctx context.Context, orgID, kbID int64) ([]*AgentMount, error)
	ListMountsForAgent(ctx context.Context, orgID int64, agentSlug string) ([]*AgentMount, error)
}
