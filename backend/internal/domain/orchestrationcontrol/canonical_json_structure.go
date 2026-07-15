package orchestrationcontrol

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"
)

const (
	maxCanonicalJSONBytes = 1 << 20
	maxCanonicalJSONDepth = 64
)

type canonicalJSONContainer struct {
	kind         json.Delim
	keys         map[string]struct{}
	expectingKey bool
}

func validateCanonicalJSONStructure(source []byte) error {
	if len(source) > maxCanonicalJSONBytes || !utf8.Valid(source) {
		return errors.New("invalid canonical JSON input")
	}

	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.UseNumber()
	var containers []canonicalJSONContainer
	rootComplete := false
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			if !rootComplete {
				return errors.New("incomplete canonical JSON input")
			}
			return nil
		}
		if err != nil || rootComplete {
			return errors.New("invalid canonical JSON input")
		}

		delim, isDelim := token.(json.Delim)
		if isDelim && (delim == '}' || delim == ']') {
			if len(containers) == 0 {
				return errors.New("invalid canonical JSON input")
			}
			containers = containers[:len(containers)-1]
			if len(containers) == 0 {
				rootComplete = true
			}
			continue
		}
		consumedKey, err := consumeCanonicalJSONObjectKey(containers, token)
		if err != nil {
			return err
		}
		if consumedKey {
			continue
		}
		markCanonicalJSONValue(containers)
		if !isDelim {
			if len(containers) == 0 {
				rootComplete = true
			}
			continue
		}
		if len(containers) == maxCanonicalJSONDepth {
			return errors.New("canonical JSON input exceeds maximum depth")
		}
		container := canonicalJSONContainer{kind: delim}
		if delim == '{' {
			container.keys = make(map[string]struct{})
			container.expectingKey = true
		}
		containers = append(containers, container)
	}
}

func consumeCanonicalJSONObjectKey(
	containers []canonicalJSONContainer,
	token json.Token,
) (bool, error) {
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
		return false, errors.New("duplicate canonical JSON key")
	}
	container.keys[key] = struct{}{}
	container.expectingKey = false
	return true, nil
}

func markCanonicalJSONValue(containers []canonicalJSONContainer) {
	if len(containers) == 0 {
		return
	}
	container := &containers[len(containers)-1]
	if container.kind == '{' {
		container.expectingKey = true
	}
}
