package infra

import (
	"context"
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
)

func (repo *aiResourceRepo) ListEffective(
	ctx context.Context,
	userID, orgID int64,
	modalities []airesource.Modality,
) ([]*airesource.ModelResource, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("AI resource effective user ID must be positive")
	}
	if orgID < 0 {
		return nil, fmt.Errorf("AI resource effective organization ID cannot be negative")
	}
	wanted := make(map[airesource.Modality]struct{}, len(modalities))
	for _, modality := range modalities {
		if !modality.Valid() {
			return nil, fmt.Errorf("invalid AI resource modality %q", modality)
		}
		wanted[modality] = struct{}{}
	}

	query := repo.db.WithContext(ctx).Model(&modelResourceRow{}).
		Select("model_resources.*").
		Joins("JOIN provider_connections ON provider_connections.id = model_resources.provider_connection_id").
		Where("model_resources.is_enabled = ? AND provider_connections.is_enabled = ?", true, true).
		Where("model_resources.status = ? AND provider_connections.status = ?", airesource.ConnectionStatusValid, airesource.ConnectionStatusValid)
	if orgID > 0 {
		query = query.Where(
			repo.db.Where("provider_connections.owner_scope = ? AND provider_connections.owner_id = ?", airesource.OwnerScopeUser, userID).
				Or("provider_connections.owner_scope = ? AND provider_connections.owner_id = ?", airesource.OwnerScopeOrg, orgID),
		)
	} else {
		query = query.Where("provider_connections.owner_scope = ? AND provider_connections.owner_id = ?", airesource.OwnerScopeUser, userID)
	}

	var rows []modelResourceRow
	if err := query.Order("CASE WHEN provider_connections.owner_scope = 'user' THEN 0 ELSE 1 END, model_resources.identifier, model_resources.id").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(wanted) > 0 {
		rows = filterResourceRowsByModality(rows, wanted)
	}
	return repo.resourcesWithDefaults(ctx, rows, wanted, true)
}

func (repo *aiResourceRepo) ListResourcesByOwner(ctx context.Context, scope airesource.OwnerScope, ownerID int64) ([]*airesource.ModelResource, error) {
	var rows []modelResourceRow
	err := repo.db.WithContext(ctx).Model(&modelResourceRow{}).
		Select("model_resources.*").
		Joins("JOIN provider_connections ON provider_connections.id = model_resources.provider_connection_id").
		Where("provider_connections.owner_scope = ? AND provider_connections.owner_id = ?", scope, ownerID).
		Order("provider_connections.identifier, model_resources.identifier, model_resources.id").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return repo.resourcesWithDefaults(ctx, rows, nil, false)
}

func filterResourceRowsByModality(rows []modelResourceRow, wanted map[airesource.Modality]struct{}) []modelResourceRow {
	filtered := make([]modelResourceRow, 0, len(rows))
	for _, row := range rows {
		for _, modality := range row.Modalities {
			if _, matches := wanted[airesource.Modality(modality)]; matches {
				filtered = append(filtered, row)
				break
			}
		}
	}
	return filtered
}
