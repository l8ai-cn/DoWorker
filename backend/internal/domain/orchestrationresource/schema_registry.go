package orchestrationresource

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"sync"
)

var (
	ErrDuplicateSchema = errors.New("orchestration resource schema already registered")
	ErrUnknownSchema   = errors.New("unknown orchestration resource schema")
)

type Schema struct {
	NewSpec  func() any
	Validate func(Metadata, any) error
}

type Registry struct {
	mu      sync.RWMutex
	schemas map[TypeMeta]registeredSchema
}

func NewRegistry() *Registry {
	return &Registry{schemas: make(map[TypeMeta]registeredSchema)}
}

func (r *Registry) Has(meta TypeMeta) bool {
	_, exists := r.lookupSchema(meta)
	return exists
}

func (r *Registry) DecodeAndValidate(manifest Manifest) (any, error) {
	if err := manifest.ValidateStored(); err != nil {
		return nil, fmt.Errorf("validate stored manifest: %w", err)
	}
	schema, exists := r.lookupSchema(manifest.TypeMeta)
	if !exists {
		return nil, fmt.Errorf(
			"%w: apiVersion=%s kind=%s",
			ErrUnknownSchema,
			manifest.APIVersion,
			manifest.Kind,
		)
	}

	target := schema.newSpec()
	if err := validateJSONStructure(manifest.Spec); err != nil {
		return nil, fmt.Errorf("validate spec JSON structure: %w", err)
	}
	if err := validateTypedJSONShape(manifest.Spec, schema.rootType); err != nil {
		return nil, fmt.Errorf("validate spec JSON fields: %w", err)
	}
	if err := decodeTypedJSON(manifest.Spec, target); err != nil {
		return nil, fmt.Errorf("decode typed spec: %w", err)
	}
	if err := schema.validate(metadataForValidation(manifest.Metadata), target); err != nil {
		return nil, fmt.Errorf("validate typed spec: %w", err)
	}
	return target, nil
}

func (r *Registry) lookupSchema(meta TypeMeta) (registeredSchema, bool) {
	r.mu.RLock()
	schema, exists := r.schemas[meta]
	r.mu.RUnlock()
	return schema, exists
}

func metadataForValidation(metadata Metadata) Metadata {
	metadata.Labels = maps.Clone(metadata.Labels)
	return metadata
}

func decodeTypedJSON(source []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return sanitizeTypedJSONDecodeError(err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return sanitizeTypedJSONDecodeError(err)
	}
	return nil
}

func validateTypedJSONShape(source []byte, targetType reflect.Type) error {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("decode structured JSON: %w", err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return err
	}
	return validateJSONValueShape(value, targetType, "spec", 1)
}
