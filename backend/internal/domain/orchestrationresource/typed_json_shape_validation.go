package orchestrationresource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

var jsonUnmarshalerType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
var jsonRawMessageType = reflect.TypeOf(json.RawMessage{})

func validateJSONValueShape(value any, targetType reflect.Type, path string, depth int) error {
	if depth > maxJSONDepth+1 {
		return fmt.Errorf("%s exceeds maximum depth %d", path, maxJSONDepth)
	}
	if value == nil {
		if targetType.Kind() == reflect.Interface || targetType == jsonRawMessageType {
			return nil
		}
		return boundedTypedJSONError(
			ErrTypedJSONType,
			fmt.Sprintf("typed JSON type error: null is not allowed at path %s", path),
		)
	}
	for targetType.Kind() == reflect.Pointer {
		if targetType.Implements(jsonUnmarshalerType) {
			return nil
		}
		targetType = targetType.Elem()
	}
	if implementsJSONUnmarshaler(targetType) {
		return nil
	}
	switch targetType.Kind() {
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		return validateJSONObjectShape(object, targetType, path, depth)
	case reflect.Slice, reflect.Array:
		items, ok := value.([]any)
		if !ok {
			return nil
		}
		for index, item := range items {
			itemPath := fmt.Sprintf("%s[%d]", path, index)
			if err := validateJSONValueShape(item, targetType.Elem(), itemPath, depth+1); err != nil {
				return err
			}
		}
	case reflect.Map:
		if targetType.Key().Kind() != reflect.String {
			return fmt.Errorf(
				"%s uses unsupported schema map key type %s",
				path,
				targetType.Key(),
			)
		}
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		keys := sortedJSONKeys(object)
		for _, key := range keys {
			if err := validateJSONValueShape(
				object[key],
				targetType.Elem(),
				path+"[map value]",
				depth+1,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateJSONObjectShape(
	object map[string]any,
	targetType reflect.Type,
	path string,
	depth int,
) error {
	fields := jsonFieldsForType(targetType)
	keys := sortedJSONKeys(object)
	for _, key := range keys {
		if _, exists := fields[key]; !exists {
			return boundedTypedJSONError(
				ErrTypedJSONUnknownField,
				fmt.Sprintf("typed JSON unknown field at path %s", path),
			)
		}
	}
	for _, key := range keys {
		fieldType := fields[key]
		if err := validateJSONValueShape(object[key], fieldType, path+"."+key, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func implementsJSONUnmarshaler(targetType reflect.Type) bool {
	return targetType.Implements(jsonUnmarshalerType) ||
		reflect.PointerTo(targetType).Implements(jsonUnmarshalerType)
}

func sortedJSONKeys(object map[string]any) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
