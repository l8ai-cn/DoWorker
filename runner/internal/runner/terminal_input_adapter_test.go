package runner

import (
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/stretchr/testify/assert"
)

func TestAdaptTerminalInput_NonCodex(t *testing.T) {
	data := []byte("hello\nworld\r")
	assert.Equal(t, data, adaptTerminalInput(data, "claude-code"))
	assert.Equal(t, data, adaptTerminalInput(data, "aider"))
	assert.Equal(t, data, adaptTerminalInput(data, ""))
}

func TestAdaptTerminalInput_CodexSingleLine(t *testing.T) {
	assert.Equal(t, []byte("\x1b[200~hello\x1b[201~"), adaptTerminalInput([]byte("hello"), "codex-cli"))
}

func TestAdaptTerminalInput_CodexSingleLineWithEnter(t *testing.T) {
	assert.Equal(t, []byte("\x1b[200~hello\x1b[201~\r"), adaptTerminalInput([]byte("hello\r"), "codex-cli"))
}

func TestAdaptTerminalInput_CodexMultiLine(t *testing.T) {
	input := []byte("Message from channel(#dev): fix the bug\n\nPlease reply.\r")
	result := adaptTerminalInput(input, "codex-cli")
	assert.Equal(t, []byte("\x1b[200~Message from channel(#dev): fix the bug\n\nPlease reply.\x1b[201~\r"), result)
}

func TestAdaptTerminalInput_CodexCRLF(t *testing.T) {
	input := []byte("line1\r\nline2\r\nline3\r")
	result := adaptTerminalInput(input, "codex")
	assert.Equal(t, []byte("\x1b[200~line1\nline2\nline3\x1b[201~\r"), result)
}

func TestAdaptTerminalInput_CodexOnlyNewlines(t *testing.T) {
	result := adaptTerminalInput([]byte("\n\n\r"), "codex-cli")
	assert.Equal(t, []byte("\x1b[200~\n\n\x1b[201~\r"), result)
}

func TestAdaptTerminalInput_CodexEmpty(t *testing.T) {
	assert.Equal(t, []byte{}, adaptTerminalInput([]byte{}, "codex"))
}

func TestAdaptTerminalInput_CodexNoTrailingEnter(t *testing.T) {
	input := []byte("line1\nline2")
	result := adaptTerminalInput(input, "codex-cli")
	assert.Equal(t, []byte("\x1b[200~line1\nline2\x1b[201~"), result)
}

func TestAdaptTerminalInput_BothCodexSlugs(t *testing.T) {
	input := []byte("hello\nworld\r")
	expected := []byte("\x1b[200~hello\nworld\x1b[201~\r")
	assert.Equal(t, expected, adaptTerminalInput(input, "codex"))
	assert.Equal(t, expected, adaptTerminalInput(input, "codex-cli"))
}

func TestAdaptTerminalInput_VideoStudioEntrypoint(t *testing.T) {
	input := []byte("请生成视频\n输出 MP4")
	expected := []byte("\x1b[200~请生成视频\n输出 MP4\x1b[201~")
	assert.Equal(t, expected, adaptTerminalInput(input, "video-studio-codex"))
}

type testAdapter struct{}

func (a *testAdapter) Adapt(data []byte) []byte {
	return []byte("adapted")
}

func TestAdaptTerminalInput_CustomAdapter(t *testing.T) {
	agentkit.RegisterInputAdapter("test-custom-agent", &testAdapter{})
	result := adaptTerminalInput([]byte("hello"), "test-custom-agent")
	assert.Equal(t, []byte("adapted"), result)
}
