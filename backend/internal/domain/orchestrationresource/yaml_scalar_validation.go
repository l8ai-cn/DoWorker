package orchestrationresource

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	jsonIntegerPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
	jsonNumberPattern  = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)
)

type yamlScalarKind uint8

const (
	yamlStringScalar yamlScalarKind = iota
	yamlBoolScalar
	yamlNullScalar
	yamlIntegerScalar
	yamlFloatScalar
)

func yamlScalarToJSONValue(node *yaml.Node) (any, error) {
	kind, err := classifyYAMLScalar(node)
	if err != nil {
		return nil, err
	}
	switch kind {
	case yamlStringScalar:
		return node.Value, nil
	case yamlBoolScalar:
		return strings.EqualFold(node.Value, "true"), nil
	case yamlNullScalar:
		return nil, nil
	case yamlIntegerScalar, yamlFloatScalar:
		return json.Number(node.Value), nil
	default:
		return nil, fmt.Errorf("unsupported YAML scalar kind %d", kind)
	}
}

func validateYAMLScalar(node *yaml.Node) error {
	_, err := classifyYAMLScalar(node)
	return err
}

func classifyYAMLScalar(node *yaml.Node) (yamlScalarKind, error) {
	if node.Style&yaml.TaggedStyle != 0 {
		return classifyExplicitYAMLScalar(node)
	}
	if node.Style&(yaml.DoubleQuotedStyle|yaml.SingleQuotedStyle|
		yaml.LiteralStyle|yaml.FoldedStyle) != 0 {
		if node.ShortTag() != "!!str" {
			return 0, unsupportedYAMLScalarTag(node)
		}
		return yamlStringScalar, nil
	}
	return classifyPlainYAMLScalar(node)
}

func classifyExplicitYAMLScalar(node *yaml.Node) (yamlScalarKind, error) {
	switch node.ShortTag() {
	case "!!str":
		return yamlStringScalar, nil
	case "!!bool":
		if !isYAMLCoreBool(node.Value) {
			return 0, fmt.Errorf("YAML boolean must use core syntax: %s", summarizeValue(node.Value))
		}
		return yamlBoolScalar, nil
	case "!!null":
		if !isYAMLCoreNull(node.Value) {
			return 0, fmt.Errorf("YAML null must use core null syntax: %s", summarizeValue(node.Value))
		}
		return yamlNullScalar, nil
	case "!!int":
		if !jsonIntegerPattern.MatchString(node.Value) {
			return 0, fmt.Errorf("YAML integer must use JSON integer syntax: %s", summarizeValue(node.Value))
		}
		return yamlIntegerScalar, nil
	case "!!float":
		if !jsonNumberPattern.MatchString(node.Value) || !strings.ContainsAny(node.Value, ".eE") {
			return 0, fmt.Errorf("YAML float must use JSON float syntax: %s", summarizeValue(node.Value))
		}
		return yamlFloatScalar, nil
	case "!!merge":
		return 0, fmt.Errorf("YAML merge tags are not allowed")
	default:
		return 0, unsupportedYAMLScalarTag(node)
	}
}

func classifyPlainYAMLScalar(node *yaml.Node) (yamlScalarKind, error) {
	switch {
	case isYAMLCoreBool(node.Value):
		return yamlBoolScalar, nil
	case isYAMLCoreNull(node.Value):
		return yamlNullScalar, nil
	case jsonIntegerPattern.MatchString(node.Value):
		return yamlIntegerScalar, nil
	case jsonNumberPattern.MatchString(node.Value):
		return yamlFloatScalar, nil
	}

	switch node.ShortTag() {
	case "!!str":
		return yamlStringScalar, nil
	case "!!merge":
		return 0, fmt.Errorf("YAML merge keys are not allowed")
	case "!!timestamp":
		return 0, fmt.Errorf("YAML timestamps must be quoted strings")
	case "!!binary":
		return 0, fmt.Errorf("YAML binary scalars are not allowed")
	case "!!int", "!!float":
		return 0, fmt.Errorf("YAML numeric scalar must use JSON syntax: %s", summarizeValue(node.Value))
	case "!!bool":
		return 0, fmt.Errorf("YAML boolean must use core syntax: %s", summarizeValue(node.Value))
	case "!!null":
		return 0, fmt.Errorf("YAML null must use core null syntax: %s", summarizeValue(node.Value))
	default:
		return 0, unsupportedYAMLScalarTag(node)
	}
}

func unsupportedYAMLScalarTag(node *yaml.Node) error {
	return fmt.Errorf("unsupported YAML scalar tag %s", summarizeValue(node.Tag))
}

func isYAMLCoreBool(value string) bool {
	return strings.EqualFold(value, "true") || strings.EqualFold(value, "false")
}

func isYAMLCoreNull(value string) bool {
	return value == "" || value == "~" || strings.EqualFold(value, "null")
}
