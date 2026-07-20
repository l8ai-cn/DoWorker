package agentworkbench

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	sessionfiledomain "github.com/anthropics/agentsmesh/backend/internal/domain/sessionfile"
	"github.com/anthropics/agentsmesh/backend/internal/infra/storage"
	filesvc "github.com/anthropics/agentsmesh/backend/internal/service/file"
	sessionfilesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionfile"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
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

func TestSessionFileArtifactVerifierUsesOneObjectSnapshotForMP4(t *testing.T) {
	dir := t.TempDir()
	writeVerifierTool(t, dir, "ffprobe", "#!/bin/sh\nexit 0\n")
	writeVerifierTool(t, dir, "ffmpeg", "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	content := []byte("video")
	sum := sha256.Sum256(content)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	bytes := uint64(len(content))
	objects := storage.NewMockStorage()
	objects.PutFileData("artifacts/video.mp4", content, "video/mp4")
	counted := &countingArtifactStorage{Storage: objects}
	files := sessionfilesvc.NewService(
		nil,
		filesvc.NewService(counted, config.StorageConfig{}),
	)
	verifier := NewSessionFileArtifactVerifier(files)

	err := verifier.Verify(context.Background(), &sessionfiledomain.File{
		Bytes: int64(len(content)), MinioKey: "artifacts/video.mp4",
	}, &agentworkbenchv2.ArtifactRepresentation{
		MediaType: "video/mp4", Digest: &digest, ByteSize: &bytes,
	})

	require.NoError(t, err)
	require.Equal(t, 1, counted.downloads)
}

type countingArtifactStorage struct {
	storage.Storage
	downloads int
}

func (s *countingArtifactStorage) Download(
	ctx context.Context,
	key string,
) (io.ReadCloser, int64, error) {
	s.downloads++
	return s.Storage.Download(ctx, key)
}

func writeVerifierTool(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o755))
}
