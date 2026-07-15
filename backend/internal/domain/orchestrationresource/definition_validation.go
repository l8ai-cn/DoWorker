package orchestrationresource

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func validateDefinitionText(
	field string,
	value string,
	maxRunes int,
	required bool,
) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s contains invalid UTF-8", field)
	}
	if required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	if utf8.RuneCountInString(value) > maxRunes {
		return fmt.Errorf("%s exceeds %d runes", field, maxRunes)
	}
	for _, character := range value {
		if unicode.Is(unicode.Bidi_Control, character) ||
			(unicode.IsControl(character) &&
				character != '\n' &&
				character != '\r' &&
				character != '\t') {
			return fmt.Errorf("%s contains forbidden control characters", field)
		}
	}
	return nil
}

func validateDefinitionReference(
	metadata Metadata,
	field string,
	kind string,
	reference Reference,
) error {
	return validateWorkerReference(metadata, field, kind, reference)
}

func validateDefinitionStringMap(
	field string,
	values map[string]string,
	maxEntries int,
	maxValueRunes int,
) error {
	if values == nil {
		return fmt.Errorf("%s must be an object", field)
	}
	if len(values) > maxEntries {
		return fmt.Errorf("%s exceeds %d entries", field, maxEntries)
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := slugkit.Validate(key); err != nil {
			return fmt.Errorf("%s key %q: %w", field, summarizeValue(key), err)
		}
		if err := validateDefinitionText(
			field+"["+key+"]",
			values[key],
			maxValueRunes,
			false,
		); err != nil {
			return err
		}
	}
	return nil
}
