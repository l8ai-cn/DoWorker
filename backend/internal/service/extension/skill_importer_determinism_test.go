package extension

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageSkillDirIsDeterministicAcrossFileTimestamps(t *testing.T) {
	dir := t.TempDir()
	skillPath := filepath.Join(dir, "SKILL.md")
	require.NoError(t, os.WriteFile(
		skillPath,
		[]byte("---\nname: deterministic\n---\n"),
		0644,
	))

	first, err := packageSkillDir(dir)
	require.NoError(t, err)

	later := time.Now().Add(24 * time.Hour)
	require.NoError(t, os.Chtimes(skillPath, later, later))
	second, err := packageSkillDir(dir)
	require.NoError(t, err)

	assert.Equal(t, first, second)
}

func TestPackageSkillDirIsDeterministicAcrossPermissions(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "scripts")
	skillPath := filepath.Join(dir, "SKILL.md")
	require.NoError(t, os.Mkdir(subdir, 0700))
	require.NoError(t, os.WriteFile(skillPath, []byte("content"), 0600))

	first, err := packageSkillDir(dir)
	require.NoError(t, err)

	require.NoError(t, os.Chmod(subdir, 0755))
	require.NoError(t, os.Chmod(skillPath, 0644))
	second, err := packageSkillDir(dir)
	require.NoError(t, err)

	assert.Equal(t, first, second)
}
