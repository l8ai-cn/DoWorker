package workerconfigmigration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type migrationPlan struct {
	report    Report
	snapshots []snapshotUpdate
	revisions []revisionUpdate
}

type snapshotRecord struct {
	ID          int64           `gorm:"column:id"`
	SpecJSON    json.RawMessage `gorm:"column:spec_json"`
	SummaryJSON json.RawMessage `gorm:"column:summary_json"`
}

type revisionRecord struct {
	ID                int64           `gorm:"column:id"`
	CanonicalManifest json.RawMessage `gorm:"column:canonical_manifest"`
	CanonicalSpec     json.RawMessage `gorm:"column:canonical_spec"`
	Digest            string          `gorm:"column:digest"`
}

type pendingPlanRecord struct {
	ID                string          `gorm:"column:id"`
	CanonicalManifest json.RawMessage `gorm:"column:canonical_manifest"`
}

type snapshotUpdate struct {
	id       int64
	specJSON json.RawMessage
}

type revisionUpdate struct {
	id       int64
	manifest json.RawMessage
	spec     json.RawMessage
	digest   string
}

func (m *Migrator) scan(ctx context.Context, db *gorm.DB) (migrationPlan, error) {
	plan := migrationPlan{}
	if err := m.scanSnapshots(ctx, db, &plan); err != nil {
		return migrationPlan{}, err
	}
	if err := m.scanRevisions(ctx, db, &plan); err != nil {
		return migrationPlan{}, err
	}
	if err := m.scanPendingPlans(ctx, db, &plan); err != nil {
		return migrationPlan{}, err
	}
	return plan, nil
}

func (m *Migrator) scanSnapshots(
	ctx context.Context,
	db *gorm.DB,
	plan *migrationPlan,
) error {
	var records []snapshotRecord
	if err := db.WithContext(ctx).Table("worker_spec_snapshots").
		Order("id").Find(&records).Error; err != nil {
		return err
	}
	for _, record := range records {
		plan.report.SnapshotsScanned++
		update, changed, err := m.migrateSnapshot(record)
		if err != nil {
			plan.report.Blockers = append(
				plan.report.Blockers,
				fmt.Sprintf("worker spec snapshot %d: %v", record.ID, err),
			)
			continue
		}
		if changed {
			plan.snapshots = append(plan.snapshots, update)
			plan.report.SnapshotUpdates++
		}
	}
	return nil
}

func (m *Migrator) scanRevisions(
	ctx context.Context,
	db *gorm.DB,
	plan *migrationPlan,
) error {
	var records []revisionRecord
	if err := db.WithContext(ctx).Table("orchestration_resource_revisions AS revision").
		Select("revision.id, revision.canonical_manifest, revision.canonical_spec, revision.digest").
		Joins("JOIN orchestration_resources AS resource ON resource.id = revision.resource_id").
		Where("resource.kind = ?", "WorkerTemplate").Order("revision.id").
		Find(&records).Error; err != nil {
		return err
	}
	for _, record := range records {
		plan.report.WorkerTemplateRevisions++
		update, changed, err := m.migrateRevision(record)
		if err != nil {
			plan.report.Blockers = append(
				plan.report.Blockers,
				fmt.Sprintf("WorkerTemplate revision %d: %v", record.ID, err),
			)
			continue
		}
		if changed {
			plan.revisions = append(plan.revisions, update)
			plan.report.WorkerTemplateUpdates++
		}
	}
	return nil
}

func (m *Migrator) scanPendingPlans(
	ctx context.Context,
	db *gorm.DB,
	plan *migrationPlan,
) error {
	var records []pendingPlanRecord
	if err := db.WithContext(ctx).Table("orchestration_resource_plans").
		Where(
			"target_kind = ? AND consumed_at IS NULL AND expires_at > ?",
			"WorkerTemplate",
			time.Now().UTC(),
		).
		Order("id").Find(&records).Error; err != nil {
		return err
	}
	for _, record := range records {
		legacy, err := hasLegacyTemplateConfig(record.CanonicalManifest)
		if err != nil {
			plan.report.Blockers = append(
				plan.report.Blockers,
				fmt.Sprintf("pending WorkerTemplate plan %s: %v", record.ID, err),
			)
			continue
		}
		if legacy {
			plan.report.PendingLegacyPlans++
			plan.report.Blockers = append(
				plan.report.Blockers,
				fmt.Sprintf("pending WorkerTemplate plan %s must be regenerated", record.ID),
			)
		}
	}
	return nil
}
