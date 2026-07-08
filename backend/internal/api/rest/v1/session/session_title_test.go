package sessionapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveSessionTitleFromPrompt(t *testing.T) {
	t.Parallel()
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name   string
		prompt string
		want   *string
	}{
		{name: "empty", prompt: "", want: nil},
		{name: "plain", prompt: "Build a gomoku game", want: strPtr("Build a gomoku game")},
		{
			name: "attachment markers stripped",
			prompt: "[Attached: src/main.ts]\n[Attached file: README.md]\nFix the login bug",
			want:   strPtr("Fix the login bug"),
		},
		{
			name: "blockquote skipped",
			prompt: "> quoted context\n\nImplement dark mode",
			want:   strPtr("Implement dark mode"),
		},
		{
			name:   "first line only",
			prompt: "Short title line\n\nLonger body follows",
			want:   strPtr("Short title line"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := deriveSessionTitleFromPrompt(tt.prompt)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, *tt.want, *got)
		})
	}
}

func TestTruncateSessionTitle(t *testing.T) {
	t.Parallel()
	long := stringsRepeat("word ", 20)
	got := truncateSessionTitle(long, 60)
	require.LessOrEqual(t, len([]rune(got)), 60)
	assert.True(t, stringsHasSuffix(got, "…"))
}

func stringsRepeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func stringsHasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
