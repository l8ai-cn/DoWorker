package connector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocPath(t *testing.T) {
	assert.Equal(t, "Team-Docs-abc123.md", DocPath("Team Docs", "abc123"))
	assert.Equal(t, "产品-规划-tok.md", DocPath("产品 / 规划", "tok"))
	assert.Equal(t, "untitled-id1.md", DocPath("///", "id1"))
	long := DocPath("very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-extra", "1234567890abcdef")
	assert.LessOrEqual(t, len(long), 80+1+12+3)
	assert.Equal(t, "90abcdef.md", long[len(long)-11:])
}

func TestNewRegistry_CoversAllExternalSourceTypes(t *testing.T) {
	reg := NewRegistry()
	for _, st := range []string{"feishu", "dingtalk", "google"} {
		conn, ok := reg[st]
		require.True(t, ok, st)
		assert.Equal(t, st, conn.SourceType())
	}
}

func TestConnectors_RejectBadConfig(t *testing.T) {
	ctx := context.Background()
	for _, conn := range NewRegistry() {
		_, err := conn.ListDocs(ctx, nil)
		assert.ErrorIs(t, err, ErrBadConfig, conn.SourceType())
		_, err = conn.ListDocs(ctx, json.RawMessage(`{}`))
		assert.ErrorIs(t, err, ErrBadConfig, conn.SourceType())
	}
}

func TestFeishuConnector_ListAndFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/open-apis/auth/v3/tenant_access_token/internal":
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "tenant_access_token": "tok"})
		case r.URL.Path == "/open-apis/wiki/v2/spaces/sp1/nodes":
			require.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
			if r.URL.Query().Get("parent_node_token") == "" {
				_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{
					"items": []map[string]any{
						{"node_token": "n1", "obj_token": "doc1", "obj_type": "docx", "title": "Root Doc", "has_child": true},
					},
					"has_more": false,
				}})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{
					"items": []map[string]any{
						{"node_token": "n2", "obj_token": "doc2", "obj_type": "docx", "title": "Child Doc"},
					},
					"has_more": false,
				}})
			}
		case r.URL.Path == "/open-apis/docx/v1/documents/doc1/raw_content":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"content": "body text"}})
		default:
			t.Errorf("unexpected call %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	conn := &FeishuConnector{HTTP: srv.Client(), BaseURL: srv.URL}
	cfg := json.RawMessage(`{"app_id":"a","app_secret":"s","space_id":"sp1"}`)

	refs, err := conn.ListDocs(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, refs, 2)
	assert.Equal(t, "doc1", refs[0].ID)
	assert.Equal(t, "doc2", refs[1].ID, "child nodes must be walked")

	doc, err := conn.FetchDoc(context.Background(), cfg, refs[0])
	require.NoError(t, err)
	assert.Contains(t, doc.Markdown, "# Root Doc")
	assert.Contains(t, doc.Markdown, "body text")
}

func TestGoogleDriveConnector_ListAndFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer at", r.Header.Get("Authorization"))
		switch {
		case r.URL.Path == "/drive/v3/files":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"files": []map[string]any{{"id": "f1", "name": "Spec"}},
			})
		case r.URL.Path == "/drive/v3/files/f1/export":
			assert.Equal(t, "text/markdown", r.URL.Query().Get("mimeType"))
			_, _ = w.Write([]byte("# Spec\ncontent"))
		default:
			t.Errorf("unexpected call %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	conn := &GoogleDriveConnector{HTTP: srv.Client(), BaseURL: srv.URL}
	cfg := json.RawMessage(`{"access_token":"at","folder_id":"folder"}`)

	refs, err := conn.ListDocs(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, refs, 1)

	doc, err := conn.FetchDoc(context.Background(), cfg, refs[0])
	require.NoError(t, err)
	assert.Equal(t, "# Spec\ncontent", doc.Markdown)
}
