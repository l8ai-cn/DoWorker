package orchestrationresource

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const maxManifestRawMessageBytes = 1 << 20

var ErrServerManagedField = errors.New("server-managed field")

type Manifest struct {
	TypeMeta
	Metadata Metadata        `json:"metadata" yaml:"metadata"`
	Spec     json.RawMessage `json:"spec" yaml:"spec"`
	Status   json.RawMessage `json:"status,omitempty" yaml:"status,omitempty"`
}

func (manifest Manifest) ValidateSubmission() error {
	if manifest.Metadata.UID != "" {
		return fmt.Errorf("metadata.uid: %w", ErrServerManagedField)
	}
	if manifest.Metadata.ResourceVersion != "" {
		return fmt.Errorf("metadata.resourceVersion: %w", ErrServerManagedField)
	}
	if manifest.Metadata.Generation != 0 {
		return fmt.Errorf("metadata.generation: %w", ErrServerManagedField)
	}
	if len(manifest.Status) != 0 {
		return fmt.Errorf("status: %w", ErrServerManagedField)
	}
	return manifest.validateBase()
}

func (manifest Manifest) ValidateStored() error {
	if err := manifest.validateBase(); err != nil {
		return err
	}
	if len(manifest.Status) == 0 {
		return nil
	}
	return validateJSONObject("status", manifest.Status, true)
}

func (manifest Manifest) validateBase() error {
	if err := manifest.TypeMeta.Validate(); err != nil {
		return err
	}
	if err := manifest.Metadata.Validate(); err != nil {
		return err
	}
	return validateJSONObject("spec", manifest.Spec, false)
}

func validateJSONObject(field string, raw json.RawMessage, allowEmpty bool) error {
	if len(raw) > maxManifestRawMessageBytes {
		return fmt.Errorf("%s exceeds %d bytes", field, maxManifestRawMessageBytes)
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return fmt.Errorf("%s must be a JSON object", field)
	}

	var object map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&object); err != nil {
		return fmt.Errorf("%s must be a JSON object: %w", field, err)
	}
	if object == nil {
		return fmt.Errorf("%s must be a JSON object", field)
	}

	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("%s must contain exactly one JSON object", field)
		}
		return fmt.Errorf("%s contains invalid JSON: %w", field, err)
	}
	if !allowEmpty && len(object) == 0 {
		return fmt.Errorf("%s must not be an empty object", field)
	}
	return nil
}
