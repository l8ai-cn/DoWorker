package workbench

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMP4DecodeArgsRequireStrictFullDecode(t *testing.T) {
	args := mp4DecodeArgs("video.mp4")

	require.Contains(t, args, "-xerror")
	require.NotContains(t, args, "-frames:v")
	require.Equal(t, "video.mp4", args[4])
}

func TestMP4DecodeArgsPreserveWindowsInputPath(t *testing.T) {
	path := `C:\Users\RUNNER~1\AppData\Local\Temp\video.mp4`

	require.Equal(t, path, mp4DecodeArgs(path)[4])
}

func TestVerifyMP4DecodeAcceptsDecodableVideo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "video.mp4")
	writeDecodableMP4(t, path)

	require.NoError(t, verifyMP4Decode(context.Background(), path))
}

func TestVerifyMP4DecodeRejectsCorruptedTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "video.mp4")
	writeCorruptedTailMP4(t, path)

	require.NoError(t, decodeFirstMP4Frame(path))
	require.Error(t, verifyMP4Decode(context.Background(), path))
}

func writeCorruptedTailMP4(t *testing.T, path string) {
	t.Helper()
	writeDecodableMP4(t, path)
	corruptLaterMP4Packet(t, path)
}

func writeDecodableMP4(t *testing.T, path string) {
	t.Helper()
	output, err := exec.Command(
		"ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", "testsrc2=s=64x64:d=2:r=10",
		"-c:v", "libx264",
		"-g", "1",
		"-keyint_min", "1",
		"-bf", "0",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		path,
	).CombinedOutput()
	require.NoErrorf(t, err, "generate MP4 fixture: %s", output)
}

func corruptLaterMP4Packet(t *testing.T, path string) {
	t.Helper()
	output, err := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "packet=pos,size",
		"-of", "json",
		path,
	).Output()
	require.NoError(t, err)
	var report mp4PacketReport
	require.NoError(t, json.Unmarshal(output, &report))
	require.Greater(t, len(report.Packets), 2)
	packet := report.Packets[len(report.Packets)*3/4]
	position, err := strconv.ParseInt(packet.Position, 10, 64)
	require.NoError(t, err)
	size, err := strconv.ParseInt(packet.Size, 10, 64)
	require.NoError(t, err)
	require.GreaterOrEqual(t, size, int64(4))
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.GreaterOrEqual(t, int64(len(content))-position, int64(4))
	copy(content[position:position+4], []byte{0x7f, 0xff, 0xff, 0xff})
	require.NoError(t, os.WriteFile(path, content, 0o644))
}

type mp4PacketReport struct {
	Packets []mp4Packet `json:"packets"`
}

type mp4Packet struct {
	Position string `json:"pos"`
	Size     string `json:"size"`
}

func decodeFirstMP4Frame(path string) error {
	args := mp4DecodeArgs(path)
	args = append(args[:5],
		"-frames:v", "1",
		"-f", "null", os.DevNull,
	)
	return exec.Command("ffmpeg", args...).Run()
}
