package acp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestACPClientRejectsUnknownAdapterBeforeStartingCommand(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "started")
	script := `IFS= read -r request
touch "$1"
id=$(printf '%s' "$request" | sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p')
printf '{"jsonrpc":"2.0","id":%s,"error":{"code":-32603,"message":"started"}}\n' "$id"`
	client := NewClient(ClientConfig{
		Command:       "sh",
		Args:          []string{"-c", script, "sh", marker},
		TransportType: "unknown-adapter",
	})
	defer client.Stop()

	err := client.Start()

	assert.ErrorContains(t, err, "unknown ACP adapter")
	_, statErr := os.Stat(marker)
	assert.True(t, os.IsNotExist(statErr), "unknown adapter must not start a process")
}
