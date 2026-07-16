package orchestrationresource

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

func newYAMLStringNode(value string) *yaml.Node {
	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
	}
	if requiresQuotedYAMLString(value) {
		node.Style = yaml.DoubleQuotedStyle
	}
	return node
}

func requiresQuotedYAMLString(value string) bool {
	if value == "" || isYAMLCoreBool(value) || isYAMLCoreNull(value) ||
		jsonNumberPattern.MatchString(value) || value == "<<" {
		return true
	}
	if strings.ContainsAny(value, "\r\n\t") ||
		strings.Contains(value, ": ") || strings.Contains(value, " #") {
		return true
	}

	first, _ := utf8.DecodeRuneInString(value)
	last, _ := utf8.DecodeLastRuneInString(value)
	if unicode.IsSpace(first) || unicode.IsSpace(last) ||
		unicode.IsControl(first) || unicode.IsControl(last) ||
		unicode.IsDigit(first) {
		return true
	}
	return strings.ContainsRune("-?:,[]{}#&*!|>'\"%@`+.", first)
}
