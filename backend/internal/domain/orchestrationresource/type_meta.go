package orchestrationresource

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const APIVersionV1Alpha1 = "agentcloud.io/v1alpha1"

var pascalCaseIdentifierPattern = regexp.MustCompile(`^[A-Z][A-Za-z0-9]{1,99}$`)

type TypeMeta struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
	Kind       string `json:"kind" yaml:"kind"`
}

func (meta TypeMeta) Validate() error {
	if meta.APIVersion != APIVersionV1Alpha1 {
		return fmt.Errorf("typeMeta.APIVersion is invalid: %s", summarizeValue(meta.APIVersion))
	}
	if !pascalCaseIdentifierPattern.MatchString(meta.Kind) {
		return fmt.Errorf("typeMeta.Kind is invalid: %s", summarizeValue(meta.Kind))
	}
	return nil
}

func summarizeValue(value string) string {
	const maxLen = 80
	sanitized := strings.ReplaceAll(strings.ReplaceAll(value, "\n", "\\n"), "\r", "\\r")
	truncated := []rune(sanitized)
	if len(truncated) > maxLen {
		sanitized = string(truncated[:maxLen])
	}
	return strconv.Quote(sanitized)
}
