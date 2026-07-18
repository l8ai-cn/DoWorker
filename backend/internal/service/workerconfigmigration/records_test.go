package workerconfigmigration

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestScanPendingPlansIgnoresExpiredPlans(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`
CREATE TABLE orchestration_resource_plans (
	id TEXT PRIMARY KEY,
	target_kind TEXT NOT NULL,
	consumed_at DATETIME,
	expires_at DATETIME NOT NULL,
	canonical_manifest BLOB NOT NULL
)`).Error; err != nil {
		t.Fatal(err)
	}

	legacy := []byte(`{"spec":{"workspace":{"configBundleRefs":[]}}}`)
	now := time.Now().UTC()
	for _, record := range []struct {
		id      string
		expires time.Time
	}{
		{id: "expired", expires: now.Add(-time.Minute)},
		{id: "active", expires: now.Add(time.Minute)},
	} {
		if err := db.Exec(
			`INSERT INTO orchestration_resource_plans
			 (id, target_kind, expires_at, canonical_manifest)
			 VALUES (?, 'WorkerTemplate', ?, ?)`,
			record.id,
			record.expires,
			legacy,
		).Error; err != nil {
			t.Fatal(err)
		}
	}

	plan := migrationPlan{}
	migrator := Migrator{definitions: testDefinitions{"do-agent": definitionWithDocuments("settings")}}
	if err := migrator.scanPendingPlans(context.Background(), db, &plan); err != nil {
		t.Fatal(err)
	}
	if plan.report.PendingLegacyPlans != 1 {
		t.Fatalf("PendingLegacyPlans = %d, want 1", plan.report.PendingLegacyPlans)
	}
	if len(plan.report.Blockers) != 1 {
		t.Fatalf("blockers = %v, want one active-plan blocker", plan.report.Blockers)
	}
}
