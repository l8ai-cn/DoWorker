package acp

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterTransport_DuplicatePanics(t *testing.T) {
	RegisterTransport("dup-transport", func(_ EventCallbacks, _ *slog.Logger) Transport { return nil })
	defer func() {
		registryMu.Lock()
		delete(registry, "dup-transport")
		registryMu.Unlock()
	}()

	assert.Panics(t, func() {
		RegisterTransport("dup-transport", nil)
	})
}

func TestNewTransport_RejectsUnknownAdapter(t *testing.T) {
	tr, err := NewTransport("totally-unknown-transport", EventCallbacks{}, slog.Default())
	assert.Nil(t, tr)
	assert.ErrorContains(t, err, "unknown ACP adapter")
}
