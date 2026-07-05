package capability

import (
	"fmt"
	"strings"
)

var allowed = map[string]map[string]struct{}{
	"resume":      {"none": {}, "cli": {}, "acp": {}},
	"permission":  {"none": {}, "acp": {}, "notification": {}},
	"usage":       {"none": {}, "exit": {}, "live": {}},
	"interrupt":   {"true": {}, "false": {}},
	"streaming":   {"true": {}, "false": {}},
	"subagents":   {"true": {}, "false": {}},
	"model_family": {"claude": {}, "gpt": {}, "gemini": {}, "multi": {}},
}

// Validate checks a CAPABILITY axis/value pair. The control axis accepts a
// comma-separated list of snake_case tokens.
func Validate(axis, value string) error {
	axis = strings.ToLower(strings.TrimSpace(axis))
	value = strings.TrimSpace(value)
	if axis == "" {
		return fmt.Errorf("CAPABILITY: axis name required")
	}
	if value == "" {
		return fmt.Errorf("CAPABILITY %s: value required", axis)
	}
	if axis == "control" {
		for _, token := range strings.Split(value, ",") {
			token = strings.TrimSpace(token)
			if token == "" {
				return fmt.Errorf("CAPABILITY control: empty token in %q", value)
			}
			for _, r := range token {
				if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
					return fmt.Errorf("CAPABILITY control: invalid token %q", token)
				}
			}
		}
		return nil
	}
	choices, ok := allowed[axis]
	if !ok {
		return fmt.Errorf("CAPABILITY: unknown axis %q", axis)
	}
	if _, ok := choices[strings.ToLower(value)]; !ok {
		return fmt.Errorf("CAPABILITY %s: invalid value %q", axis, value)
	}
	return nil
}
