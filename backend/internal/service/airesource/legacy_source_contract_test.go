package airesource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type forbiddenLegacySymbol struct {
	name    string
	parts   []string
	allowed []string
}

func TestNoLegacyModelCredentialContracts(t *testing.T) {
	root := repoRoot(t)
	forbidden := []forbiddenLegacySymbol{
		{name: "agent credential service", parts: []string{"UserAgent", "CredentialService"}},
		{name: "credential profile field", parts: []string{"credential", "_profile_id"}},
		{name: "implicit auth fallback", parts: []string{"useAgent", "DefaultAuth"}},
		{name: "primary credential bundle auto mount", parts: []string{"AppendPrimary", "CredentialBundle"}},
		{name: "legacy model config field", parts: []string{"model", "_config_id"}},
		{name: "legacy model config route", parts: []string{"/model-", "configs"}},
		{name: "legacy model config client", parts: []string{"list", "ModelConfigs"}},
		{name: "legacy ai model token field", parts: []string{"ai_", "model_id"}, allowed: []string{
			"backend/internal/service/airesource/migration",
			"backend/migrations/",
		}},
	}
	failures := []string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel := filepath.ToSlash(relPath)
		if shouldSkipLegacyContractPath(rel, entry) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		for _, symbol := range forbidden {
			needle := strings.Join(symbol.parts, "")
			if !strings.Contains(text, needle) || allowedLegacyContractPath(rel, symbol.allowed) {
				continue
			}
			failures = append(failures, rel+": "+symbol.name+" contains "+needle)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(failures) > 0 {
		t.Fatalf("legacy model credential contracts remain:\n%s", strings.Join(failures, "\n"))
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root with go.mod not found")
		}
		dir = parent
	}
}

func shouldSkipLegacyContractPath(rel string, entry os.DirEntry) bool {
	if rel == "backend/internal/service/airesource/legacy_source_contract_test.go" {
		return true
	}
	for _, prefix := range []string{
		".git", "bazel-", "node_modules", "docs/superpowers/plans/", "report.xml",
	} {
		if strings.HasPrefix(rel, prefix) {
			return true
		}
	}
	if entry.IsDir() {
		return strings.HasPrefix(entry.Name(), ".")
	}
	return !(strings.HasSuffix(rel, ".go") || strings.HasSuffix(rel, ".proto") ||
		strings.HasSuffix(rel, ".ts") || strings.HasSuffix(rel, ".tsx") ||
		strings.HasSuffix(rel, ".json") || strings.HasSuffix(rel, ".md"))
}

func allowedLegacyContractPath(rel string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(rel, prefix) {
			return true
		}
	}
	return false
}
