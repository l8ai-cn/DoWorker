package orchestrationresource

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var (
	errYAMLOutputTooLarge = fmt.Errorf(
		"YAML output exceeds maximum %d bytes",
		maxYAMLManifestBytes,
	)
	errYAMLLineTooLong = fmt.Errorf(
		"YAML output line exceeds maximum %d bytes",
		maxYAMLLineBytes,
	)
)

type limitedYAMLBuffer struct {
	buffer    bytes.Buffer
	limit     int
	lineBytes int
	writeErr  error
}

func newLimitedYAMLBuffer(limit int) *limitedYAMLBuffer {
	return &limitedYAMLBuffer{limit: limit}
}

func (buffer *limitedYAMLBuffer) Write(value []byte) (int, error) {
	if buffer.writeErr != nil {
		return 0, buffer.writeErr
	}
	if len(value) > buffer.limit-buffer.buffer.Len() {
		return buffer.rejectWrite(errYAMLOutputTooLarge)
	}
	lineBytes, err := nextYAMLLineBytes(buffer.lineBytes, value)
	if err != nil {
		return buffer.rejectWrite(err)
	}
	written, err := buffer.buffer.Write(value)
	if err == nil {
		buffer.lineBytes = lineBytes
	}
	return written, err
}

func (buffer *limitedYAMLBuffer) Bytes() []byte {
	return buffer.buffer.Bytes()
}

func (buffer *limitedYAMLBuffer) rejectWrite(err error) (int, error) {
	buffer.writeErr = err
	return 0, err
}

func nextYAMLLineBytes(current int, value []byte) (int, error) {
	lineBytes := current
	for index := 0; index < len(value); index++ {
		switch value[index] {
		case '\n':
			lineBytes = 0
		case '\r':
			lineBytes = 0
			if index+1 < len(value) && value[index+1] == '\n' {
				index++
			}
		default:
			lineBytes++
			if lineBytes > maxYAMLLineBytes {
				return 0, errYAMLLineTooLong
			}
		}
	}
	return lineBytes, nil
}

func countYAMLNodesFromJSON(source []byte) (int, error) {
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.UseNumber()
	nodeCount := 1
	for {
		current, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return nodeCount, nil
		}
		if err != nil {
			return 0, fmt.Errorf(
				"count normalized JSON nodes: %s",
				summarizeValue(err.Error()),
			)
		}
		if delimiter, ok := current.(json.Delim); ok &&
			(delimiter == '}' || delimiter == ']') {
			continue
		}
		nodeCount++
		if nodeCount > maxYAMLNodes {
			return 0, yamlNodeCountError()
		}
	}
}

func yamlNodeCountError() error {
	return fmt.Errorf("encoded YAML manifest exceeds maximum %d nodes", maxYAMLNodes)
}
