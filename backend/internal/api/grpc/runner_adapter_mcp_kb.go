package grpc

import (
	"context"
	"errors"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	kbservice "github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase"
)

func (a *GRPCRunnerAdapter) mcpKbList(ctx context.Context, tc *middleware.TenantContext) (interface{}, *mcpError) {
	if a.knowledgebaseService == nil {
		return nil, newMcpError(501, "knowledge base service not configured")
	}
	kbs, err := a.knowledgebaseService.List(ctx, tc.OrganizationID, "")
	if err != nil {
		return nil, kbErrToMcp(err)
	}
	items := make([]map[string]interface{}, 0, len(kbs))
	for _, kb := range kbs {
		items = append(items, map[string]interface{}{
			"slug":           kb.Slug,
			"name":           kb.Name,
			"description":    kb.Description,
			"source_type":    kb.SourceType,
			"default_branch": kb.DefaultBranch,
		})
	}
	return map[string]interface{}{"knowledge_bases": items}, nil
}

func (a *GRPCRunnerAdapter) mcpKbSearch(ctx context.Context, tc *middleware.TenantContext, payload []byte) (interface{}, *mcpError) {
	if a.knowledgebaseService == nil {
		return nil, newMcpError(501, "knowledge base service not configured")
	}
	var params struct {
		Query   string   `json:"query"`
		KBSlugs []string `json:"kb_slugs"`
		Limit   int      `json:"limit"`
	}
	if err := unmarshalPayload(payload, &params); err != nil {
		return nil, err
	}
	if strings.TrimSpace(params.Query) == "" {
		return nil, newMcpError(400, "query is required")
	}
	matches, err := a.knowledgebaseService.Search(ctx, tc.OrganizationID, params.KBSlugs, params.Query, params.Limit)
	if err != nil {
		return nil, kbErrToMcp(err)
	}
	return map[string]interface{}{"matches": matches}, nil
}

func (a *GRPCRunnerAdapter) mcpKbRead(ctx context.Context, tc *middleware.TenantContext, payload []byte) (interface{}, *mcpError) {
	if a.knowledgebaseService == nil {
		return nil, newMcpError(501, "knowledge base service not configured")
	}
	var params struct {
		KBSlug string `json:"kb_slug"`
		Path   string `json:"path"`
	}
	if err := unmarshalPayload(payload, &params); err != nil {
		return nil, err
	}
	if params.KBSlug == "" || params.Path == "" {
		return nil, newMcpError(400, "kb_slug and path are required")
	}
	file, err := a.knowledgebaseService.ReadFile(ctx, tc.OrganizationID, params.KBSlug, params.Path)
	if err != nil {
		return nil, kbErrToMcp(err)
	}
	return map[string]interface{}{
		"kb_slug": params.KBSlug,
		"path":    file.Path,
		"content": file.Content,
		"size":    file.Size,
	}, nil
}

func (a *GRPCRunnerAdapter) mcpKbWrite(ctx context.Context, tc *middleware.TenantContext, podKey string, payload []byte) (interface{}, *mcpError) {
	if a.knowledgebaseService == nil {
		return nil, newMcpError(501, "knowledge base service not configured")
	}
	var params struct {
		KBSlug  string `json:"kb_slug"`
		Path    string `json:"path"`
		Content string `json:"content"`
		Message string `json:"message"`
	}
	if err := unmarshalPayload(payload, &params); err != nil {
		return nil, err
	}
	if params.KBSlug == "" || params.Path == "" {
		return nil, newMcpError(400, "kb_slug and path are required")
	}
	message := params.Message
	if message == "" {
		message = "kb_write: update " + params.Path
	}
	if err := a.knowledgebaseService.CommitFile(
		ctx, tc.OrganizationID, params.KBSlug, params.Path, params.Content, message, "pod:"+podKey,
	); err != nil {
		return nil, kbErrToMcp(err)
	}
	return map[string]interface{}{"status": "committed", "kb_slug": params.KBSlug, "path": params.Path}, nil
}

func kbErrToMcp(err error) *mcpError {
	switch {
	case errors.Is(err, kbservice.ErrNotFound):
		return newMcpError(404, err.Error())
	case errors.Is(err, kbservice.ErrInvalidInput):
		return newMcpError(400, err.Error())
	case errors.Is(err, kbservice.ErrNotConfigured):
		return newMcpError(501, err.Error())
	default:
		return newMcpError(500, err.Error())
	}
}
