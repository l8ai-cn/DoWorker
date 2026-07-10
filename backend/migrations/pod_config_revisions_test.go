package migrations

import (
	"strings"
	"testing"
)

func TestMigration000195PodConfigRevisions(t *testing.T) {
	up, err := FS.ReadFile("000195_pod_config_revisions.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upSQL := string(up)
	for _, fragment := range []string{
		"ALTER TABLE pods ADD COLUMN model_resource_id BIGINT",
		"ALTER TABLE pods ADD COLUMN generation BIGINT NOT NULL DEFAULT 0",
		"CONSTRAINT pods_generation_nonnegative CHECK (generation >= 0)",
		"ALTER TABLE pods ADD COLUMN active_config_revision_id BIGINT",
		"ALTER TABLE pods ADD COLUMN pending_config_revision_id BIGINT",
		"ALTER TABLE pods ADD COLUMN reinitialize_dispatched_at TIMESTAMPTZ",
		"ALTER TABLE pods ADD COLUMN archived_at TIMESTAMPTZ",
		"ALTER TABLE pods ADD COLUMN archived_by_id BIGINT",
		"ALTER TABLE pods ADD COLUMN purge_after TIMESTAMPTZ",
		"FOREIGN KEY (model_resource_id) REFERENCES model_resources(id) ON DELETE SET NULL",
		"FOREIGN KEY (archived_by_id) REFERENCES users(id) ON DELETE SET NULL",
		"CREATE TABLE pod_config_revisions",
		"CONSTRAINT pod_config_revisions_pod_revision_key UNIQUE (pod_id, revision)",
		"CONSTRAINT pod_config_revisions_revision_positive CHECK (revision > 0)",
		"CONSTRAINT pod_config_revisions_status_check CHECK (status IN ('draft', 'applying', 'active', 'failed'))",
		"CONSTRAINT pod_config_revisions_config_summary_object CHECK (jsonb_typeof(config_summary) = 'object')",
		"FOREIGN KEY (pod_id) REFERENCES pods(id) ON DELETE CASCADE",
		"FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE RESTRICT",
		"CREATE INDEX idx_pods_model_resource_id",
		"CREATE INDEX idx_pods_active_config_revision_id",
		"CREATE INDEX idx_pods_pending_config_revision_id",
		"CREATE INDEX idx_pod_config_revisions_pod_id",
		"CREATE INDEX idx_pod_config_revisions_status",
		"CREATE INDEX idx_pod_config_revisions_model_resource_id",
		"CREATE UNIQUE INDEX idx_pod_config_revisions_one_active_per_pod",
		"FOREIGN KEY (active_config_revision_id) REFERENCES pod_config_revisions(id) ON DELETE SET NULL",
		"FOREIGN KEY (pending_config_revision_id) REFERENCES pod_config_revisions(id) ON DELETE SET NULL",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Errorf("up migration must contain %q", fragment)
		}
	}

	down, err := FS.ReadFile("000195_pod_config_revisions.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downSQL := string(down)
	for _, fragment := range []string{
		"ALTER TABLE pods DROP CONSTRAINT IF EXISTS pods_pending_config_revision_id_fkey",
		"ALTER TABLE pods DROP CONSTRAINT IF EXISTS pods_active_config_revision_id_fkey",
		"ALTER TABLE pods DROP CONSTRAINT IF EXISTS pods_generation_nonnegative",
		"DROP INDEX IF EXISTS idx_pod_config_revisions_one_active_per_pod",
		"DROP INDEX IF EXISTS idx_pods_model_resource_id",
		"DROP TABLE IF EXISTS pod_config_revisions",
		"ALTER TABLE pods DROP COLUMN IF EXISTS active_config_revision_id",
		"ALTER TABLE pods DROP COLUMN IF EXISTS generation",
		"ALTER TABLE pods DROP COLUMN IF EXISTS model_resource_id",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Errorf("down migration must contain %q", fragment)
		}
	}

	constraintIndex := strings.Index(downSQL, "ALTER TABLE pods DROP CONSTRAINT IF EXISTS pods_pending_config_revision_id_fkey")
	tableIndex := strings.Index(downSQL, "DROP TABLE IF EXISTS pod_config_revisions")
	columnIndex := strings.Index(downSQL, "ALTER TABLE pods DROP COLUMN IF EXISTS active_config_revision_id")
	if constraintIndex < 0 || tableIndex < 0 || columnIndex < 0 {
		t.Fatal("down migration must drop constraints, table, and columns")
	}
	if constraintIndex > tableIndex || tableIndex > columnIndex {
		t.Error("down migration must drop constraints before table and columns")
	}
}
