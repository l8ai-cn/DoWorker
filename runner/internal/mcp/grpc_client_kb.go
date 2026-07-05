package mcp

import (
	"context"
)

// ==================== KnowledgeBaseClient ====================

func (c *GRPCCollaborationClient) KbList(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.call(ctx, "kb_list", map[string]interface{}{}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *GRPCCollaborationClient) KbSearch(ctx context.Context, query string, kbSlugs []string, limit int) (map[string]interface{}, error) {
	params := map[string]interface{}{"query": query}
	if len(kbSlugs) > 0 {
		params["kb_slugs"] = kbSlugs
	}
	if limit > 0 {
		params["limit"] = limit
	}
	var result map[string]interface{}
	if err := c.call(ctx, "kb_search", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *GRPCCollaborationClient) KbRead(ctx context.Context, kbSlug, path string) (map[string]interface{}, error) {
	params := map[string]interface{}{"kb_slug": kbSlug, "path": path}
	var result map[string]interface{}
	if err := c.call(ctx, "kb_read", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *GRPCCollaborationClient) KbWrite(ctx context.Context, kbSlug, path, content, message string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"kb_slug": kbSlug,
		"path":    path,
		"content": content,
	}
	if message != "" {
		params["message"] = message
	}
	var result map[string]interface{}
	if err := c.call(ctx, "kb_write", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}
