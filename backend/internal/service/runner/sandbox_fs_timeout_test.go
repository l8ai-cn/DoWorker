package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSandboxFsCommandTimeoutAllowsArtifactUpload(t *testing.T) {
	assert.Equal(t, SandboxFsUploadTimeout, sandboxFsCommandTimeout("upload"))
	assert.Equal(t, SandboxFsTimeout, sandboxFsCommandTimeout("stat"))
}
