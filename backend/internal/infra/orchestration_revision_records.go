package infra

import (
	"context"
	"encoding/json"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
)

type orchestrationRevisionRecord struct {
	ID                   int64           `gorm:"column:id;primaryKey"`
	OrganizationID       int64           `gorm:"column:organization_id"`
	ResourceID           int64           `gorm:"column:resource_id"`
	Revision             int64           `gorm:"column:revision"`
	Generation           int64           `gorm:"column:generation"`
	ResourceVersion      int64           `gorm:"column:resource_version"`
	CanonicalManifest    json.RawMessage `gorm:"column:canonical_manifest;type:jsonb"`
	CanonicalSpec        json.RawMessage `gorm:"column:canonical_spec;type:jsonb"`
	ResolvedReferences   json.RawMessage `gorm:"column:resolved_refs;type:jsonb"`
	Digest               string          `gorm:"column:digest"`
	WorkerSpecSnapshotID *int64          `gorm:"column:worker_spec_snapshot_id"`
	ActorID              int64           `gorm:"column:actor_id"`
	CreatedAt            time.Time       `gorm:"column:created_at"`
}

func (orchestrationRevisionRecord) TableName() string {
	return "orchestration_resource_revisions"
}

func orchestrationRevisionRecordFromDomain(
	revision orchestrationcontrol.ResourceRevision,
	scope orchestrationcontrol.Scope,
) (orchestrationRevisionRecord, error) {
	if err := revision.Validate(scope); err != nil {
		return orchestrationRevisionRecord{}, err
	}
	resolved, err := orchestrationcontrol.CanonicalJSONArray(
		revision.ResolvedReferences,
	)
	if err != nil {
		return orchestrationRevisionRecord{}, err
	}
	var snapshotID *int64
	if revision.WorkerSpecSnapshotID > 0 {
		value := revision.WorkerSpecSnapshotID
		snapshotID = &value
	}
	return orchestrationRevisionRecord{
		OrganizationID: revision.OrganizationID, ResourceID: revision.ResourceID,
		Revision: revision.Revision, Generation: revision.Generation,
		ResourceVersion:   revision.ResourceVersion,
		CanonicalManifest: revision.CanonicalManifest,
		CanonicalSpec:     revision.CanonicalSpec, ResolvedReferences: resolved,
		Digest: revision.Digest, WorkerSpecSnapshotID: snapshotID,
		ActorID: revision.ActorID, CreatedAt: revision.CreatedAt,
	}, nil
}

func (repo *orchestrationResourceRepo) revisionDomain(
	ctx context.Context,
	scope orchestrationcontrol.Scope,
	record orchestrationRevisionRecord,
) (orchestrationcontrol.ResourceRevision, error) {
	var head orchestrationResourceRecord
	err := repo.db.WithContext(ctx).
		Where(
			"organization_id = ? AND id = ?",
			scope.OrganizationID,
			record.ResourceID,
		).
		First(&head).Error
	if err != nil {
		return orchestrationcontrol.ResourceRevision{}, corruptRecord("revision head")
	}
	return orchestrationRevisionDomain(scope, record, head)
}

func orchestrationRevisionDomain(
	scope orchestrationcontrol.Scope,
	record orchestrationRevisionRecord,
	head orchestrationResourceRecord,
) (orchestrationcontrol.ResourceRevision, error) {
	manifest, err := orchestrationcontrol.CanonicalJSONObject(record.CanonicalManifest)
	if err != nil {
		return orchestrationcontrol.ResourceRevision{}, corruptRecord("revision manifest")
	}
	spec, err := orchestrationcontrol.CanonicalJSONObject(record.CanonicalSpec)
	if err != nil {
		return orchestrationcontrol.ResourceRevision{}, corruptRecord("revision spec")
	}
	resolvedJSON, err := orchestrationcontrol.CanonicalJSONArray(record.ResolvedReferences)
	if err != nil {
		return orchestrationcontrol.ResourceRevision{}, corruptRecord("revision references")
	}
	var resolved []orchestrationcontrol.ResolvedReference
	if err := decodeStrictJSON(resolvedJSON, &resolved); err != nil {
		return orchestrationcontrol.ResourceRevision{}, corruptRecord("revision references")
	}
	snapshotID := int64(0)
	if record.WorkerSpecSnapshotID != nil {
		snapshotID = *record.WorkerSpecSnapshotID
	}
	revision := orchestrationcontrol.ResourceRevision{
		OrganizationID: record.OrganizationID, ResourceID: record.ResourceID,
		Identity: head.identity(), Revision: record.Revision,
		Generation: record.Generation, ResourceVersion: record.ResourceVersion,
		CanonicalManifest: manifest, CanonicalSpec: spec,
		ResolvedReferences: resolved, Digest: record.Digest,
		WorkerSpecSnapshotID: snapshotID, ActorID: record.ActorID,
		CreatedAt: record.CreatedAt.UTC(),
	}
	if err := revision.Validate(scope); err != nil {
		return orchestrationcontrol.ResourceRevision{}, corruptRecord("resource revision")
	}
	return revision, nil
}

func (record orchestrationResourceRecord) identity() orchestrationcontrol.ResourceIdentity {
	return orchestrationcontrol.ResourceIdentity{
		ResourceTarget: orchestrationcontrol.ResourceTarget{
			TypeMeta:  structTypeMeta(record.APIVersion, record.Kind),
			Namespace: stringSlug(record.Namespace),
			Name:      stringSlug(record.Name),
		},
		UID: record.UID,
	}
}
