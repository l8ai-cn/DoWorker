package workflow

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRepoQuery implements RepoQueryForWorkflow for testing.
type mockRepoQuery struct {
	repo *gitprovider.Repository
	err  error
}

func (m *mockRepoQuery) GetByID(_ context.Context, _ int64) (*gitprovider.Repository, error) {
	return m.repo, m.err
}

func newTestOrchestrator(repoQuery RepoQueryForWorkflow) *WorkflowOrchestrator {
	return &WorkflowOrchestrator{
		logger:    slog.Default(),
		repoQuery: repoQuery,
	}
}

func TestBuildLoopAgentfileLayer_PermissionMode(t *testing.T) {
	o := newTestOrchestrator(nil)
	workflow := &workflowDomain.Workflow{PermissionMode: "bypassPermissions"}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	assert.Contains(t, layer, `CONFIG permission_mode = "bypassPermissions"`)
}

func TestBuildLoopAgentfileLayer_DefaultPermissionMode(t *testing.T) {
	o := newTestOrchestrator(nil)
	workflow := &workflowDomain.Workflow{PermissionMode: ""}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	assert.Contains(t, layer, `CONFIG permission_mode = "bypassPermissions"`)
}

func TestBuildLoopAgentfileLayer_CustomPermissionMode(t *testing.T) {
	o := newTestOrchestrator(nil)
	workflow := &workflowDomain.Workflow{PermissionMode: "askUser"}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	assert.Contains(t, layer, `CONFIG permission_mode = "askUser"`)
	assert.NotContains(t, layer, "bypassPermissions")
}

func TestBuildLoopAgentfileLayer_ConfigOverrides(t *testing.T) {
	o := newTestOrchestrator(nil)
	overrides := map[string]interface{}{
		"model":       "opus",
		"mcp_enabled": true,
	}
	raw, err := json.Marshal(overrides)
	require.NoError(t, err)

	workflow := &workflowDomain.Workflow{ConfigOverrides: json.RawMessage(raw)}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	assert.Contains(t, layer, `CONFIG model = "opus"`)
	assert.Contains(t, layer, `CONFIG mcp_enabled = true`)
}

func TestBuildLoopAgentfileLayer_ConfigOverrides_SkipsPermissionMode(t *testing.T) {
	o := newTestOrchestrator(nil)
	overrides := map[string]interface{}{
		"permission_mode": "shouldBeIgnored",
		"model":           "sonnet",
	}
	raw, err := json.Marshal(overrides)
	require.NoError(t, err)

	workflow := &workflowDomain.Workflow{ConfigOverrides: json.RawMessage(raw)}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	// permission_mode from ConfigOverrides is skipped; only the top-level one is used
	lines := strings.Split(layer, "\n")
	permCount := 0
	for _, line := range lines {
		if strings.Contains(line, "permission_mode") {
			permCount++
		}
	}
	assert.Equal(t, 1, permCount, "permission_mode should appear exactly once")
	assert.Contains(t, layer, `CONFIG model = "sonnet"`)
}

func TestBuildLoopAgentfileLayer_Prompt(t *testing.T) {
	o := newTestOrchestrator(nil)
	workflow := &workflowDomain.Workflow{}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "deploy the app")

	assert.Contains(t, layer, `PROMPT "deploy the app"`)
}

func TestBuildLoopAgentfileLayer_PromptEmpty(t *testing.T) {
	o := newTestOrchestrator(nil)
	workflow := &workflowDomain.Workflow{}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	assert.NotContains(t, layer, "PROMPT")
}

func TestBuildLoopAgentfileLayer_PromptEscaping(t *testing.T) {
	o := newTestOrchestrator(nil)
	workflow := &workflowDomain.Workflow{}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, `say "hello" and use \n newline`)

	assert.Contains(t, layer, `PROMPT "say \"hello\" and use \\n newline"`)
}

func TestBuildLoopAgentfileLayer_RepoSlug(t *testing.T) {
	repoID := int64(42)
	mock := &mockRepoQuery{
		repo: &gitprovider.Repository{Slug: "my-org/my-repo", DefaultBranch: "main"},
	}
	o := newTestOrchestrator(mock)
	workflow := &workflowDomain.Workflow{RepositoryID: &repoID}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	assert.Contains(t, layer, `REPO "my-org/my-repo"`)
	// Default branch is used when BranchName is nil
	assert.Contains(t, layer, `BRANCH "main"`)
}

func TestBuildLoopAgentfileLayer_Branch(t *testing.T) {
	repoID := int64(42)
	branchName := "feature/deploy"
	mock := &mockRepoQuery{
		repo: &gitprovider.Repository{Slug: "org/repo", DefaultBranch: "main"},
	}
	o := newTestOrchestrator(mock)
	workflow := &workflowDomain.Workflow{
		RepositoryID: &repoID,
		BranchName:   &branchName,
	}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	assert.Contains(t, layer, `REPO "org/repo"`)
	assert.Contains(t, layer, `BRANCH "feature/deploy"`)
	assert.NotContains(t, layer, `BRANCH "main"`)
}

func TestBuildLoopAgentfileLayer_RepoQueryError(t *testing.T) {
	repoID := int64(99)
	mock := &mockRepoQuery{err: assert.AnError}
	o := newTestOrchestrator(mock)
	workflow := &workflowDomain.Workflow{RepositoryID: &repoID}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	assert.NotContains(t, layer, "REPO")
	assert.NotContains(t, layer, "BRANCH")
}

func TestBuildLoopAgentfileLayer_NoRepoQuery(t *testing.T) {
	repoID := int64(1)
	o := newTestOrchestrator(nil) // repoQuery is nil
	workflow := &workflowDomain.Workflow{RepositoryID: &repoID}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	assert.NotContains(t, layer, "REPO")
}

func TestBuildLoopAgentfileLayer_FullCombination(t *testing.T) {
	repoID := int64(10)
	branch := "develop"
	overrides := map[string]interface{}{"model": "opus"}
	raw, _ := json.Marshal(overrides)

	mock := &mockRepoQuery{
		repo: &gitprovider.Repository{Slug: "team/project", DefaultBranch: "main"},
	}
	o := newTestOrchestrator(mock)

	workflow := &workflowDomain.Workflow{
		PermissionMode:  "bypassPermissions",
		ConfigOverrides: json.RawMessage(raw),
		RepositoryID:    &repoID,
		BranchName:      &branch,
	}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "run tests")

	assert.Contains(t, layer, `PROMPT "run tests"`)
	assert.Contains(t, layer, `CONFIG permission_mode = "bypassPermissions"`)
	assert.Contains(t, layer, `CONFIG model = "opus"`)
	assert.Contains(t, layer, `REPO "team/project"`)
	assert.Contains(t, layer, `BRANCH "develop"`)
}

// ---------- USE_ENV_BUNDLE (workflow bundle binding) ----------

func TestBuildLoopAgentfileLayer_UsedEnvBundles_SingleEmitsLine(t *testing.T) {
	o := newTestOrchestrator(nil)
	workflow := &workflowDomain.Workflow{
		PermissionMode: "bypassPermissions",
		UsedEnvBundles: []string{"work-creds"},
	}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	assert.Contains(t, layer, `USE_ENV_BUNDLE "work-creds"`)
}

func TestBuildLoopAgentfileLayer_UsedEnvBundles_MultipleEmitLinesInOrder(t *testing.T) {
	// Each name becomes its own USE_ENV_BUNDLE line, in array order.
	// Later entries override earlier ones at backend eval time, so order
	// is part of the wire contract.
	o := newTestOrchestrator(nil)
	workflow := &workflowDomain.Workflow{
		PermissionMode: "bypassPermissions",
		UsedEnvBundles: []string{"base", "overlay-1", "overlay-2"},
	}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	idxBase := strings.Index(layer, `USE_ENV_BUNDLE "base"`)
	idxOverlay1 := strings.Index(layer, `USE_ENV_BUNDLE "overlay-1"`)
	idxOverlay2 := strings.Index(layer, `USE_ENV_BUNDLE "overlay-2"`)
	assert.True(t, idxBase >= 0 && idxOverlay1 > idxBase && idxOverlay2 > idxOverlay1,
		"USE_ENV_BUNDLE lines must appear in array order; got layer:\n%s", layer)
}

func TestBuildLoopAgentfileLayer_EmptyUsedEnvBundles_OmitsLine(t *testing.T) {
	o := newTestOrchestrator(nil)
	workflow := &workflowDomain.Workflow{
		PermissionMode: "bypassPermissions",
		UsedEnvBundles: nil,
	}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	assert.NotContains(t, layer, "USE_ENV_BUNDLE")
}

func TestBuildLoopAgentfileLayer_UsedEnvBundles_EmptyStringEntrySkipped(t *testing.T) {
	// Empty-string array entries are skipped silently — guards against
	// stray "" sneaking in from form state without crashing the run.
	o := newTestOrchestrator(nil)
	workflow := &workflowDomain.Workflow{
		PermissionMode: "bypassPermissions",
		UsedEnvBundles: []string{"", "work-creds", ""},
	}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	assert.Contains(t, layer, `USE_ENV_BUNDLE "work-creds"`)
	assert.NotContains(t, layer, `USE_ENV_BUNDLE ""`)
}

func TestBuildLoopAgentfileLayer_UsedEnvBundlesEscapesQuotes(t *testing.T) {
	o := newTestOrchestrator(nil)
	tricky := `name with "quotes" and \ backslash`
	workflow := &workflowDomain.Workflow{
		PermissionMode: "bypassPermissions",
		UsedEnvBundles: []string{tricky},
	}

	layer := o.buildWorkflowAgentfileLayer(context.Background(), workflow, "")

	// Should serialize as: USE_ENV_BUNDLE "name with \"quotes\" and \\ backslash"
	assert.Contains(t, layer, `USE_ENV_BUNDLE "name with \"quotes\" and \\ backslash"`)
}
