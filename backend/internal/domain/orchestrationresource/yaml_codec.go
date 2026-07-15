package orchestrationresource

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func DecodeYAMLSubmission(source []byte) (Manifest, error) {
	if err := preflightYAML(source); err != nil {
		return Manifest{}, err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(source))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		if errors.Is(err, io.EOF) {
			return Manifest{}, fmt.Errorf("YAML manifest must contain exactly one document")
		}
		return Manifest{}, fmt.Errorf("decode YAML manifest: %s", summarizeValue(err.Error()))
	}

	var trailing yaml.Node
	err := decoder.Decode(&trailing)
	switch {
	case err == nil:
		return Manifest{}, fmt.Errorf("YAML manifest must contain exactly one document")
	case !errors.Is(err, io.EOF):
		return Manifest{}, fmt.Errorf("decode trailing YAML document: %s", summarizeValue(err.Error()))
	}

	if err := validateYAMLTree(&document); err != nil {
		return Manifest{}, err
	}
	value, err := yamlNodeToJSONValue(&document)
	if err != nil {
		return Manifest{}, fmt.Errorf("convert YAML manifest: %w", err)
	}
	normalized, err := json.Marshal(value)
	if err != nil {
		return Manifest{}, fmt.Errorf("normalize YAML manifest as JSON: %w", err)
	}
	return DecodeJSONSubmission(normalized)
}

func EncodeYAML(resource Manifest) ([]byte, error) {
	normalized, err := EncodeJSON(resource)
	if err != nil {
		return nil, err
	}
	if _, err := countYAMLNodesFromJSON(normalized); err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(normalized))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("decode normalized JSON manifest: %w", err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return nil, err
	}

	root, err := jsonValueToYAMLNode(value)
	if err != nil {
		return nil, err
	}
	document := yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{root},
	}

	buffer := newLimitedYAMLBuffer(maxYAMLManifestBytes)
	encoder := yaml.NewEncoder(buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(&document); err != nil {
		if buffer.writeErr != nil {
			return nil, fmt.Errorf("encode YAML manifest: %w", buffer.writeErr)
		}
		return nil, fmt.Errorf("encode YAML manifest: %s", summarizeValue(err.Error()))
	}
	if err := encoder.Close(); err != nil {
		if buffer.writeErr != nil {
			return nil, fmt.Errorf("close YAML encoder: %w", buffer.writeErr)
		}
		return nil, fmt.Errorf("close YAML encoder: %s", summarizeValue(err.Error()))
	}
	return buffer.Bytes(), nil
}

func jsonValueToYAMLNode(value any) (*yaml.Node, error) {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		for _, key := range keys {
			child, err := jsonValueToYAMLNode(typed[key])
			if err != nil {
				return nil, err
			}
			node.Content = append(node.Content,
				newYAMLStringNode(key),
				child,
			)
		}
		return node, nil
	case []any:
		node := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, item := range typed {
			child, err := jsonValueToYAMLNode(item)
			if err != nil {
				return nil, err
			}
			node.Content = append(node.Content, child)
		}
		return node, nil
	case string:
		return newYAMLStringNode(typed), nil
	case bool:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!bool",
			Value: strconv.FormatBool(typed),
		}, nil
	case nil:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}, nil
	case json.Number:
		if !jsonNumberPattern.MatchString(typed.String()) {
			return nil, fmt.Errorf("invalid normalized JSON number %s", summarizeValue(typed.String()))
		}
		tag := "!!int"
		if strings.ContainsAny(typed.String(), ".eE") {
			tag = "!!float"
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: typed.String()}, nil
	default:
		return nil, fmt.Errorf("unsupported normalized JSON value type %T", value)
	}
}
