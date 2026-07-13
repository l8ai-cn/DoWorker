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
