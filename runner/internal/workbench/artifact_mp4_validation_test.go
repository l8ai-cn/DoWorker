package workbench

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifyMP4DecodeDoesNotLimitFrameCount(t *testing.T) {
	dir := t.TempDir()
	writeMP4ValidationTool(t, dir, "ffmpeg", `#!/bin/sh
for arg in "$@"; do
  if [ "$arg" = "-frames:v" ]; then
    printf 'frame limit bypasses full verification' >&2
    exit 23
  fi
done
exit 0
`)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	require.NoError(t, verifyMP4Decode(context.Background(), filepath.Join(dir, "video.mp4")))
}

func TestVerifyMP4DecodeRejectsCorruptedTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "video.mp4")
	writeCorruptedTailMP4(t, path)

	require.NoError(t, decodeFirstMP4Frame(path))
	require.Error(t, verifyMP4Decode(context.Background(), path))
}

func writeCorruptedTailMP4(t *testing.T, path string) {
	t.Helper()
	output, err := exec.Command(
		"ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", "testsrc2=s=64x64:d=6:r=30",
		"-c:v", "libx264",
		"-g", "90",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		path,
	).CombinedOutput()
	require.NoErrorf(t, err, "generate MP4 fixture: %s", output)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	start := len(content) * 9 / 10
	for i := start; i < start+2048 && i < len(content); i++ {
		content[i] = 0
	}
	require.NoError(t, os.WriteFile(path, content, 0o644))
}

func decodeFirstMP4Frame(path string) error {
	return exec.Command(
		"ffmpeg",
		"-xerror",
		"-v", "error",
		"-i", path,
		"-map", "0:v:0",
		"-frames:v", "1",
		"-f", "null",
		os.DevNull,
	).Run()
}

func writeMP4ValidationTool(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o755))
}
