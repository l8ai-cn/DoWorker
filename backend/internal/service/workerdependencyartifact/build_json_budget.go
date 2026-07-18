package workerdependencyartifact

import (
	"encoding"
	"encoding/json"
	"reflect"
)

const maxBudgetValueDepth = 64

var (
	jsonMarshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

func consumeJSONBudget(
	value reflect.Value,
	consume func(int) bool,
	text func(...string) bool,
	count func(int) bool,
	depth int,
) bool {
	if !value.IsValid() {
		return true
	}
	if depth > maxBudgetValueDepth {
		return false
	}
	for value.Kind() == reflect.Interface {
		if value.IsNil() {
			return true
		}
		value = value.Elem()
	}
	if value.Type().Implements(jsonMarshalerType) ||
		value.Type().Implements(textMarshalerType) {
		return false
	}
	switch value.Kind() {
	case reflect.String:
		return text(value.String())
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return consume(16)
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String ||
			value.Type().Key().Implements(textMarshalerType) ||
			!count(value.Len()) {
			return false
		}
		iterator := value.MapRange()
		for iterator.Next() {
			if !text(iterator.Key().String()) ||
				!consumeJSONBudget(
					iterator.Value(),
					consume,
					text,
					count,
					depth+1,
				) {
				return false
			}
		}
		return true
	case reflect.Slice, reflect.Array:
		if value.Kind() == reflect.Slice && value.IsNil() {
			return true
		}
		if value.Kind() == reflect.Slice &&
			value.Type().Elem().Kind() == reflect.Uint8 {
			return false
		}
		if !count(value.Len()) {
			return false
		}
		for index := 0; index < value.Len(); index++ {
			if !consumeJSONBudget(
				value.Index(index),
				consume,
				text,
				count,
				depth+1,
			) {
				return false
			}
		}
		return true
	case reflect.Pointer:
		if value.IsNil() {
			return true
		}
		return consumeJSONBudget(
			value.Elem(),
			consume,
			text,
			count,
			depth+1,
		)
	default:
		return false
	}
}
