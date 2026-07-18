package sessionfile

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeContentTypeRecognizesAgentDocuments(t *testing.T) {
	require.Equal(t, "text/plain", normalizeContentType("", "main.go"))
	require.Equal(t, "text/markdown", normalizeContentType("", "README.md"))
	require.Equal(t, "application/json", normalizeContentType("", "input.json"))
	require.Equal(t, "text/csv", normalizeContentType("", "data.csv"))
	require.Equal(
		t,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		normalizeContentType("application/octet-stream", "report.docx"),
	)
}
