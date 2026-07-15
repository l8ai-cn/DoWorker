package orchestrationresource

import "fmt"

func validateWorkerWorkspace(
	metadata Metadata,
	workspace WorkerTemplateWorkspaceSpec,
) error {
	if workspace.RepositoryRef != nil {
		if err := validateWorkerReference(
			metadata,
			"workspace.repositoryRef",
			KindRepository,
			*workspace.RepositoryRef,
		); err != nil {
			return err
		}
	}
	if err := validateWorkerReferenceSlice(
		metadata,
		"workspace.skillRefs",
		KindSkill,
		workspace.SkillRefs,
	); err != nil {
		return err
	}
	if err := validateWorkerKnowledgeMounts(
		metadata,
		workspace.KnowledgeMounts,
	); err != nil {
		return err
	}
	if err := validateWorkerReferenceSlice(
		metadata,
		"workspace.environmentBundleRefs",
		KindEnvironmentBundle,
		workspace.EnvironmentBundleRefs,
	); err != nil {
		return err
	}
	return validateWorkerReferenceSlice(
		metadata,
		"workspace.configBundleRefs",
		KindEnvironmentBundle,
		workspace.ConfigBundleRefs,
	)
}

func validateWorkerReferenceSlice(
	metadata Metadata,
	field string,
	expectedKind string,
	references []Reference,
) error {
	fields := make([]workerReferenceField, len(references))
	for index, ref := range references {
		fields[index] = workerReferenceField{
			path: fmt.Sprintf("%s[%d]", field, index),
			ref:  ref,
		}
	}
	return validateWorkerReferenceFields(metadata, field, expectedKind, fields)
}

func validateWorkerKnowledgeMounts(
	metadata Metadata,
	mounts []WorkerTemplateKnowledgeMount,
) error {
	fields := make([]workerReferenceField, len(mounts))
	for index, mount := range mounts {
		fields[index] = workerReferenceField{
			path: fmt.Sprintf("workspace.knowledgeMounts[%d].ref", index),
			ref:  mount.Ref,
		}
	}
	return validateWorkerReferenceFields(
		metadata,
		"workspace.knowledgeMounts",
		KindKnowledgeBase,
		fields,
	)
}
