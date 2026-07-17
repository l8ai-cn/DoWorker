package orchestrationresource

import (
	"fmt"
	"unicode/utf8"
)

const (
	maxYAMLManifestBytes = 256 << 10
	maxYAMLLineBytes     = 64 << 10
)

func preflightYAML(source []byte) error {
	if len(source) > maxYAMLManifestBytes {
		return fmt.Errorf("YAML manifest exceeds %d bytes", maxYAMLManifestBytes)
	}
	if !utf8.Valid(source) {
		return fmt.Errorf("YAML manifest must be valid UTF-8")
	}
	return validateYAMLLineLengths(source)
}

func validateYAMLLineLengths(source []byte) error {
	line := 1
	lineBytes := 0
	for index := 0; index < len(source); index++ {
		switch source[index] {
		case '\n':
			line++
			lineBytes = 0
		case '\r':
			if index+1 < len(source) && source[index+1] == '\n' {
				index++
			}
			line++
			lineBytes = 0
		default:
			lineBytes++
			if lineBytes > maxYAMLLineBytes {
				return fmt.Errorf(
					"YAML line %d exceeds maximum %d bytes",
					line,
					maxYAMLLineBytes,
				)
			}
		}
	}
	return nil
}
