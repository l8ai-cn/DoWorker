package orchestrationresource

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

const maxWorkerOptionsRevisionRunes = 128

func workerTemplateSchema() Schema {
	return Schema{
		NewSpec: func() any { return &WorkerTemplateSpec{} },
		Validate: func(metadata Metadata, value any) error {
			return validateWorkerTemplate(metadata, value.(*WorkerTemplateSpec))
		},
	}
}

func validateWorkerTemplate(metadata Metadata, spec *WorkerTemplateSpec) error {
	if err := validateWorkerOptionsRevision(spec.OptionsRevision); err != nil {
		return err
	}
	if err := slugkit.Validate(spec.WorkerType.String()); err != nil {
		return fmt.Errorf("workerType: %w", err)
	}
	if spec.ModelRef != nil {
		if err := validateWorkerReference(
			metadata,
			"modelRef",
			KindModelBinding,
			*spec.ModelRef,
		); err != nil {
			return err
		}
	}
	if err := validateWorkerReferenceMap(
		metadata,
		"toolRefs",
		"tool role",
		KindToolBinding,
		spec.ToolRefs,
		slugkit.Validate,
		false,
	); err != nil {
		return err
	}
	if err := validateWorkerRuntime(metadata, spec.Runtime); err != nil {
		return err
	}
	if err := validateWorkerTypeConfig(metadata, spec.TypeConfig); err != nil {
		return err
	}
	if err := validateWorkerWorkspace(metadata, spec.Workspace); err != nil {
		return err
	}
	return validateWorkerTemplateSemantics(*spec)
}

func validateWorkerOptionsRevision(value string) error {
	if value == "" ||
		!utf8.ValidString(value) ||
		strings.TrimSpace(value) != value ||
		utf8.RuneCountInString(value) > maxWorkerOptionsRevisionRunes {
		return fmt.Errorf(
			"optionsRevision must contain 1-%d normalized runes",
			maxWorkerOptionsRevisionRunes,
		)
	}
	for _, character := range value {
		if unicode.IsControl(character) ||
			unicode.Is(unicode.Bidi_Control, character) {
			return fmt.Errorf("optionsRevision must not contain control characters")
		}
	}
	return nil
}

func validateWorkerRuntime(
	metadata Metadata,
	runtime WorkerTemplateRuntimeSpec,
) error {
	if err := validateWorkerReference(
		metadata,
		"runtime.computeTargetRef",
		KindComputeTarget,
		runtime.ComputeTargetRef,
	); err != nil {
		return err
	}
	if runtime.ResourceProfileRef != nil {
		if err := validateWorkerReference(
			metadata,
			"runtime.resourceProfileRef",
			KindResourceProfile,
			*runtime.ResourceProfileRef,
		); err != nil {
			return err
		}
	}
	if runtime.ResourceProfileRef != nil && runtime.CustomResources != nil {
		return fmt.Errorf(
			"runtime.resourceProfileRef and customResources are mutually exclusive",
		)
	}
	return nil
}

func validateWorkerTypeConfig(
	metadata Metadata,
	config WorkerTemplateTypeConfigSpec,
) error {
	if err := validateWorkerConfigKeys(
		"config value field",
		config.Values,
	); err != nil {
		return err
	}
	return validateWorkerReferenceMap(
		metadata,
		"typeConfig.secretRefs",
		"secret config field",
		KindEnvironmentBundle,
		config.SecretRefs,
		workerspec.ValidateConfigField,
		true,
	)
}

func validateWorkerConfigKeys(field string, values map[string]any) error {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := workerspec.ValidateConfigField(key); err != nil {
			return fmt.Errorf("%s is invalid", field)
		}
	}
	return nil
}

func validateWorkerReferenceMap(
	metadata Metadata,
	field string,
	keyField string,
	expectedKind string,
	references map[string]Reference,
	validateKey func(string) error,
	allowDuplicateTargets bool,
) error {
	keys := make([]string, 0, len(references))
	for key := range references {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fields := make([]workerReferenceField, 0, len(keys))
	for _, key := range keys {
		if err := validateKey(key); err != nil {
			return fmt.Errorf("%s is invalid", keyField)
		}
		fields = append(fields, workerReferenceField{
			path: field + "[map value]",
			ref:  references[key],
		})
	}
	if allowDuplicateTargets {
		return validateWorkerReferenceFieldValues(metadata, expectedKind, fields)
	}
	return validateWorkerReferenceFields(metadata, field, expectedKind, fields)
}
