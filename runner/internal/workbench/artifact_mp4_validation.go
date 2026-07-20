package workbench

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const mp4ValidationTimeout = 30 * time.Second

func validateDeclaredFileContent(
	root *os.Root,
	display string,
	mediaType string,
) error {
	if mediaType != "video/mp4" {
		return nil
	}
	if _, err := root.Stat(filepath.FromSlash(display)); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), mp4ValidationTimeout)
	defer cancel()
	hostPath := filepath.Join(root.Name(), filepath.FromSlash(display))
	if err := requireMediaTool("ffprobe"); err != nil {
		return err
	}
	if err := requireMediaTool("ffmpeg"); err != nil {
		return err
	}
	if err := verifyMP4Probe(ctx, hostPath); err != nil {
		return invalidMP4Error(display, err)
	}
	if err := verifyMP4Decode(ctx, hostPath); err != nil {
		return invalidMP4Error(display, err)
	}
	return nil
}

func requireMediaTool(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s is required to verify MP4 artifacts", name)
	}
	return nil
}

func verifyMP4Probe(ctx context.Context, path string) error {
	output, err := exec.CommandContext(
		ctx,
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name,width,height",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffprobe failed: %s", strings.TrimSpace(string(output)))
	}
	if strings.TrimSpace(string(output)) == "" {
		return fmt.Errorf("ffprobe found no video stream")
	}
	return nil
}

func verifyMP4Decode(ctx context.Context, path string) error {
	output, err := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-xerror",
		"-v", "error",
		"-i", path,
		"-map", "0:v:0",
		"-f", "null",
		os.DevNull,
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg decode failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func invalidMP4Error(display string, cause error) error {
	return fmt.Errorf("workspace path %q is not a decodable MP4 file: %w", display, cause)
}
