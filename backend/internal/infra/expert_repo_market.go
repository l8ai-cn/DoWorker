package infra

import (
	"context"
	"errors"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"gorm.io/gorm"
)

func (r *ExpertRepository) GetByMarketApplication(
	ctx context.Context,
	orgID, applicationID int64,
) (*expertdom.Expert, error) {
	var row expertdom.Expert
	err := r.db.WithContext(ctx).
		Where(
			"organization_id = ? AND source_market_application_id = ?",
			orgID,
			applicationID,
		).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, expertdom.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ExpertRepository) UpdateMarketRelease(
	ctx context.Context,
	orgID, expertID, applicationID int64,
	update expertdom.MarketReleaseUpdate,
) error {
	result := r.db.WithContext(ctx).
		Model(&expertdom.Expert{}).
		Where(
			"organization_id = ? AND id = ? AND source_market_application_id = ? AND revision = ?",
			orgID,
			expertID,
			applicationID,
			update.ExpectedRevision,
		).
		Updates(marketReleaseUpdates(update))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	return r.marketUpdateMiss(ctx, orgID, expertID, applicationID)
}

func marketReleaseUpdates(update expertdom.MarketReleaseUpdate) map[string]any {
	return map[string]any{
		"name":                     update.Name,
		"description":              update.Description,
		"agent_slug":               update.AgentSlug,
		"prompt":                   update.Prompt,
		"interaction_mode":         update.InteractionMode,
		"automation_level":         update.AutomationLevel,
		"perpetual":                update.Perpetual,
		"used_env_bundles":         update.UsedEnvBundles,
		"skill_slugs":              update.SkillSlugs,
		"knowledge_mounts":         update.KnowledgeMounts,
		"config_overrides":         update.ConfigOverrides,
		"agentfile_layer":          update.AgentfileLayer,
		"metadata":                 update.Metadata,
		"worker_spec_snapshot_id":  update.WorkerSpecSnapshotID,
		"source_market_release_id": update.SourceMarketReleaseID,
		"revision":                 gorm.Expr("revision + 1"),
		"updated_at":               gorm.Expr("CURRENT_TIMESTAMP"),
	}
}

func (r *ExpertRepository) marketUpdateMiss(
	ctx context.Context,
	orgID, expertID, applicationID int64,
) error {
	var count int64
	err := r.db.WithContext(ctx).Model(&expertdom.Expert{}).
		Where(
			"organization_id = ? AND id = ? AND source_market_application_id = ?",
			orgID,
			expertID,
			applicationID,
		).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return expertdom.ErrConflict
	}
	return expertdom.ErrNotFound
}
