package runner

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandboxFsUploadStreamsWorkspaceFile(t *testing.T) {
	want := []byte("complete seedance video")
	var got []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "video/mp4", r.Header.Get("Content-Type"))
		var err error
		got, err = io.ReadAll(r.Body)
		require.NoError(t, err)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "result.mp4"), want, 0o644))

	result, err := (&RunnerMessageHandler{}).sandboxFsUpload(root, "result.mp4", server.URL)

	require.NoError(t, err)
	require.Empty(t, result.GetError())
	assert.Equal(t, want, got)
	assert.Equal(t, int64(len(want)), result.GetFileBytes())
	assert.Equal(t, "video/mp4", result.GetContentType())
}

func TestSandboxFsUploadRejectsSymlinkOutsideWorkspace(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "secret.mp4")
	require.NoError(t, os.WriteFile(outside, []byte("secret"), 0o644))
	root := t.TempDir()
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "result.mp4")))
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	t.Cleanup(server.Close)

	result, err := (&RunnerMessageHandler{}).sandboxFsUpload(root, "result.mp4", server.URL)

	require.NoError(t, err)
	assert.NotEmpty(t, result.GetError())
	assert.False(t, called)
}

func TestSandboxFsUploadRejectsRedirect(t *testing.T) {
	targetCalled := false
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		targetCalled = true
	}))
	t.Cleanup(target.Close)
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Redirect(w, &http.Request{}, target.URL, http.StatusTemporaryRedirect)
	}))
	t.Cleanup(redirect.Close)
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "result.mp4"), []byte("video"), 0o644))

	result, err := (&RunnerMessageHandler{}).sandboxFsUpload(root, "result.mp4", redirect.URL)

	require.NoError(t, err)
	assert.Contains(t, result.GetError(), "HTTP 307")
	assert.False(t, targetCalled)
}
