package infra

import (
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"gorm.io/gorm"
)

func loadCurrentRevision(
	tx *gorm.DB,
	scope orchestrationcontrol.Scope,
	head orchestrationResourceRecord,
) (orchestrationcontrol.ResourceRevision, error) {
	var record orchestrationRevisionRecord
	err := tx.Where(
		"organization_id = ? AND resource_id = ? AND revision = ?",
		scope.OrganizationID,
		head.ID,
		head.ActiveRevision,
	).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return orchestrationcontrol.ResourceRevision{}, corruptRecord(
			"active revision",
		)
	}
	if err != nil {
		return orchestrationcontrol.ResourceRevision{}, err
	}
	return orchestrationRevisionDomain(scope, record, head)
}
