package orchestrationresource

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func DecodeJSONSubmission(source []byte) (Manifest, error) {
	if err := validateJSONStructure(source); err != nil {
		return Manifest{}, err
	}
	if err := validateJSONEnvelope(source); err != nil {
		return Manifest{}, err
	}

	var manifest Manifest
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf(
			"decode JSON manifest: %w",
			sanitizeTypedJSONDecodeError(err),
		)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return Manifest{}, err
	}
	if err := manifest.ValidateSubmission(); err != nil {
		return Manifest{}, fmt.Errorf("validate JSON submission: %w", err)
	}
	return manifest, nil
}

func EncodeJSON(resource Manifest) ([]byte, error) {
	if err := resource.ValidateStored(); err != nil {
		return nil, fmt.Errorf("validate stored manifest: %w", err)
	}
	compact, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("encode JSON manifest: %w", err)
	}
	if err := validateJSONStructure(compact); err != nil {
		return nil, fmt.Errorf("validate encoded JSON manifest: %w", err)
	}
	if len(compact)+1 > maxManifestBytes {
		return nil, fmt.Errorf(
			"encoded JSON manifest exceeds %d bytes including trailing newline",
			maxManifestBytes,
		)
	}
	return append(compact, '\n'), nil
}

func validateJSONEnvelope(source []byte) error {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(source, &envelope); err != nil || envelope == nil {
		if err != nil {
			return fmt.Errorf("JSON manifest must be an object: %w", err)
		}
		return fmt.Errorf("JSON manifest must be an object")
	}
	if err := rejectUnknownJSONFields("JSON manifest", envelope, isManifestJSONField); err != nil {
		return err
	}

	rawMetadata, exists := envelope["metadata"]
	if !exists {
		return nil
	}
	var metadata map[string]json.RawMessage
	if err := json.Unmarshal(rawMetadata, &metadata); err != nil || metadata == nil {
		if err != nil {
			return fmt.Errorf("metadata must be a JSON object: %w", err)
		}
		return fmt.Errorf("metadata must be a JSON object")
	}
	return rejectUnknownJSONFields("metadata", metadata, isMetadataJSONField)
}

func rejectUnknownJSONFields(
	objectName string,
	object map[string]json.RawMessage,
	allowed func(string) bool,
) error {
	for key := range object {
		if !allowed(key) {
			return fmt.Errorf("%s contains %w", objectName, ErrUnknownJSONField)
		}
	}
	return nil
}

func isManifestJSONField(key string) bool {
	switch key {
	case "apiVersion", "kind", "metadata", "spec", "status":
		return true
	default:
		return false
	}
}

func isMetadataJSONField(key string) bool {
	switch key {
	case "name", "namespace", "displayName", "labels", "uid", "resourceVersion", "generation":
		return true
	default:
		return false
	}
}

func requireJSONEOF(decoder *json.Decoder) error {
	var trailing json.RawMessage
	err := decoder.Decode(&trailing)
	switch {
	case errors.Is(err, io.EOF):
		return nil
	case err == nil:
		return fmt.Errorf("trailing JSON data")
	default:
		return fmt.Errorf("decode trailing JSON data: %w", err)
	}
}
