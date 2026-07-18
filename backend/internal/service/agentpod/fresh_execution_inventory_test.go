package agentpod

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type freshExecutionInventoryEntry struct {
	Mode   string   `json:"mode"`
	Fields []string `json:"fields"`
}

var freshExecutionInventoryFields = []string{
	"RunnerID",
	"AgentSlug",
	"RepositoryID",
	"AgentfileLayer",
	"WorkerSpecDraft",
	"WorkerSpecSnapshotID",
	"WorkerSpecPromptOverride",
	"SourcePodKey",
	"ModelResourceID",
	"LocalPath",
	"BranchName",
	"KnowledgeMounts",
	"AutomationLevel",
}
var forbiddenForSnapshotOrPlan = map[string]struct{}{
	"AgentSlug":       {},
	"RepositoryID":    {},
	"AgentfileLayer":  {},
	"ModelResourceID": {},
	"LocalPath":       {},
	"BranchName":      {},
}

func TestFreshExecutionInventory(t *testing.T) {
	got := discoverFreshExecutionInventory(t)

	expected := map[string]freshExecutionInventoryEntry{
		"backend/internal/api/connect/pod/create_pod_request.go:buildResumePodRequest": {
			Mode: "lineage", Fields: []string{"SourcePodKey"},
		},
		"backend/internal/api/connect/pod/create_pod_request.go:buildWorkerSpecPodRequest": {
			Mode:   "plan",
			Fields: []string{"WorkerSpecDraft"},
		},
		"backend/internal/api/grpc/runner_adapter_mcp_pod.go:mcpCreatePod": {
			Mode:   "legacy",
			Fields: []string{"RunnerID", "AgentSlug", "RepositoryID", "AgentfileLayer", "SourcePodKey"},
		},
		"backend/internal/api/rest/v1/pod_create.go:CreatePod": {
			Mode:   "legacy",
			Fields: []string{"RunnerID", "AgentSlug", "RepositoryID", "AgentfileLayer", "SourcePodKey", "ModelResourceID", "AutomationLevel"},
		},
		"backend/internal/api/rest/v1/quick_task.go:CreateQuickTask": {
			Mode:   "legacy",
			Fields: []string{"RunnerID", "AgentSlug", "RepositoryID", "AgentfileLayer"},
		},
		"backend/internal/api/rest/v1/session/hosts.go:handleBindHostRunner": {
			Mode:   "legacy",
			Fields: []string{"RunnerID", "AgentSlug", "AgentfileLayer", "LocalPath"},
		},
		"backend/internal/api/rest/v1/session/session_create_pod_request.go:sessionCreatePodRequest": {
			Mode:   "legacy",
			Fields: []string{"AgentSlug", "AgentfileLayer", "ModelResourceID", "LocalPath"},
		},
		"backend/internal/api/rest/v1/session/session_fork.go:handleForkSession": {
			Mode:   "legacy",
			Fields: []string{"RunnerID", "AgentSlug", "AgentfileLayer"},
		},
		"backend/internal/api/rest/v1/session/session_import.go:handleImportSession": {
			Mode:   "legacy",
			Fields: []string{"AgentSlug", "AgentfileLayer"},
		},
		"backend/internal/api/rest/v1/session/session_message_pod.go:ensureMessagePod": {Mode: "legacy", Fields: []string{"AgentSlug", "SourcePodKey"}},
		"backend/internal/api/rest/v1/session/session_switch.go:rebuildSessionPod": {
			Mode:   "legacy",
			Fields: []string{"RunnerID", "AgentSlug", "AgentfileLayer"},
		},
		"backend/internal/service/coordinator/dispatch.go:claimAndDispatch": {
			Mode:   "legacy",
			Fields: []string{"AgentSlug", "RepositoryID", "AgentfileLayer"},
		},
		"backend/internal/service/expert/run.go:Run": {
			Mode:   "snapshot",
			Fields: []string{"WorkerSpecSnapshotID", "WorkerSpecPromptOverride"},
		},
		"backend/internal/service/goalloop/goal_loop_start.go:Start": {
			Mode:   "snapshot",
			Fields: []string{"WorkerSpecSnapshotID", "WorkerSpecPromptOverride"},
		},
		"backend/internal/service/mesh/ticket_pod_orchestration.go:CreatePodForTicket": {
			Mode:   "legacy",
			Fields: []string{"RunnerID", "AgentSlug", "AgentfileLayer", "AutomationLevel"},
		},
		"backend/cmd/server/orchestration_worker_launcher.go:MaterializeWorkerPod": {
			Mode:   "snapshot",
			Fields: []string{"WorkerSpecSnapshotID", "WorkerSpecPromptOverride"},
		},
		"backend/internal/service/workflow/workflow_pod_request.go:buildWorkflowRunLineagePodRequest": {
			Mode:   "lineage",
			Fields: []string{"WorkerSpecPromptOverride", "SourcePodKey"},
		},
		"backend/internal/service/workflow/workflow_pod_request.go:buildWorkflowRunSnapshotPodRequest": {
			Mode:   "snapshot",
			Fields: []string{"WorkerSpecSnapshotID", "WorkerSpecPromptOverride"},
		},
	}

	assertModeAndFieldRules(t, got)
	assertInventoryDiff(t, expected, got)
}

func assertModeAndFieldRules(t *testing.T, inventory map[string]freshExecutionInventoryEntry) {
	t.Helper()
	for key, entry := range inventory {
		if entry.Mode != "legacy" && entry.Mode != "plan" && entry.Mode != "snapshot" && entry.Mode != "lineage" {
			t.Errorf("%s: invalid mode %q; expected legacy|plan|snapshot|lineage", key, entry.Mode)
		}
		assertFreshExecutionFieldRules(t, key, entry)
	}
}

func assertFreshExecutionFieldRules(t *testing.T, key string, entry freshExecutionInventoryEntry) {
	t.Helper()
	if entry.Mode == "snapshot" && !containsField(entry.Fields, "WorkerSpecSnapshotID") {
		t.Errorf("%s snapshot mode must include WorkerSpecSnapshotID", key)
	}
	if entry.Mode == "snapshot" || entry.Mode == "plan" {
		for _, field := range entry.Fields {
			if _, ok := forbiddenForSnapshotOrPlan[field]; ok {
				t.Errorf("%s mode %s must not include %s", key, entry.Mode, field)
			}
		}
	}
}

func assertInventoryDiff(t *testing.T, expected, got map[string]freshExecutionInventoryEntry) {
	t.Helper()
	keys := unionStrings(keysFromMap(expected), keysFromMap(got))
	sort.Strings(keys)

	changed := make([]string, 0)
	for _, key := range keys {
		exp, okExpected := expected[key]
		act, okActual := got[key]
		if !okExpected {
			changed = append(changed, fmt.Sprintf("unexpected construction: %s", key))
			continue
		}
		if !okActual {
			changed = append(changed, fmt.Sprintf("missing construction: %s", key))
			continue
		}
		if !reflect.DeepEqual(exp, act) {
			changed = append(changed, fmt.Sprintf("- %s\n  expected: %s\n  actual: %s", key, toJSON(exp), toJSON(act)))
		}
	}

	if len(changed) > 0 {
		t.Fatalf("fresh execution inventory drift detected:\n%s", strings.Join(changed, "\n"))
	}
}

func discoverFreshExecutionInventory(t *testing.T) map[string]freshExecutionInventoryEntry {
	t.Helper()
	backendDir := detectBackendDir(t)
	repoDir := filepath.Dir(backendDir)
	fset := token.NewFileSet()
	inventory := map[string]freshExecutionInventoryEntry{}
	duplicates := map[string]int{}

	walkErr := filepath.WalkDir(backendDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		relPath, err := filepath.Rel(repoDir, path)
		if err != nil {
			return fmt.Errorf("resolve relative path %s: %w", path, err)
		}
		relPath = filepath.ToSlash(relPath)

		collectFreshExecutionInventoryInFile(file, relPath, duplicates, inventory)
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk backend directory: %v", walkErr)
	}

	return inventory
}

func collectFreshExecutionInventoryInFile(file *ast.File, relPath string, duplicates map[string]int, inventory map[string]freshExecutionInventoryEntry) {
	funcNames := []string{}
	var walk func(ast.Node)
	walk = func(node ast.Node) {
		ast.Inspect(node, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			switch value := n.(type) {
			case *ast.FuncDecl:
				if value.Body == nil {
					return false
				}
				funcNames = append(funcNames, value.Name.Name)
				walk(value.Body)
				funcNames = funcNames[:len(funcNames)-1]
				return false
			case *ast.CompositeLit:
				if isOrchestrateCreatePodRequest(value.Type) {
					funcName := "file_scope"
					if len(funcNames) > 0 {
						funcName = funcNames[len(funcNames)-1]
					}
					key := freshExecutionKey(relPath, funcName, duplicates)
					inventory[key] = freshExecutionInventoryEntry{
						Mode:   classifyMode(relPath, funcName),
						Fields: collectFreshExecutionInventoryFields(value.Elts),
					}
				}
			}
			return true
		})
	}
	walk(file)
}

func freshExecutionKey(relPath, funcName string, duplicates map[string]int) string {
	base := relPath + ":" + funcName
	if n, ok := duplicates[base]; ok {
		n++
		duplicates[base] = n
		return fmt.Sprintf("%s#%d", base, n)
	}
	duplicates[base] = 0
	return base
}

func classifyMode(relPath, funcName string) string {
	if strings.HasSuffix(relPath, "internal/api/connect/pod/create_pod_request.go") {
		switch funcName {
		case "buildWorkerSpecPodRequest":
			return "plan"
		case "buildResumePodRequest":
			return "lineage"
		}
	}
	if strings.HasSuffix(
		relPath,
		"cmd/server/orchestration_worker_launcher.go",
	) && funcName == "MaterializeWorkerPod" {
		return "snapshot"
	}
	if strings.HasSuffix(relPath, "internal/service/goalloop/goal_loop_start.go") && funcName == "Start" {
		return "snapshot"
	}
	if strings.HasSuffix(relPath, "internal/service/expert/run.go") && funcName == "Run" {
		return "snapshot"
	}
	if strings.HasSuffix(relPath, "internal/service/workflow/workflow_pod_request.go") {
		switch funcName {
		case "buildWorkflowRunSnapshotPodRequest":
			return "snapshot"
		case "buildWorkflowRunLineagePodRequest":
			return "lineage"
		}
	}
	return "legacy"
}

func collectFreshExecutionInventoryFields(elements []ast.Expr) []string {
	present := map[string]struct{}{}
	for _, elem := range elements {
		pair, ok := elem.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		name, ok := freshExecutionFieldName(pair.Key)
		if !ok {
			continue
		}
		present[name] = struct{}{}
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
	for _, key := range left {
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	for _, key := range right {
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
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
