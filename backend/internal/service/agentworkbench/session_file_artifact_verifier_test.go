package agentworkbench

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyMP4DecodesFailsClosedWhenDecodeFails(t *testing.T) {
	dir := t.TempDir()
	writeVerifierTool(t, dir, "ffprobe", "#!/bin/sh\nexit 0\n")
	writeVerifierTool(t, dir, "ffmpeg", "#!/bin/sh\nprintf 'decode failed' >&2\nexit 23\n")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	video := filepath.Join(dir, "video.mp4")
	require.NoError(t, os.WriteFile(video, []byte("not-a-real-video"), 0o644))

	err := verifyMP4Decodes(context.Background(), video)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ffmpeg artifact verification failed")
	assert.Contains(t, err.Error(), "decode failed")
}

func TestVerifyMP4DecodesDoesNotLimitFrameCount(t *testing.T) {
	dir := t.TempDir()
	writeVerifierTool(t, dir, "ffprobe", "#!/bin/sh\nexit 0\n")
	writeVerifierTool(t, dir, "ffmpeg", `#!/bin/sh
strict=false
for arg in "$@"; do
  if [ "$arg" = "-xerror" ]; then
    strict=true
  fi
  if [ "$arg" = "-frames:v" ]; then
    printf 'frame limit bypasses full verification' >&2
    exit 23
  fi
done
if [ "$strict" != true ]; then
  printf 'decode errors are not fail-closed' >&2
  exit 23
fi
exit 0
`)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	video := filepath.Join(dir, "video.mp4")
	require.NoError(t, os.WriteFile(video, []byte("video"), 0o644))

	require.NoError(t, verifyMP4Decodes(context.Background(), video))
}

func writeVerifierTool(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o755))
}
