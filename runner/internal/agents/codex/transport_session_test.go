package codex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeadlessAutomationFields(t *testing.T) {
	params := headlessAutomationFields("/tmp/ws")
	assert.Equal(t, "never", params["approvalPolicy"])
	assert.Equal(t, "danger-full-access", params["sandbox"])
	assert.Equal(t, "/tmp/ws", params["cwd"])
}

func TestMergeHeadlessFields(t *testing.T) {
	params := mergeHeadlessFields(map[string]any{"threadId": "t1"}, "/ws")
	assert.Equal(t, "t1", params["threadId"])
	assert.Equal(t, "/ws", params["cwd"])
	assert.Equal(t, "danger-full-access", params["sandbox"])
}
