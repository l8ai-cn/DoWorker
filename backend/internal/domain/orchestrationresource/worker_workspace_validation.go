package orchestrationresource

import (
	"fmt"
	"strings"
)

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
	return validateWorkerConfigDocumentBindings(
		metadata,
		workspace.ConfigDocumentBindings,
	)
}

func validateWorkerConfigDocumentBindings(
	metadata Metadata,
	bindings []WorkerTemplateConfigDocumentBinding,
) error {
	fields := make([]workerReferenceField, len(bindings))
	documents := make(map[string]struct{}, len(bindings))
	for index, binding := range bindings {
		if binding.DocumentID == "" ||
			strings.TrimSpace(binding.DocumentID) != binding.DocumentID {
			return fmt.Errorf(
				"workspace.configDocumentBindings[%d].documentId must be normalized",
				index,
			)
		}
		if _, exists := documents[binding.DocumentID]; exists {
			return fmt.Errorf(
				"workspace.configDocumentBindings contains duplicate document",
			)
		}
		documents[binding.DocumentID] = struct{}{}
		fields[index] = workerReferenceField{
			path: fmt.Sprintf(
				"workspace.configDocumentBindings[%d].configBundleRef",
				index,
			),
			ref: binding.ConfigBundleRef,
		}
	}
	return validateWorkerReferenceFields(
		metadata,
		"workspace.configDocumentBindings",
		KindEnvironmentBundle,
		fields,
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
