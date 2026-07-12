package runner

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSandboxFsDownloadWritesBinaryFile(t *testing.T) {
	want := []byte{0, 1, 2, 0xff, 4}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(want)
	}))
	t.Cleanup(server.Close)

	root := t.TempDir()
	result, err := (&RunnerMessageHandler{}).sandboxFsDownload(root, "uploads/input.bin", server.URL)

	require.NoError(t, err)
	require.Empty(t, result.GetError())
	got, err := os.ReadFile(filepath.Join(root, "uploads", "input.bin"))
	require.NoError(t, err)
	require.Equal(t, want, got)
}
