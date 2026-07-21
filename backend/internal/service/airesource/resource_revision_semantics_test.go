package airesource

import (
	"testing"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/stretchr/testify/assert"
)

func TestResourceRuntimeChangeIgnoresOrderButNotDuplicates(t *testing.T) {
	current := &domain.ModelResource{
		ModelID:      "provider/model",
		Modalities:   []domain.Modality{domain.ModalityChat, domain.ModalityMultimodal},
		Capabilities: []domain.Capability{domain.CapabilityTextGeneration, domain.CapabilityVisionInput},
	}
	reordered := &domain.ModelResource{
		ModelID:      current.ModelID,
		Modalities:   []domain.Modality{domain.ModalityMultimodal, domain.ModalityChat},
		Capabilities: []domain.Capability{domain.CapabilityVisionInput, domain.CapabilityTextGeneration},
	}
	assert.False(t, resourceRuntimeChanged(current, reordered))

	reordered.Modalities = append(reordered.Modalities, domain.ModalityChat)
	assert.True(t, resourceRuntimeChanged(current, reordered))
	reordered.Modalities = reordered.Modalities[:2]
	reordered.ModelID = "provider/new-model"
	assert.True(t, resourceRuntimeChanged(current, reordered))
}
