package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSandboxFsCommandTimeoutAllowsArtifactUpload(t *testing.T) {
	assert.Equal(t, SandboxFsUploadTimeout, sandboxFsCommandTimeout("upload"))
	assert.Equal(t, SandboxFsUploadTimeout, sandboxFsCommandTimeout("read_verified_bytes"))
	assert.Equal(t, SandboxFsTimeout, sandboxFsCommandTimeout("stat"))
}
