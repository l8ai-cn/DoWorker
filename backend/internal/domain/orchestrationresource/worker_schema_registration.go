package orchestrationresource

import "fmt"

const (
	KindModelBinding      = "ModelBinding"
	KindRepository        = "Repository"
	KindSkill             = "Skill"
	KindKnowledgeBase     = "KnowledgeBase"
	KindEnvironmentBundle = "EnvironmentBundle"
	KindComputeTarget     = "ComputeTarget"
	KindResourceProfile   = "ResourceProfile"
	KindToolBinding       = "ToolBinding"
	KindWorkerTemplate    = "WorkerTemplate"
)

func RegisterWorkerSchemas(registry *Registry) error {
	if registry == nil {
		return fmt.Errorf("worker schema registry must not be nil")
	}
	registrations := []struct {
		kind   string
		schema Schema
	}{
		{
			KindModelBinding,
			positiveIDBindingSchema("resourceId", func(spec *ModelBindingSpec) int64 {
				return spec.ResourceID
			}),
		},
		{
			KindRepository,
			positiveIDBindingSchema("repositoryId", func(spec *RepositoryBindingSpec) int64 {
				return spec.RepositoryID
			}),
		},
		{
			KindSkill,
			positiveIDBindingSchema("skillId", func(spec *SkillBindingSpec) int64 {
				return spec.SkillID
			}),
		},
		{
			KindKnowledgeBase,
			positiveIDBindingSchema("knowledgeBaseId", func(spec *KnowledgeBaseBindingSpec) int64 {
				return spec.KnowledgeBaseID
			}),
		},
		{
			KindEnvironmentBundle,
			positiveIDBindingSchema("environmentBundleId", func(spec *EnvironmentBundleBindingSpec) int64 {
				return spec.EnvironmentBundleID
			}),
		},
		{
			KindComputeTarget,
			positiveIDBindingSchema("computeTargetId", func(spec *ComputeTargetBindingSpec) int64 {
				return spec.ComputeTargetID
			}),
		},
		{
			KindResourceProfile,
			positiveIDBindingSchema("resourceProfileId", func(spec *ResourceProfileBindingSpec) int64 {
				return spec.ResourceProfileID
			}),
		},
		{KindToolBinding, toolBindingSchema()},
		{KindWorkerTemplate, workerTemplateSchema()},
	}
	for _, registration := range registrations {
		if err := registry.Register(TypeMeta{
			APIVersion: APIVersionV1Alpha1,
			Kind:       registration.kind,
		}, registration.schema); err != nil {
			return err
		}
	}
	return nil
}
