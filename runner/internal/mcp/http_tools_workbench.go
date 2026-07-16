package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

func (s *HTTPServer) SetWorkbenchArtifactPublisher(
	publisher WorkbenchArtifactPublisher,
) {
	s.artifactPublisher = publisher
}

func (s *HTTPServer) createWorkbenchPublishArtifactTool() *MCPTool {
	return &MCPTool{
		Name:        "workbench.publish_artifact",
		Description: "Publish an existing workspace file or a typed group of representations to the Agent Workbench results pane. Call only after every referenced file exists. Use a new artifact_id for a different tool execution; increment revision by exactly one only when the same producer updates the same artifact. Omit producer.tool_execution_id so Runner binds this MCP execution automatically.",
		InputSchema: workbenchArtifactInputSchema(),
		PodHandler: func(
			ctx context.Context,
			pod *PodInfo,
			args map[string]interface{},
		) (interface{}, error) {
			if s.artifactPublisher == nil {
				return nil, fmt.Errorf("workbench artifact publisher is unavailable")
			}
			declaration, ok := args["declaration"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("declaration must be an object")
			}
			raw, err := json.Marshal(declaration)
			if err != nil {
				return nil, err
			}
			executionID := "workbench-publish:" + uuid.NewString()
			return s.artifactPublisher.PublishWorkbenchArtifact(
				ctx,
				pod.PodKey,
				executionID,
				raw,
			)
		},
	}
}

func workbenchArtifactInputSchema() map[string]interface{} {
	identifier := map[string]interface{}{
		"type": "string", "pattern": "^[a-z0-9]+(?:-[a-z0-9]+)*$",
		"minLength": 2, "maxLength": 100,
	}
	return map[string]interface{}{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"declaration"},
		"properties": map[string]interface{}{
			"declaration": map[string]interface{}{
				"type":                 "object",
				"additionalProperties": false,
				"required": []string{
					"schema_version", "artifact_id", "revision", "role",
					"primary_representation_id", "producer", "representations",
				},
				"properties": map[string]interface{}{
					"schema_version": map[string]interface{}{
						"type":  "string",
						"const": "agentsmesh.agent-workbench.artifact/v1",
					},
					"artifact_id": identifier,
					"revision": map[string]interface{}{
						"type": "integer", "minimum": 1,
					},
					"role": map[string]interface{}{
						"type": "string", "minLength": 1, "maxLength": 64,
					},
					"primary_representation_id": identifier,
					"producer":                  workbenchProducerSchema(),
					"representations": map[string]interface{}{
						"type": "array", "minItems": 1, "maxItems": 64,
						"items": workbenchRepresentationSchema(),
					},
					"manifest": map[string]interface{}{
						"type":     "object",
						"required": []string{"kind"},
					},
				},
			},
		},
	}
}

func workbenchProducerSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"namespace", "type"},
		"properties": map[string]interface{}{
			"namespace":  map[string]interface{}{"type": "string", "minLength": 1},
			"type":       map[string]interface{}{"type": "string", "minLength": 1},
			"id":         map[string]interface{}{"type": "string"},
			"command_id": map[string]interface{}{"type": "string"},
		},
	}
}

func workbenchRepresentationSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{
			"representation_id", "path", "media_type",
		},
		"properties": map[string]interface{}{
			"representation_id": map[string]interface{}{"type": "string"},
			"path":              map[string]interface{}{"type": "string"},
			"media_type":        map[string]interface{}{"type": "string"},
			"role":              map[string]interface{}{"type": "string"},
			"dimensions": map[string]interface{}{
				"type":     "object",
				"required": []string{"width", "height"},
				"properties": map[string]interface{}{
					"width":  map[string]interface{}{"type": "integer", "minimum": 1},
					"height": map[string]interface{}{"type": "integer", "minimum": 1},
				},
			},
			"duration_millis": map[string]interface{}{
				"type": "integer", "minimum": 0,
			},
		},
	}
}
