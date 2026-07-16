package orchestrationresource

import (
	"fmt"
	"reflect"
)

type registeredSchema struct {
	rootType reflect.Type
	validate func(Metadata, any) error
}

func (r *Registry) Register(meta TypeMeta, schema Schema) error {
	if err := meta.Validate(); err != nil {
		return fmt.Errorf("validate schema type: %w", err)
	}
	frozen, err := freezeSchema(schema)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.schemas == nil {
		r.schemas = make(map[TypeMeta]registeredSchema)
	}
	if _, exists := r.schemas[meta]; exists {
		return fmt.Errorf("%w: apiVersion=%s kind=%s", ErrDuplicateSchema, meta.APIVersion, meta.Kind)
	}
	r.schemas[meta] = frozen
	return nil
}

func freezeSchema(schema Schema) (registeredSchema, error) {
	if schema.NewSpec == nil {
		return registeredSchema{}, fmt.Errorf("schema NewSpec must not be nil")
	}
	if schema.Validate == nil {
		return registeredSchema{}, fmt.Errorf("schema Validate must not be nil")
	}

	prototype := schema.NewSpec()
	rootType := reflect.TypeOf(prototype)
	rootValue := reflect.ValueOf(prototype)
	if prototype == nil ||
		rootType.Kind() != reflect.Pointer ||
		rootValue.IsNil() ||
		rootType.Elem().Kind() != reflect.Struct {
		return registeredSchema{}, fmt.Errorf("schema NewSpec must return a non-nil pointer to struct")
	}
	if rootType != reflect.PointerTo(rootType.Elem()) {
		return registeredSchema{}, fmt.Errorf("schema NewSpec must return an ordinary pointer to struct")
	}
	if rootType.Implements(jsonUnmarshalerType) ||
		rootType.Elem().Implements(jsonUnmarshalerType) {
		return registeredSchema{}, fmt.Errorf("schema root type must not implement json.Unmarshaler")
	}
	return registeredSchema{rootType: rootType, validate: schema.Validate}, nil
}

func (schema registeredSchema) newSpec() any {
	return reflect.New(schema.rootType.Elem()).Interface()
}
