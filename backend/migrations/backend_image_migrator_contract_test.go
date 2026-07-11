package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestBackendImageIncludesAIResourceMigrator(t *testing.T) {
	dockerfile, err := os.ReadFile("../Dockerfile")
	if err != nil {
		t.Fatalf("read backend Dockerfile: %v", err)
	}
	content := string(dockerfile)

	for _, required := range []string{
		"-o /out/migrate-ai-resources ./backend/cmd/migrate-ai-resources",
		"COPY --from=build --chown=1000:1000 /out/migrate-ai-resources /app/migrate-ai-resources",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("backend image must contain %q", required)
		}
	}
}
