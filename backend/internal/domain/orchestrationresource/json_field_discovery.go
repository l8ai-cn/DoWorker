package orchestrationresource

import (
	"reflect"
	"strings"
	"unicode"
)

type jsonFieldCandidate struct {
	fieldType reflect.Type
	depth     int
	tagged    bool
}

func discoverJSONFields(targetType reflect.Type) map[string]reflect.Type {
	candidates := make(map[string][]jsonFieldCandidate)
	collectJSONFields(targetType, 0, make(map[reflect.Type]bool), candidates)
	fields := make(map[string]reflect.Type, len(candidates))
	for name, namedCandidates := range candidates {
		if fieldType, ok := dominantJSONField(namedCandidates); ok {
			fields[name] = fieldType
		}
	}
	return fields
}

func collectJSONFields(
	targetType reflect.Type,
	depth int,
	visiting map[reflect.Type]bool,
	candidates map[string][]jsonFieldCandidate,
) {
	if visiting[targetType] {
		return
	}
	visiting[targetType] = true
	defer delete(visiting, targetType)

	for index := 0; index < targetType.NumField(); index++ {
		field := targetType.Field(index)
		embeddedType := field.Type
		if embeddedType.Kind() == reflect.Pointer {
			embeddedType = embeddedType.Elem()
		}
		if field.Anonymous {
			if !field.IsExported() && embeddedType.Kind() != reflect.Struct {
				continue
			}
		} else if !field.IsExported() {
			continue
		}

		tagName, tagged, ignored := parseJSONFieldTag(field.Tag.Get("json"))
		if ignored {
			continue
		}
		if field.Anonymous && tagName == "" && embeddedType.Kind() == reflect.Struct {
			collectJSONFields(embeddedType, depth+1, visiting, candidates)
			continue
		}

		name := tagName
		if name == "" {
			name = field.Name
		}
		candidate := jsonFieldCandidate{fieldType: field.Type, depth: depth, tagged: tagged}
		candidates[name] = append(candidates[name], candidate)
	}
}

func parseJSONFieldTag(tag string) (name string, tagged bool, ignored bool) {
	if tag == "-" {
		return "", false, true
	}
	name, _, _ = strings.Cut(tag, ",")
	if !isValidJSONTagName(name) {
		name = ""
	}
	return name, name != "", false
}

func isValidJSONTagName(name string) bool {
	if name == "" {
		return false
	}
	for _, char := range name {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", char):
		case !unicode.IsLetter(char) && !unicode.IsDigit(char):
			return false
		}
	}
	return true
}

func dominantJSONField(candidates []jsonFieldCandidate) (reflect.Type, bool) {
	minDepth := candidates[0].depth
	for _, candidate := range candidates[1:] {
		if candidate.depth < minDepth {
			minDepth = candidate.depth
		}
	}

	var tagged, untagged []jsonFieldCandidate
	for index := range candidates {
		candidate := candidates[index]
		if candidate.depth != minDepth {
			continue
		}
		if candidate.tagged {
			tagged = append(tagged, candidate)
		} else {
			untagged = append(untagged, candidate)
		}
	}
	if len(tagged) == 1 {
		return tagged[0].fieldType, true
	}
	if len(tagged) > 1 || len(untagged) != 1 {
		return nil, false
	}
	return untagged[0].fieldType, true
}
