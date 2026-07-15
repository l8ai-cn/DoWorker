package orchestrationresource

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

const (
	maxManifestBytes = 1 << 20
	maxJSONDepth     = 64
)

type jsonContainer struct {
	kind         json.Delim
	keys         map[string]struct{}
	expectingKey bool
}

func validateJSONStructure(source []byte) error {
	if len(source) > maxManifestBytes {
		return fmt.Errorf("JSON manifest exceeds %d bytes", maxManifestBytes)
	}
	if !utf8.Valid(source) {
		return fmt.Errorf("JSON manifest must be valid UTF-8")
	}
	return scanJSONManifest(source)
}

func scanJSONManifest(source []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.UseNumber()

	var containers []jsonContainer
	rootComplete := false
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			if !rootComplete {
				return fmt.Errorf("JSON manifest must contain exactly one document")
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("scan JSON manifest: %w", err)
		}
		if rootComplete {
			return fmt.Errorf("trailing JSON data")
		}

		delim, isDelim := token.(json.Delim)
		if isDelim && (delim == '}' || delim == ']') {
			if len(containers) == 0 {
				return fmt.Errorf("scan JSON manifest: unexpected closing delimiter %q", delim)
			}
			containers = containers[:len(containers)-1]
			if len(containers) == 0 {
				rootComplete = true
			}
			continue
		}
		consumedKey, err := consumeJSONObjectKey(containers, token)
		if err != nil {
			return err
		}
		if consumedKey {
			continue
		}
		markJSONValue(containers)
		if !isDelim {
			if len(containers) == 0 {
				rootComplete = true
			}
			continue
		}
		if len(containers) == maxJSONDepth {
			return fmt.Errorf("JSON manifest exceeds maximum depth %d", maxJSONDepth)
		}

		container := jsonContainer{kind: delim}
		if delim == '{' {
			container.keys = make(map[string]struct{})
			container.expectingKey = true
		}
		containers = append(containers, container)
	}
}

func consumeJSONObjectKey(containers []jsonContainer, token json.Token) (bool, error) {
	if len(containers) == 0 {
		return false, nil
	}
	container := &containers[len(containers)-1]
	if container.kind != '{' || !container.expectingKey {
		return false, nil
	}

	key, ok := token.(string)
	if !ok {
		return false, nil
	}
	if _, exists := container.keys[key]; exists {
		return false, ErrDuplicateJSONKey
	}
	container.keys[key] = struct{}{}
	container.expectingKey = false
	return true, nil
}

func markJSONValue(containers []jsonContainer) {
	if len(containers) == 0 {
		return
	}
	container := &containers[len(containers)-1]
	if container.kind == '{' {
		container.expectingKey = true
	}
}
