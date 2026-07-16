package orchestrationresource

import "fmt"

type ModelBindingSpec struct {
	ResourceID int64 `json:"resourceId" yaml:"resourceId"`
}

type RepositoryBindingSpec struct {
	RepositoryID int64 `json:"repositoryId" yaml:"repositoryId"`
}

type SkillBindingSpec struct {
	SkillID int64 `json:"skillId" yaml:"skillId"`
}

type KnowledgeBaseBindingSpec struct {
	KnowledgeBaseID int64 `json:"knowledgeBaseId" yaml:"knowledgeBaseId"`
}

type EnvironmentBundleBindingSpec struct {
	EnvironmentBundleID int64 `json:"environmentBundleId" yaml:"environmentBundleId"`
}

type ComputeTargetBindingSpec struct {
	ComputeTargetID int64 `json:"computeTargetId" yaml:"computeTargetId"`
}

type ResourceProfileBindingSpec struct {
	ResourceProfileID int64 `json:"resourceProfileId" yaml:"resourceProfileId"`
}

type ToolBindingSpec struct {
	ModelRef Reference `json:"modelRef" yaml:"modelRef"`
}

func positiveIDBindingSchema[T any](
	field string,
	readID func(*T) int64,
) Schema {
	return Schema{
		NewSpec: func() any { return new(T) },
		Validate: func(_ Metadata, value any) error {
			if readID(value.(*T)) <= 0 {
				return fmt.Errorf("%s must be greater than 0", field)
			}
			return nil
		},
	}
}

func toolBindingSchema() Schema {
	return Schema{
		NewSpec: func() any { return &ToolBindingSpec{} },
		Validate: func(metadata Metadata, value any) error {
			spec := value.(*ToolBindingSpec)
			return validateWorkerReference(
				metadata,
				"modelRef",
				KindModelBinding,
				spec.ModelRef,
			)
		},
	}
}
