package infra

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type orchestrationResourceRecord struct {
	ID              int64           `gorm:"column:id;primaryKey"`
	OrganizationID  int64           `gorm:"column:organization_id"`
	UID             string          `gorm:"column:uid"`
	APIVersion      string          `gorm:"column:api_version"`
	Kind            string          `gorm:"column:kind"`
	Namespace       string          `gorm:"column:namespace"`
	Name            string          `gorm:"column:name"`
	DisplayName     string          `gorm:"column:display_name"`
	Labels          json.RawMessage `gorm:"column:labels;type:jsonb"`
	Status          json.RawMessage `gorm:"column:status;type:jsonb"`
	Generation      int64           `gorm:"column:generation"`
	ResourceVersion int64           `gorm:"column:resource_version"`
	ActiveRevision  int64           `gorm:"column:active_revision"`
	CreatedByID     int64           `gorm:"column:created_by_id"`
	UpdatedByID     int64           `gorm:"column:updated_by_id"`
	CreatedAt       time.Time       `gorm:"column:created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at"`
}

func (orchestrationResourceRecord) TableName() string {
	return "orchestration_resources"
}

func orchestrationResourceRecordFromDomain(
	head orchestrationcontrol.ResourceHead,
	scope orchestrationcontrol.Scope,
) (orchestrationResourceRecord, error) {
	if err := head.Validate(scope); err != nil {
		return orchestrationResourceRecord{}, err
	}
	labels, err := canonicalLabels(head.Labels)
	if err != nil {
		return orchestrationResourceRecord{}, err
	}
	return orchestrationResourceRecord{
		ID: head.ID, OrganizationID: head.OrganizationID,
		UID: head.Identity.UID, APIVersion: head.Identity.APIVersion,
		Kind: head.Identity.Kind, Namespace: head.Identity.Namespace.String(),
		Name: head.Identity.Name.String(), DisplayName: head.DisplayName,
		Labels: labels, Status: head.Status, Generation: head.Generation,
		ResourceVersion: head.ResourceVersion, ActiveRevision: head.Revision,
		CreatedByID: head.CreatedByID, UpdatedByID: head.UpdatedByID,
		CreatedAt: head.CreatedAt, UpdatedAt: head.UpdatedAt,
	}, nil
}

func (record orchestrationResourceRecord) domain(
	scope orchestrationcontrol.Scope,
) (orchestrationcontrol.ResourceHead, error) {
	labelsJSON, err := orchestrationcontrol.CanonicalJSONObject(record.Labels)
	if err != nil {
		return orchestrationcontrol.ResourceHead{}, corruptRecord("resource labels")
	}
	var labels map[string]string
	if err := json.Unmarshal(labelsJSON, &labels); err != nil {
		return orchestrationcontrol.ResourceHead{}, corruptRecord("resource labels")
	}
	status, err := orchestrationcontrol.CanonicalJSONObject(record.Status)
	if err != nil {
		return orchestrationcontrol.ResourceHead{}, corruptRecord("resource status")
	}
	head := orchestrationcontrol.ResourceHead{
		ID: record.ID, OrganizationID: record.OrganizationID,
		Identity: orchestrationcontrol.ResourceIdentity{
			ResourceTarget: orchestrationcontrol.ResourceTarget{
				TypeMeta: orchestrationresource.TypeMeta{
					APIVersion: record.APIVersion, Kind: record.Kind,
				},
				Namespace: slugkit.Slug(record.Namespace),
				Name:      slugkit.Slug(record.Name),
			},
			UID: record.UID,
		},
		DisplayName: record.DisplayName, Labels: labels, Status: status,
		Revision: record.ActiveRevision, Generation: record.Generation,
		ResourceVersion: record.ResourceVersion, CreatedByID: record.CreatedByID,
		UpdatedByID: record.UpdatedByID, CreatedAt: record.CreatedAt.UTC(),
		UpdatedAt: record.UpdatedAt.UTC(),
	}
	if err := head.Validate(scope); err != nil {
		return orchestrationcontrol.ResourceHead{}, fmt.Errorf(
			"%w: resource head",
			orchestrationcontrol.ErrCorrupt,
		)
	}
	return head, nil
}

func canonicalLabels(labels map[string]string) ([]byte, error) {
	if labels == nil {
		labels = map[string]string{}
	}
	return orchestrationcontrol.CanonicalJSONObject(labels)
}
