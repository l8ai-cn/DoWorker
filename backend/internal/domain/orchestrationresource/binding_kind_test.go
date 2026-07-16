package orchestrationresource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBindingKind(t *testing.T) {
	for _, kind := range []string{
		KindModelBinding,
		KindRepository,
		KindSkill,
		KindKnowledgeBase,
		KindEnvironmentBundle,
		KindComputeTarget,
		KindResourceProfile,
		KindToolBinding,
	} {
		assert.True(t, IsBindingKind(kind), kind)
	}
	assert.False(t, IsBindingKind(KindWorkerTemplate))
	assert.False(t, IsBindingKind(KindPrompt))
}
