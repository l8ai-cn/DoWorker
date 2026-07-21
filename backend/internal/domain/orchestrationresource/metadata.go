package orchestrationresource

import (
	"fmt"
	"sort"
	"unicode"
	"unicode/utf8"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

const (
	maxMetadataLabels   = 64
	maxDisplayNameRunes = 200
	maxServerFieldRunes = 128
)

type Metadata struct {
	Name            slugkit.Slug      `json:"name" yaml:"name"`
	Namespace       slugkit.Slug      `json:"namespace" yaml:"namespace"`
	DisplayName     string            `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Labels          map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	UID             string            `json:"uid,omitempty" yaml:"uid,omitempty"`
	ResourceVersion string            `json:"resourceVersion,omitempty" yaml:"resourceVersion,omitempty"`
	Generation      int64             `json:"generation,omitempty" yaml:"generation,omitempty"`
}

func (metadata Metadata) Validate() error {
	if err := slugkit.Validate(metadata.Name.String()); err != nil {
		return fmt.Errorf("metadata.name: %w", err)
	}
	if err := slugkit.Validate(metadata.Namespace.String()); err != nil {
		return fmt.Errorf("metadata.namespace: %w", err)
	}
	if err := validateMetadataText("metadata.displayName", metadata.DisplayName, maxDisplayNameRunes); err != nil {
		return err
	}
	if len(metadata.Labels) > maxMetadataLabels {
		return fmt.Errorf("metadata.labels exceeds %d entries", maxMetadataLabels)
	}

	labelKeys := make([]string, 0, len(metadata.Labels))
	for key := range metadata.Labels {
		labelKeys = append(labelKeys, key)
	}
	sort.Strings(labelKeys)
	for _, key := range labelKeys {
		if err := slugkit.Validate(key); err != nil {
			return fmt.Errorf("metadata.labels[%s] key: %w", summarizeValue(key), err)
		}
		value := metadata.Labels[key]
		if value == "" {
			continue
		}
		if err := slugkit.Validate(value); err != nil {
			return fmt.Errorf("metadata.labels[%s] value: %w", summarizeValue(key), err)
		}
	}

	if metadata.Generation < 0 {
		return fmt.Errorf("metadata.generation must not be negative")
	}
	if err := validateMetadataText("metadata.uid", metadata.UID, maxServerFieldRunes); err != nil {
		return err
	}
	if err := validateMetadataText(
		"metadata.resourceVersion",
		metadata.ResourceVersion,
		maxServerFieldRunes,
	); err != nil {
		return err
	}
	return nil
}

func validateMetadataText(field, value string, maxRunes int) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s contains invalid UTF-8", field)
	}
	if utf8.RuneCountInString(value) > maxRunes {
		return fmt.Errorf("%s exceeds %d runes", field, maxRunes)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Errorf("%s contains Unicode control character %U", field, r)
		}
		if unicode.Is(unicode.Bidi_Control, r) {
			return fmt.Errorf("%s contains Unicode bidi control character %U", field, r)
		}
	}
	return nil
}
