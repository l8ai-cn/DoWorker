package coordinator

import (
	"testing"

	workerruntime "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateManagedRunnerImagesRejectsUnreleasedFormalWorker(t *testing.T) {
	err := validateManagedRunnerImages(
		workerruntime.DefaultCatalog(),
		[]string{"codex-cli", "aider"},
		map[string]string{
			"codex-cli": "repo.aiedulab.cn:8443/agentsmesh/runner-codex-cli@sha256:963c99fb047c0a4fed518eb9949e805fd31329a8395526fbb1fe34d8254ebea1",
		},
	)

	require.Error(t, err)
	assert.ErrorContains(t, err, "codex-cli")
	assert.ErrorContains(t, err, "not released")
}

func TestValidateManagedRunnerImagesAllowsNonFormalWorker(t *testing.T) {
	err := validateManagedRunnerImages(
		workerruntime.DefaultCatalog(),
		[]string{"codex-cli"},
		map[string]string{
			"e2e-echo": "repo.aiedulab.cn:8443/agentsmesh/runner-e2e-echo@sha256:077eb4511113ddb80dd8e09d7b46ffe3668d6b69d1840c1cbe849e97595087fa",
		},
	)

	assert.NoError(t, err)
}
