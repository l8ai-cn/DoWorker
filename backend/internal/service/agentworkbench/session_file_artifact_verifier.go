package agentworkbench

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	sessionfiledomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/sessionfile"
	sessionfilesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessionfile"
	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
)

const artifactMediaVerifyTimeout = 30 * time.Second

type ArtifactVerifier interface {
	Verify(
		context.Context,
		*sessionfiledomain.File,
		*agentworkbenchv2.ArtifactRepresentation,
	) error
}

type SessionFileArtifactVerifier struct {
	files *sessionfilesvc.Service
}

func NewSessionFileArtifactVerifier(
	files *sessionfilesvc.Service,
) *SessionFileArtifactVerifier {
	return &SessionFileArtifactVerifier{files: files}
}

func (v *SessionFileArtifactVerifier) Verify(
	ctx context.Context,
	file *sessionfiledomain.File,
	representation *agentworkbenchv2.ArtifactRepresentation,
) error {
	if v == nil || v.files == nil || file == nil || representation == nil {
		return ErrIngressConfiguration
	}
	reader, size, err := v.files.Open(ctx, file)
	if err != nil {
		return fmt.Errorf("open artifact snapshot: %w", err)
	}
	defer reader.Close()
	if size >= 0 && size != file.Bytes {
		return fmt.Errorf("artifact snapshot size mismatch")
	}
	if representation.GetMediaType() == "video/mp4" {
		return v.verifyVideo(ctx, reader, file, representation)
	}
	return verifyArtifactStream(reader, file.Bytes, representation.GetDigest())
}

func (v *SessionFileArtifactVerifier) verifyVideo(
	ctx context.Context,
	reader io.Reader,
	file *sessionfiledomain.File,
	representation *agentworkbenchv2.ArtifactRepresentation,
) error {
	tmp, err := os.CreateTemp("", "agentcloud-artifact-*.mp4")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	hash := sha256.New()
	written, copyErr := io.Copy(tmp, io.TeeReader(reader, hash))
	closeErr := tmp.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := verifyArtifactDigest(
		written,
		file.Bytes,
		representation.GetDigest(),
		hash,
	); err != nil {
		return err
	}
	return verifyMP4Decodes(ctx, tmpPath)
}

func verifyArtifactStream(
	reader io.Reader,
	expectedBytes int64,
	expectedDigest string,
) error {
	hash := sha256.New()
	written, err := io.Copy(hash, reader)
	if err != nil {
		return err
	}
	return verifyArtifactDigest(written, expectedBytes, expectedDigest, hash)
}

func verifyArtifactDigest(
	actualBytes int64,
	expectedBytes int64,
	expectedDigest string,
	hash interface{ Sum([]byte) []byte },
) error {
	actualDigest := "sha256:" + hex.EncodeToString(hash.Sum(nil))
	if actualBytes != expectedBytes || actualDigest != expectedDigest {
		return fmt.Errorf("artifact snapshot integrity mismatch")
	}
	return nil
}

func verifyMP4Decodes(ctx context.Context, path string) error {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return fmt.Errorf("ffprobe is required to verify MP4 artifacts")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg is required to verify MP4 artifacts")
	}
	verifyCtx, cancel := context.WithTimeout(ctx, artifactMediaVerifyTimeout)
	defer cancel()
	if err := runMediaCommand(
		verifyCtx,
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name,width,height,duration",
		"-of", "json",
		path,
	); err != nil {
		return err
	}
	return runMediaCommand(
		verifyCtx,
		"ffmpeg",
		"-xerror",
		"-v", "error",
		"-i", path,
		"-map", "0:v:0",
		"-f", "null",
		os.DevNull,
	)
}

func runMediaCommand(ctx context.Context, name string, args ...string) error {
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s artifact verification failed: %s", filepath.Base(name), string(output))
	}
	return nil
}
