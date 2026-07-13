package extension

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSkillDir_ReadsNormalizedSkillConfigTags(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "SKILL.md"),
		[]byte("---\nname: video-editing\n---\n"),
		0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "skill.json"),
		[]byte(`{"schema":2,"slug":"video-editing","tags":[" Video ","editing","VIDEO"]}`),
		0644,
	))

	info, err := parseSkillDir(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{"editing", "video"}, info.Tags)
}

func TestParseSkillDir_RejectsSkillConfigSymlinkOutsideClone(t *testing.T) {
	dir := t.TempDir()
	outsideDir := t.TempDir()
	outsidePath := filepath.Join(outsideDir, "skill.json")
	outsideContent := []byte(`{"schema":2,"tags":["outside"]}`)
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "SKILL.md"),
		[]byte("---\nname: video-editing\n---\n"),
		0644,
	))
	require.NoError(t, os.WriteFile(outsidePath, outsideContent, 0644))
	require.NoError(t, os.Symlink(outsidePath, filepath.Join(dir, "skill.json")))

	_, err := parseSkillDir(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "regular file")
	assert.Equal(t, outsideContent, mustReadFile(t, outsidePath))
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return content
}
