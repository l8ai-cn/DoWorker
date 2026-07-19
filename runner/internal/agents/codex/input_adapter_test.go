package codex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodexInputAdapter_SingleLine(t *testing.T) {
	a := &codexInputAdapter{}
	assert.Equal(t, []byte("\x1b[200~hello\x1b[201~"), a.Adapt([]byte("hello")))
}

func TestCodexInputAdapter_SingleLineWithEnter(t *testing.T) {
	a := &codexInputAdapter{}
	assert.Equal(t, []byte("\x1b[200~hello\x1b[201~\r"), a.Adapt([]byte("hello\r")))
}

func TestCodexInputAdapter_MultiLine(t *testing.T) {
	a := &codexInputAdapter{}
	result := a.Adapt([]byte("line1\nline2\r"))
	assert.Equal(t, []byte("\x1b[200~line1\nline2\x1b[201~\r"), result)
}

func TestCodexInputAdapter_CRLF(t *testing.T) {
	a := &codexInputAdapter{}
	result := a.Adapt([]byte("a\r\nb\r\nc\r"))
	assert.Equal(t, []byte("\x1b[200~a\nb\nc\x1b[201~\r"), result)
}

func TestCodexInputAdapter_OnlyNewlines(t *testing.T) {
	a := &codexInputAdapter{}
	assert.Equal(t, []byte("\x1b[200~\n\n\x1b[201~\r"), a.Adapt([]byte("\n\n\r")))
}

func TestCodexInputAdapter_Empty(t *testing.T) {
	a := &codexInputAdapter{}
	assert.Equal(t, []byte{}, a.Adapt([]byte{}))
}

func TestCodexInputAdapter_NoTrailingEnter(t *testing.T) {
	a := &codexInputAdapter{}
	assert.Equal(t, []byte("\x1b[200~a\nb\x1b[201~"), a.Adapt([]byte("a\nb")))
}
