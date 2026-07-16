package orchestrationresource

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const (
	maxYAMLDepth = 64
	maxYAMLNodes = 10_000
)

type yamlNodeFrame struct {
	node           *yaml.Node
	containerDepth int
}

func validateYAMLTree(document *yaml.Node) error {
	stack := []yamlNodeFrame{{node: document}}
	nodeCount := 0
	for len(stack) > 0 {
		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if frame.node == nil {
			return fmt.Errorf("YAML manifest contains a nil node")
		}

		nodeCount++
		if frame.node.Kind == yaml.AliasNode {
			return fmt.Errorf("YAML aliases are not allowed")
		}
		if frame.node.Anchor != "" {
			return fmt.Errorf("YAML anchors are not allowed")
		}
		remainingNodes := maxYAMLNodes - nodeCount
		if len(stack) > remainingNodes ||
			len(frame.node.Content) > remainingNodes-len(stack) {
			return fmt.Errorf("YAML manifest exceeds maximum %d nodes", maxYAMLNodes)
		}

		childDepth, err := validateYAMLNode(frame)
		if err != nil {
			return err
		}
		for index := len(frame.node.Content) - 1; index >= 0; index-- {
			stack = append(stack, yamlNodeFrame{
				node:           frame.node.Content[index],
				containerDepth: childDepth,
			})
		}
	}
	return nil
}

func validateYAMLNode(frame yamlNodeFrame) (int, error) {
	node := frame.node
	switch node.Kind {
	case yaml.DocumentNode:
		if node.Tag != "" || len(node.Content) != 1 {
			return 0, fmt.Errorf("YAML document must contain exactly one root node")
		}
		return frame.containerDepth, nil
	case yaml.MappingNode:
		if node.ShortTag() != "!!map" {
			return 0, fmt.Errorf("unsupported YAML mapping tag %s", summarizeValue(node.ShortTag()))
		}
		if len(node.Content)%2 != 0 {
			return 0, fmt.Errorf("YAML mapping content must contain an even number of nodes")
		}
		if err := validateYAMLMappingKeys(node); err != nil {
			return 0, err
		}
		return nextYAMLContainerDepth(frame.containerDepth)
	case yaml.SequenceNode:
		if node.ShortTag() != "!!seq" {
			return 0, fmt.Errorf("unsupported YAML sequence tag %s", summarizeValue(node.ShortTag()))
		}
		return nextYAMLContainerDepth(frame.containerDepth)
	case yaml.ScalarNode:
		return frame.containerDepth, validateYAMLScalar(node)
	default:
		return 0, fmt.Errorf("unsupported YAML node kind %d", node.Kind)
	}
}

func nextYAMLContainerDepth(current int) (int, error) {
	next := current + 1
	if next > maxYAMLDepth {
		return 0, fmt.Errorf("YAML manifest exceeds maximum depth %d", maxYAMLDepth)
	}
	return next, nil
}

func validateYAMLMappingKeys(mapping *yaml.Node) error {
	keys := make(map[string]struct{}, len(mapping.Content)/2)
	for index := 0; index < len(mapping.Content); index += 2 {
		key := mapping.Content[index]
		if key.Kind != yaml.ScalarNode {
			return fmt.Errorf("YAML mapping keys must be strings")
		}
		kind, err := classifyYAMLScalar(key)
		if err != nil {
			return err
		}
		if kind != yamlStringScalar {
			return fmt.Errorf("YAML mapping keys must be strings")
		}
		if _, exists := keys[key.Value]; exists {
			return ErrDuplicateYAMLKey
		}
		keys[key.Value] = struct{}{}
	}
	return nil
}

func yamlNodeToJSONValue(node *yaml.Node) (any, error) {
	switch node.Kind {
	case yaml.DocumentNode:
		return yamlNodeToJSONValue(node.Content[0])
	case yaml.MappingNode:
		value := make(map[string]any, len(node.Content)/2)
		for index := 0; index < len(node.Content); index += 2 {
			child, err := yamlNodeToJSONValue(node.Content[index+1])
			if err != nil {
				return nil, err
			}
			value[node.Content[index].Value] = child
		}
		return value, nil
	case yaml.SequenceNode:
		value := make([]any, len(node.Content))
		for index, childNode := range node.Content {
			child, err := yamlNodeToJSONValue(childNode)
			if err != nil {
				return nil, err
			}
			value[index] = child
		}
		return value, nil
	case yaml.ScalarNode:
		return yamlScalarToJSONValue(node)
	default:
		return nil, fmt.Errorf("unsupported YAML node kind %d", node.Kind)
	}
}
