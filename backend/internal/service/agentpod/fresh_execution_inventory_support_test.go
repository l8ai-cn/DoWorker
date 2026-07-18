package agentpod

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func collectFreshExecutionInventoryFields(elements []ast.Expr) []string {
	present := map[string]struct{}{}
	for _, elem := range elements {
		pair, ok := elem.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		name, ok := freshExecutionFieldName(pair.Key)
		if ok {
			present[name] = struct{}{}
		}
	}
	out := make([]string, 0, len(freshExecutionInventoryFields))
	for _, field := range freshExecutionInventoryFields {
		if _, ok := present[field]; ok {
			out = append(out, field)
		}
	}
	return out
}

func freshExecutionFieldName(expr ast.Expr) (string, bool) {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name, true
	case *ast.SelectorExpr:
		return value.Sel.Name, true
	default:
		return "", false
	}
}

func isOrchestrateCreatePodRequest(expression ast.Expr) bool {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name == "OrchestrateCreatePodRequest"
	case *ast.SelectorExpr:
		return value.Sel.Name == "OrchestrateCreatePodRequest"
	default:
		return false
	}
}

func detectBackendDir(t *testing.T) string {
	t.Helper()
	path, err := os.Getwd()
	if err != nil {
		t.Fatalf("detect backend root: %v", err)
	}
	for {
		if filepath.Base(path) == "backend" {
			return path
		}
		parent := filepath.Dir(path)
		if parent == path {
			t.Fatalf("cannot locate backend directory from cwd")
		}
		path = parent
	}
}

func containsField(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func formatInventory(inventory map[string]freshExecutionInventoryEntry) string {
	lines := make([]string, 0, len(inventory))
	keys := keysFromMap(inventory)
	sort.Strings(keys)
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf(
			"%s: %s",
			key,
			toJSON(inventory[key]),
		))
	}
	return strings.Join(lines, "\n")
}

func keysFromMap(values map[string]freshExecutionInventoryEntry) []string {
	out := make([]string, 0, len(values))
	for key := range values {
		out = append(out, key)
	}
	return out
}

func unionStrings(left, right []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(left)+len(right))
	for _, keys := range [][]string{left, right} {
		for _, key := range keys {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, key)
		}
	}
	return out
}

func toJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "<marshal-error>"
	}
	return string(data)
}
