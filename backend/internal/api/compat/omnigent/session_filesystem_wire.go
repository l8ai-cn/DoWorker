package omnigent

import (
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func listWire(entries []*runnerv1.SandboxFsEntry, workspaceRoot string) map[string]any {
	data := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		data = append(data, map[string]any{
			"id":          e.GetPath(),
			"name":        e.GetName(),
			"path":        e.GetPath(),
			"type":        e.GetType(),
			"bytes":       nullableInt(e.GetBytes(), e.GetType() == "directory"),
			"modified_at": nullableInt(e.GetModifiedAt(), false),
		})
	}
	out := map[string]any{"object": "list", "data": data, "has_more": false}
	if workspaceRoot != "" {
		out["workspace_root"] = workspaceRoot
	}
	return out
}

func changesWire(changes []*runnerv1.SandboxFsChange) map[string]any {
	data := make([]map[string]any, 0, len(changes))
	for _, ch := range changes {
		data = append(data, map[string]any{
			"path":        ch.GetPath(),
			"name":        ch.GetName(),
			"status":      ch.GetStatus(),
			"bytes":       nullableInt(ch.GetBytes(), false),
			"modified_at": nullableInt(ch.GetModifiedAt(), false),
		})
	}
	return map[string]any{"object": "list", "data": data, "has_more": false}
}

func fileContentWire(path string, res *runnerv1.SandboxFsResultEvent) map[string]any {
	enc := res.GetEncoding()
	if enc == "" {
		enc = "utf-8"
	}
	return map[string]any{
		"object":       "session.environment.filesystem.file_content",
		"path":         path,
		"content_type": res.GetContentType(),
		"encoding":     enc,
		"content":      res.GetContent(),
		"bytes":        res.GetFileBytes(),
		"truncated":    res.GetTruncated(),
	}
}

func nullableInt(v int64, nullWhenZero bool) any {
	if nullWhenZero && v == 0 {
		return nil
	}
	if v == 0 {
		return nil
	}
	return v
}
