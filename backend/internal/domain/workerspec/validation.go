package workerspec

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

const (
	maxAliasRunes       = 100
	maxConfigFieldRunes = 128
)

var definitionHashPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func NormalizeAndValidate(spec Spec) (Spec, error) {
	normalized, err := Normalize(spec)
	if err != nil {
		return Spec{}, err
	}
	if err := Validate(normalized); err != nil {
		return Spec{}, err
	}
	return normalized, nil
}

func Validate(spec Spec) error {
	return validateSpec(spec, validateModelBinding)
}

func validateSpec(
	spec Spec,
	validateMainModelBinding func(ModelBinding) error,
) error {
	if spec.Version != VersionV1 {
		return fmt.Errorf("workerspec version %d is unsupported", spec.Version)
	}
	if err := validateMainModelBinding(spec.Runtime.ModelBinding); err != nil {
		return err
	}
	if err := validateToolModelBindings(spec.Runtime.ToolModelBindings); err != nil {
		return err
	}
	if err := validateToolModelBindings(spec.Runtime.ToolModelBindings); err != nil {
		return err
	}
	if err := validateToolModelBindings(spec.Runtime.ToolModelBindings); err != nil {
		return err
	}
	if err := validateWorkerType(spec.Runtime.WorkerType); err != nil {
		return err
	}
	if err := ValidateRuntimePlacement(spec.Runtime.Image, spec.Placement); err != nil {
		return err
	}
	if err := validateTypeConfig(spec.TypeConfig); err != nil {
		return err
	}
	if err := validateWorkspace(spec.Workspace); err != nil {
		return err
	}
	if err := validateLifecycle(spec.Lifecycle); err != nil {
		return err
	}
	if utf8.RuneCountInString(spec.Metadata.Alias) > maxAliasRunes {
		return fmt.Errorf("metadata alias exceeds %d characters", maxAliasRunes)
	}
	if spec.Metadata.SourceExpertID != nil && *spec.Metadata.SourceExpertID <= 0 {
		return fmt.Errorf("metadata source expert id must be positive")
	}
	return nil
}

func validateModelBinding(binding ModelBinding) error {
	if binding.IsEmpty() {
		return nil
	}
	switch {
	case binding.ResourceID <= 0:
		return fmt.Errorf("runtime model binding resource id must be positive")
	case binding.ResourceRevision <= 0:
		return fmt.Errorf("runtime model binding resource revision must be positive")
	case binding.ConnectionID <= 0:
		return fmt.Errorf("runtime model binding connection id must be positive")
	case binding.ConnectionRevision <= 0:
		return fmt.Errorf("runtime model binding connection revision must be positive")
	}
	if err := slugkit.Validate(binding.ProviderKey.String()); err != nil {
		return fmt.Errorf("runtime model binding provider key: %w", err)
	}
	if err := slugkit.Validate(binding.ProtocolAdapter.String()); err != nil {
		return fmt.Errorf("runtime model binding protocol adapter: %w", err)
	}
	if strings.TrimSpace(binding.ModelID) == "" {
		return fmt.Errorf("runtime model binding model id is required")
	}
	return nil
}

func validateWorkerType(workerType WorkerType) error {
	if err := slugkit.Validate(workerType.Slug.String()); err != nil {
		return fmt.Errorf("worker type slug: %w", err)
	}
	if !definitionHashPattern.MatchString(workerType.DefinitionHash) {
		return fmt.Errorf("worker type definition hash must be lowercase SHA-256 hex")
	}
	return nil
}

func validateTypeConfig(config TypeConfig) error {
	if config.SchemaVersion == 0 {
		return fmt.Errorf("type config schema version must be positive")
	}
	if config.Values == nil {
		return fmt.Errorf("type config values must be an object")
	}
	if _, err := json.Marshal(config.Values); err != nil {
		return fmt.Errorf("type config values: %w", err)
	}
	if config.SecretRefs == nil {
		return fmt.Errorf("type config secret refs must be an object")
	}
	for field := range config.Values {
		if err := ValidateConfigField(field); err != nil {
			return err
		}
		if _, secret := config.SecretRefs[field]; secret {
			return fmt.Errorf(
				"config field %q cannot appear in both values and secret refs",
				field,
			)
		}
	}
	for field, reference := range config.SecretRefs {
		if err := ValidateConfigField(field); err != nil {
			return err
		}
		if err := slugkit.Validate(reference.Kind.String()); err != nil {
			return fmt.Errorf("secret ref %q kind: %w", field, err)
		}
		if reference.ID <= 0 {
			return fmt.Errorf("secret ref %q id must be positive", field)
		}
	}
	switch config.InteractionMode {
	case InteractionModePTY, InteractionModeACP:
	default:
		return fmt.Errorf("invalid interaction mode %q", config.InteractionMode)
	}
	switch config.AutomationLevel {
	case AutomationLevelInteractive, AutomationLevelAutoEdit, AutomationLevelAutonomous:
	default:
		return fmt.Errorf("invalid automation level %q", config.AutomationLevel)
	}
	return nil
}

func ValidateConfigField(field string) error {
	if field == "" || strings.TrimSpace(field) != field ||
		utf8.RuneCountInString(field) > maxConfigFieldRunes {
		return fmt.Errorf("invalid config field %q", field)
	}
	return nil
}
