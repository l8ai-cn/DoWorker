package orchestrationresource

import (
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type ExpertResourceSpec struct {
	WorkerTemplateRef Reference  `json:"workerTemplateRef" yaml:"workerTemplateRef"`
	PromptRef         *Reference `json:"promptRef,omitempty" yaml:"promptRef,omitempty"`
	Description       string     `json:"description" yaml:"description"`
	Category          string     `json:"category" yaml:"category"`
	ReleaseNotes      string     `json:"releaseNotes" yaml:"releaseNotes"`
}

func expertResourceSchema() Schema {
	return Schema{
		NewSpec: func() any { return &ExpertResourceSpec{} },
		Validate: func(metadata Metadata, value any) error {
			spec := value.(*ExpertResourceSpec)
			if err := validateDefinitionReference(
				metadata,
				"workerTemplateRef",
				KindWorkerTemplate,
				spec.WorkerTemplateRef,
			); err != nil {
				return err
			}
			if spec.PromptRef != nil {
				if err := validateDefinitionReference(
					metadata,
					"promptRef",
					KindPrompt,
					*spec.PromptRef,
				); err != nil {
					return err
				}
			}
			if spec.Category != "" {
				if err := slugkit.Validate(spec.Category); err != nil {
					return fmt.Errorf("category: %w", err)
				}
			}
			if err := validateDefinitionText(
				"description",
				spec.Description,
				4_000,
				false,
			); err != nil {
				return err
			}
			return validateDefinitionText(
				"releaseNotes",
				spec.ReleaseNotes,
				4_000,
				false,
			)
		},
	}
}
