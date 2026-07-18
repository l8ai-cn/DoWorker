package workerspec

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type TypeFieldKind string

const (
	TypeFieldBoolean TypeFieldKind = "boolean"
	TypeFieldString  TypeFieldKind = "string"
	TypeFieldNumber  TypeFieldKind = "number"
	TypeFieldSelect  TypeFieldKind = "select"
	TypeFieldSecret  TypeFieldKind = "secret"
)

type TypeFieldSchema struct {
	Kind        TypeFieldKind
	Options     []string
	Default     any
	Required    bool
	Description string
}

type SecretRequirementGroup struct {
	ID    string
	AnyOf []string
}

type TypeSchema struct {
	Version                 uint32
	Fields                  map[string]TypeFieldSchema
	SecretRequirementGroups []SecretRequirementGroup
}

func ValidateTypeConfigAgainstSchema(config TypeConfig, schema TypeSchema) error {
	if err := validateTypeSchema(schema); err != nil {
		return err
	}
	if config.SchemaVersion != schema.Version {
		return fmt.Errorf(
			"type config schema version %d does not match worker type schema version %d",
			config.SchemaVersion,
			schema.Version,
		)
	}
	values, err := cloneJSONValues(config.Values)
	if err != nil {
		return fmt.Errorf("type config values: %w", err)
	}
	for field, value := range values {
		definition, exists := schema.Fields[field]
		if !exists {
			return fmt.Errorf("type config field %q is not declared", field)
		}
		if definition.Kind == TypeFieldSecret {
			return fmt.Errorf("type config field %q must use secret_refs", field)
		}
		if err := validateTypeFieldValue(field, value, definition); err != nil {
			return err
		}
	}
	for field := range config.SecretRefs {
		definition, exists := schema.Fields[field]
		if !exists {
			return fmt.Errorf("type config secret ref %q is not declared", field)
		}
		if definition.Kind != TypeFieldSecret {
			return fmt.Errorf("type config field %q does not accept secret_refs", field)
		}
	}
	for field, definition := range schema.Fields {
		if definition.Kind != TypeFieldSecret || !definition.Required {
			continue
		}
		if _, exists := config.SecretRefs[field]; !exists {
			return fmt.Errorf("type config secret ref %q is required", field)
		}
	}
	for _, group := range schema.SecretRequirementGroups {
		hasReference := false
		for _, field := range group.AnyOf {
			if _, exists := config.SecretRefs[field]; exists {
				hasReference = true
				break
			}
		}
		if !hasReference {
			return fmt.Errorf(
				"at least one secret ref is required for credential group %q",
				group.ID,
			)
		}
	}
	return nil
}

func validateTypeSchema(schema TypeSchema) error {
	if schema.Version == 0 {
		return fmt.Errorf("worker type schema version must be positive")
	}
	if schema.Fields == nil {
		return fmt.Errorf("worker type schema fields must be an object")
	}
	for field, definition := range schema.Fields {
		if err := ValidateConfigField(field); err != nil {
			return fmt.Errorf("worker type schema: %w", err)
		}
		switch definition.Kind {
		case TypeFieldBoolean, TypeFieldString, TypeFieldNumber, TypeFieldSecret:
			if len(definition.Options) != 0 {
				return fmt.Errorf("worker type schema field %q cannot declare options", field)
			}
		case TypeFieldSelect:
			if len(definition.Options) == 0 {
				return fmt.Errorf("worker type schema select field %q requires options", field)
			}
			if err := validateUniqueOptions(field, definition.Options); err != nil {
				return err
			}
		default:
			return fmt.Errorf("worker type schema field %q has invalid kind %q", field, definition.Kind)
		}
	}
	groupIDs := make(map[string]struct{}, len(schema.SecretRequirementGroups))
	groupTargets := make(map[string]struct{})
	for _, group := range schema.SecretRequirementGroups {
		if group.ID == "" || len(group.AnyOf) < 2 {
			return fmt.Errorf("worker type schema has invalid credential requirement group")
		}
		if _, exists := groupIDs[group.ID]; exists {
			return fmt.Errorf("worker type schema has duplicate credential requirement group %q", group.ID)
		}
		groupIDs[group.ID] = struct{}{}
		for _, field := range group.AnyOf {
			definition, exists := schema.Fields[field]
			if !exists || definition.Kind != TypeFieldSecret {
				return fmt.Errorf("credential requirement group %q references invalid secret field %q", group.ID, field)
			}
			if _, exists := groupTargets[field]; exists {
				return fmt.Errorf("credential requirement groups reuse secret field %q", field)
			}
			groupTargets[field] = struct{}{}
		}
	}
	return nil
}

func validateTypeFieldValue(field string, value any, definition TypeFieldSchema) error {
	switch definition.Kind {
	case TypeFieldBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("type config field %q must be boolean", field)
		}
	case TypeFieldString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("type config field %q must be string", field)
		}
	case TypeFieldNumber:
		number, ok := value.(json.Number)
		if !ok {
			return fmt.Errorf("type config field %q must be number", field)
		}
		if _, err := strconv.ParseFloat(number.String(), 64); err != nil {
			return fmt.Errorf("type config field %q must be number", field)
		}
	case TypeFieldSelect:
		selected, ok := value.(string)
		if !ok {
			return fmt.Errorf("type config field %q must be a select option", field)
		}
		for _, option := range definition.Options {
			if selected == option {
				return nil
			}
		}
		return fmt.Errorf("type config field %q has invalid option %q", field, selected)
	}
	return nil
}

func validateUniqueOptions(field string, options []string) error {
	seen := make(map[string]struct{}, len(options))
	for _, option := range options {
		if _, exists := seen[option]; exists {
			return fmt.Errorf("worker type schema field %q has duplicate option %q", field, option)
		}
		seen[option] = struct{}{}
	}
	return nil
}
