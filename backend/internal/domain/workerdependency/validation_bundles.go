package workerdependency

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func validateBundles(document Document, bundles []RuntimeBundle) error {
	seen := make(map[string]struct{}, len(bundles))
	domainIDs := make(map[int64]struct{}, len(bundles))
	documentIDs := make(map[string]struct{}, len(bundles))
	modelFields := modelManagedEnvironmentFields(document)
	for _, bundle := range bundles {
		if err := validatePin(
			document,
			bundle.Pin,
			resource.KindEnvironmentBundle,
		); err != nil {
			return err
		}
		key := referenceKey(bundle.Pin)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate EnvironmentBundle dependency")
		}
		seen[key] = struct{}{}
		if _, exists := domainIDs[bundle.Pin.DomainID]; exists {
			return fmt.Errorf(
				"duplicate EnvironmentBundle domain id %d",
				bundle.Pin.DomainID,
			)
		}
		domainIDs[bundle.Pin.DomainID] = struct{}{}
		switch bundle.Kind {
		case envbundle.KindRuntime, envbundle.KindShared, envbundle.KindConfig:
		default:
			return fmt.Errorf(
				"materialized EnvironmentBundle kind %q is invalid",
				bundle.Kind,
			)
		}
		if err := validateRuntimeValues(bundle.Values); err != nil {
			return err
		}
		for _, value := range bundle.Values {
			if isSensitiveFieldName(value.Name) {
				return fmt.Errorf(
					"EnvironmentBundle field %q must remain a Secret reference",
					value.Name,
				)
			}
			if containsRawSecretText(value.Value) {
				return fmt.Errorf(
					"EnvironmentBundle field %q contains raw secret-like data",
					value.Name,
				)
			}
			if _, exists := modelFields[value.Name]; exists {
				return fmt.Errorf(
					"EnvironmentBundle field %q is managed by a model resource",
					value.Name,
				)
			}
		}
		digest, err := DigestRuntimeValues(bundle.Values)
		if err != nil {
			return err
		}
		if digest != bundle.ContentDigest {
			return fmt.Errorf("EnvironmentBundle content digest does not match values")
		}
		if err := validateConfigDocument(bundle); err != nil {
			return err
		}
		if bundle.ConfigDocument != nil {
			id := bundle.ConfigDocument.ID
			if _, exists := documentIDs[id]; exists {
				return fmt.Errorf("duplicate config document id %q", id)
			}
			documentIDs[id] = struct{}{}
		}
	}
	return nil
}

func validateRuntimeValues(values []RuntimeValue) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if err := workerspec.ValidateConfigField(value.Name); err != nil {
			return fmt.Errorf("EnvironmentBundle value name: %w", err)
		}
		if _, exists := seen[value.Name]; exists {
			return fmt.Errorf("duplicate EnvironmentBundle value %q", value.Name)
		}
		seen[value.Name] = struct{}{}
	}
	return nil
}

func validateConfigDocument(bundle RuntimeBundle) error {
	if bundle.Kind != envbundle.KindConfig {
		if bundle.ConfigDocument != nil {
			return fmt.Errorf("non-config EnvironmentBundle has config document metadata")
		}
		return nil
	}
	if bundle.ConfigDocument == nil {
		return fmt.Errorf("config EnvironmentBundle requires document metadata")
	}
	document := *bundle.ConfigDocument
	if err := requireNormalized("config document id", document.ID); err != nil {
		return err
	}
	if document.Format != "json" {
		return fmt.Errorf("config document format must be json")
	}
	if err := requireNormalized("config document target path", document.TargetPath); err != nil {
		return err
	}
	var raw string
	for _, value := range bundle.Values {
		if value.Name == envbundle.ConfigJSONDataKey {
			raw = value.Value
			break
		}
	}
	if raw == "" {
		return fmt.Errorf(
			"config EnvironmentBundle requires %q",
			envbundle.ConfigJSONDataKey,
		)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil || decoded == nil {
		return fmt.Errorf("config EnvironmentBundle requires a JSON object")
	}
	return nil
}
