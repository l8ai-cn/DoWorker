package orchestrationresource

import "fmt"

type workerReferenceField struct {
	path string
	ref  Reference
}

func validateWorkerReference(
	metadata Metadata,
	path string,
	expectedKind string,
	ref Reference,
) error {
	_, err := validatedWorkerReferenceKey(metadata, path, expectedKind, ref)
	return err
}

func validateWorkerReferenceFields(
	metadata Metadata,
	collection string,
	expectedKind string,
	fields []workerReferenceField,
) error {
	seen := make(map[string]string, len(fields))
	for _, field := range fields {
		key, err := validatedWorkerReferenceKey(
			metadata,
			field.path,
			expectedKind,
			field.ref,
		)
		if err != nil {
			return err
		}
		if previous, exists := seen[key]; exists {
			return fmt.Errorf(
				"%s: duplicate reference at %s and %s",
				collection,
				previous,
				field.path,
			)
		}
		seen[key] = field.path
	}
	return nil
}

func validateWorkerReferenceFieldValues(
	metadata Metadata,
	expectedKind string,
	fields []workerReferenceField,
) error {
	for _, field := range fields {
		if _, err := validatedWorkerReferenceKey(
			metadata,
			field.path,
			expectedKind,
			field.ref,
		); err != nil {
			return err
		}
	}
	return nil
}

func validatedWorkerReferenceKey(
	metadata Metadata,
	path string,
	expectedKind string,
	ref Reference,
) (string, error) {
	namespace := metadata.Namespace.String()
	if err := ref.ValidateDraft(namespace); err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	if ref.Kind != expectedKind {
		return "", fmt.Errorf(
			"%s.kind must be %s",
			path,
			expectedKind,
		)
	}
	apiVersion := ref.APIVersion
	if apiVersion == "" {
		apiVersion = APIVersionV1Alpha1
	}
	refNamespace := ref.Namespace.String()
	if refNamespace == "" {
		refNamespace = namespace
	}
	return fmt.Sprintf(
		"%s\x00%s\x00%s\x00%s\x00%d",
		apiVersion,
		ref.Kind,
		refNamespace,
		ref.Name,
		ref.Revision,
	), nil
}
