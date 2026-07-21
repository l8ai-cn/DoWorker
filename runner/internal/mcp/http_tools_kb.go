package mcp

import (
	"context"
	"fmt"

	"github.com/l8ai-cn/agentcloud/runner/internal/mcp/tools"
)

// Knowledge Base tools — llm-wiki workflow over org KBs. Mounted KBs are
// also available as plain git checkouts under kb/{slug}; these tools cover
// cross-KB search and access to KBs that are not mounted in this pod.

func (s *HTTPServer) createKbListTool() *MCPTool {
	return &MCPTool{
		Name:        "kb_list",
		Description: "List knowledge bases available in this organization. Mounted KBs are also cloned under kb/{slug} in the sandbox.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, client tools.CollaborationClient, args map[string]interface{}) (interface{}, error) {
			return client.KbList(ctx)
		},
	}
}

func (s *HTTPServer) createKbSearchTool() *MCPTool {
	return &MCPTool{
		Name:        "kb_search",
		Description: "Search wiki pages (plus llms.txt/AGENTS.md) across knowledge bases with a case-insensitive text query.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Text to search for",
				},
				"kb_slugs": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Limit search to these KB slugs. Omit to search all KBs in the organization.",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum matches (default: 20)",
				},
			},
			"required": []string{"query"},
		},
		Handler: func(ctx context.Context, client tools.CollaborationClient, args map[string]interface{}) (interface{}, error) {
			query := getStringArg(args, "query")
			if query == "" {
				return nil, fmt.Errorf("query is required")
			}
			var slugs []string
			if raw, ok := args["kb_slugs"].([]interface{}); ok {
				for _, v := range raw {
					if s, ok := v.(string); ok {
						slugs = append(slugs, s)
					}
				}
			}
			return client.KbSearch(ctx, query, slugs, getIntArg(args, "limit"))
		},
	}
}

func (s *HTTPServer) createKbReadTool() *MCPTool {
	return &MCPTool{
		Name:        "kb_read",
		Description: "Read one file from a knowledge base (e.g. llms.txt, wiki/index.md). Prefer reading mounted KBs directly from kb/{slug}.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"kb_slug": map[string]interface{}{
					"type":        "string",
					"description": "Knowledge base slug (see kb_list)",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File path inside the KB repo, e.g. 'wiki/index.md'",
				},
			},
			"required": []string{"kb_slug", "path"},
		},
		Handler: func(ctx context.Context, client tools.CollaborationClient, args map[string]interface{}) (interface{}, error) {
			kbSlug := getStringArg(args, "kb_slug")
			path := getStringArg(args, "path")
			if kbSlug == "" || path == "" {
				return nil, fmt.Errorf("kb_slug and path are required")
			}
			return client.KbRead(ctx, kbSlug, path)
		},
	}
}

func (s *HTTPServer) createKbWriteTool() *MCPTool {
	return &MCPTool{
		Name:        "kb_write",
		Description: "Commit one file to a knowledge base via the platform (for KBs not mounted read-write). For rw-mounted KBs prefer editing kb/{slug} and git commit+push. Follow the KB's AGENTS.md and update wiki/log.md.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"kb_slug": map[string]interface{}{
					"type":        "string",
					"description": "Knowledge base slug (see kb_list)",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File path inside the KB repo, e.g. 'wiki/topic.md'",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Full new file content",
				},
				"message": map[string]interface{}{
					"type":        "string",
					"description": "Commit message (optional)",
				},
			},
			"required": []string{"kb_slug", "path", "content"},
		},
		Handler: func(ctx context.Context, client tools.CollaborationClient, args map[string]interface{}) (interface{}, error) {
			kbSlug := getStringArg(args, "kb_slug")
			path := getStringArg(args, "path")
			if kbSlug == "" || path == "" {
				return nil, fmt.Errorf("kb_slug and path are required")
			}
			content, ok := args["content"].(string)
			if !ok {
				return nil, fmt.Errorf("content is required")
			}
			return client.KbWrite(ctx, kbSlug, path, content, getStringArg(args, "message"))
		},
	}
}
