package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration000210SkillTagsContract(t *testing.T) {
	up, err := FS.ReadFile("000219_skill_tags.up.sql")
	require.NoError(t, err)
	upSQL := string(up)

	for _, fragment := range []string{
		"ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}'",
		"CREATE INDEX idx_skills_tags ON skills USING GIN (tags)",
	} {
		require.Contains(t, upSQL, fragment)
	}

	down, err := FS.ReadFile("000219_skill_tags.down.sql")
	require.NoError(t, err)
	downSQL := string(down)

	require.Contains(t, downSQL, "DROP INDEX IF EXISTS idx_skills_tags")
	require.Contains(t, downSQL, "DROP COLUMN IF EXISTS tags")
}
