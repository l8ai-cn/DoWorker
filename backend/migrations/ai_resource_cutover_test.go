package migrations

import (
	"strings"
	"testing"
)

func TestMigration000198AIResourceCutover(t *testing.T) {
	up, err := FS.ReadFile("000198_ai_resource_cutover.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upSQL := string(up)
	for _, fragment := range []string{
		"ai_models must be migrated before AI resource cutover",
		"credential EnvBundles must be migrated before AI resource cutover",
		"UPDATE virtual_api_keys",
		"model_resource_id",
		"SET NOT NULL",
		"REFERENCES model_resources(id)",
		"DROP COLUMN ai_model_id",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Errorf("up migration must contain %q", fragment)
		}
	}

	down, err := FS.ReadFile("000198_ai_resource_cutover.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downSQL := string(down)
	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS ai_model_id",
		"UPDATE virtual_api_keys",
		"ai_model_id cannot be restored from AI resource migration map",
		"REFERENCES ai_models(id)",
		"DROP COLUMN IF EXISTS model_resource_id",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Errorf("down migration must contain %q", fragment)
		}
	}
}
