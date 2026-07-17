package orchestrationresource

import (
	"reflect"
	"sync"
)

var jsonFieldCache sync.Map

func jsonFieldsForType(targetType reflect.Type) map[string]reflect.Type {
	if cached, exists := jsonFieldCache.Load(targetType); exists {
		return cached.(map[string]reflect.Type)
	}
	discovered := discoverJSONFields(targetType)
	cached, _ := jsonFieldCache.LoadOrStore(targetType, discovered)
	return cached.(map[string]reflect.Type)
}
