package infra

import (
	"context"
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (repo *aiResourceRepo) SetDefault(ctx context.Context, resourceID int64, modality airesource.Modality) error {
	if !modality.Valid() {
		return fmt.Errorf("invalid AI resource modality %q", modality)
	}
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var resource modelResourceRow
		if err := tx.First(&resource, resourceID).Error; err != nil {
			return err
		}
		if !resourceSupports(resource, modality) {
			return fmt.Errorf("model resource %d does not support modality %q", resourceID, modality)
		}
		var connection providerConnectionRow
		if err := tx.First(&connection, resource.ProviderConnectionID).Error; err != nil {
			return err
		}
		row := modelResourceDefaultRow{
			OwnerScope: connection.OwnerScope, OwnerID: connection.OwnerID,
			Modality: modality, ModelResourceID: resourceID,
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "owner_scope"}, {Name: "owner_id"}, {Name: "modality"}},
			DoUpdates: clause.AssignmentColumns([]string{"model_resource_id", "updated_at"}),
		}).Create(&row).Error
	})
}

func resourceSupports(resource modelResourceRow, modality airesource.Modality) bool {
	for _, supported := range resource.Modalities {
		if airesource.Modality(supported) == modality {
			return true
		}
	}
	return false
}

func (repo *aiResourceRepo) loadDefaultRows(ctx context.Context, resourceIDs []int64) (map[int64][]modelResourceDefaultRow, error) {
	defaults := make(map[int64][]modelResourceDefaultRow, len(resourceIDs))
	if len(resourceIDs) == 0 {
		return defaults, nil
	}
	var rows []modelResourceDefaultRow
	if err := repo.db.WithContext(ctx).Where("model_resource_id IN ?", resourceIDs).
		Order("modality, model_resource_id").Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		defaults[row.ModelResourceID] = append(defaults[row.ModelResourceID], row)
	}
	return defaults, nil
}

func defaultModalities(rows []modelResourceDefaultRow, wanted map[airesource.Modality]struct{}) []airesource.Modality {
	modalities := make([]airesource.Modality, 0, len(rows))
	for _, row := range rows {
		if len(wanted) > 0 {
			if _, included := wanted[row.Modality]; !included {
				continue
			}
		}
		modalities = append(modalities, row.Modality)
	}
	return modalities
}

func effectiveDefaultModalities(
	rows []modelResourceDefaultRow,
	wanted map[airesource.Modality]struct{},
	personal map[airesource.Modality]struct{},
) []airesource.Modality {
	filtered := rows[:0]
	for _, row := range rows {
		if row.OwnerScope == airesource.OwnerScopeOrg {
			if _, overridden := personal[row.Modality]; overridden {
				continue
			}
		}
		filtered = append(filtered, row)
	}
	return defaultModalities(filtered, wanted)
}
