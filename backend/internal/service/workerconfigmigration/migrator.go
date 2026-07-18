package workerconfigmigration

import (
	"context"
	"fmt"

	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	"gorm.io/gorm"
)

type DefinitionCatalog interface {
	Get(string) (workerdefinition.Definition, bool)
}

type Migrator struct {
	db          *gorm.DB
	definitions DefinitionCatalog
	registry    *resource.Registry
}

func New(db *gorm.DB, definitions DefinitionCatalog) (*Migrator, error) {
	if db == nil || definitions == nil {
		return nil, fmt.Errorf("worker config document migrator dependencies are incomplete")
	}
	registry := resource.NewRegistry()
	if err := resource.RegisterWorkerSchemas(registry); err != nil {
		return nil, fmt.Errorf("register worker resource schemas: %w", err)
	}
	return &Migrator{db: db, definitions: definitions, registry: registry}, nil
}

func (m *Migrator) Check(ctx context.Context) (Report, error) {
	plan, err := m.scan(ctx, m.db.WithContext(ctx))
	return plan.report, err
}

func (m *Migrator) Run(ctx context.Context) (Report, error) {
	var plan migrationPlan
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockMigrationTables(tx); err != nil {
			return err
		}
		var err error
		plan, err = m.scan(ctx, tx)
		if err != nil {
			return err
		}
		if !plan.report.Ready() {
			return fmt.Errorf(
				"worker config document migration is blocked by %d record(s)",
				len(plan.report.Blockers),
			)
		}
		if plan.report.Clean() {
			return nil
		}
		if err := removeImmutableTriggers(tx); err != nil {
			return err
		}
		if err := applyPlan(tx, plan); err != nil {
			return err
		}
		return restoreImmutableTriggers(tx)
	})
	return plan.report, err
}

func lockMigrationTables(tx *gorm.DB) error {
	return tx.Exec(`
LOCK TABLE worker_spec_snapshots, orchestration_resource_revisions,
  orchestration_resource_plans IN ACCESS EXCLUSIVE MODE`).Error
}

func removeImmutableTriggers(tx *gorm.DB) error {
	for _, statement := range []string{
		"DROP TRIGGER worker_spec_snapshots_immutable ON worker_spec_snapshots",
		"DROP TRIGGER orchestration_resource_revisions_immutable ON orchestration_resource_revisions",
	} {
		if err := tx.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func restoreImmutableTriggers(tx *gorm.DB) error {
	for _, statement := range []string{
		`CREATE TRIGGER worker_spec_snapshots_immutable
BEFORE UPDATE ON worker_spec_snapshots
FOR EACH ROW EXECUTE FUNCTION prevent_worker_spec_snapshot_update()`,
		`CREATE TRIGGER orchestration_resource_revisions_immutable
BEFORE UPDATE OR DELETE ON orchestration_resource_revisions
FOR EACH ROW EXECUTE FUNCTION prevent_orchestration_resource_revision_mutation()`,
	} {
		if err := tx.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func applyPlan(tx *gorm.DB, plan migrationPlan) error {
	for _, update := range plan.snapshots {
		if err := tx.Table("worker_spec_snapshots").Where("id = ?", update.id).
			Update("spec_json", update.specJSON).Error; err != nil {
			return err
		}
	}
	for _, update := range plan.revisions {
		if err := tx.Table("orchestration_resource_revisions").
			Where("id = ?", update.id).Updates(map[string]any{
			"canonical_manifest": update.manifest,
			"canonical_spec":     update.spec,
			"digest":             update.digest,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
