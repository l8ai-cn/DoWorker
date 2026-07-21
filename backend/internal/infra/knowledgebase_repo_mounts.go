package infra

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/knowledgebase"
)

func (r *knowledgeBaseRepo) ReplaceAgentMounts(ctx context.Context, orgID, kbID int64, mounts []*knowledgebase.AgentMount) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.
		Where("organization_id = ? AND knowledge_base_id = ?", orgID, kbID).
		Delete(&knowledgebase.AgentMount{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, m := range mounts {
		m.OrganizationID = orgID
		m.KnowledgeBaseID = kbID
		if err := tx.Create(m).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *knowledgeBaseRepo) ListAgentMounts(ctx context.Context, orgID, kbID int64) ([]*knowledgebase.AgentMount, error) {
	var mounts []*knowledgebase.AgentMount
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND knowledge_base_id = ?", orgID, kbID).
		Order("agent_slug").
		Find(&mounts).Error
	return mounts, err
}

func (r *knowledgeBaseRepo) ListMountsForAgent(ctx context.Context, orgID int64, agentSlug string) ([]*knowledgebase.AgentMount, error) {
	var mounts []*knowledgebase.AgentMount
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND agent_slug = ?", orgID, agentSlug).
		Find(&mounts).Error
	return mounts, err
}
