package orchestrationworker

import (
	"context"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerCreationCompilerReturnsPreparedWorkerSpecArtifact(t *testing.T) {
	service := &workerPreflightStub{
		revision: "runtime-catalog-7",
		result: workercreation.PreflightResult{
			Resolved:        &workercreation.Prepared{},
			OptionsRevision: "runtime-catalog-7",
		},
	}
	compiler := newWorkerCreationCompiler(
		service,
		func(*workercreation.Prepared) []byte {
			return []byte(`{"version":1}`)
		},
	)

	result, err := compiler.Compile(
		context.Background(),
		workerTemplateScope(),
		workercreation.Draft{OptionsRevision: "runtime-catalog-7"},
	)

	require.NoError(t, err)
	assert.JSONEq(t, `{"version":1}`, string(result.ArtifactJSON))
	assert.Empty(t, result.Issues)
}

func TestWorkerCreationCompilerRedactsPreflightMessages(t *testing.T) {
	service := &workerPreflightStub{
		revision: "runtime-catalog-7",
		result: workercreation.PreflightResult{
			BlockingErrors: []workercreation.Issue{{
				Code: "invalid-draft", Field: "worker_spec.type_config.automation_level",
				Message: "sk-live-must-not-appear", Severity: "blocking",
			}},
			Warnings: []workercreation.Issue{{
				Code: "workspace-warning", Field: "worker_spec.workspace.branch",
				Message: "ghp_must-not-appear", Severity: "warning",
			}},
			OptionsRevision: "runtime-catalog-7",
		},
	}
	compiler := newWorkerCreationCompiler(service, nil)

	result, err := compiler.Compile(
		context.Background(),
		workerTemplateScope(),
		workercreation.Draft{OptionsRevision: "runtime-catalog-7"},
	)

	require.NoError(t, err)
	require.Len(t, result.Issues, 2)
	assert.Equal(t, control.PlanIssueBlocking, result.Issues[0].Severity)
	assert.Equal(t, "/spec/typeConfig/automationLevel", result.Issues[0].Path)
	assert.NotContains(t, result.Issues[0].Message, "sk-live")
	assert.Equal(t, control.PlanIssueWarning, result.Issues[1].Severity)
	assert.Equal(t, "/spec/workspace/branch", result.Issues[1].Path)
	assert.NotContains(t, result.Issues[1].Message, "ghp_")
	assert.Empty(t, result.ArtifactJSON)
}

func TestWorkerCreationCompilerPreservesHyphenatedToolModelRolePath(t *testing.T) {
	service := &workerPreflightStub{
		revision: "runtime-catalog-7",
		result: workercreation.PreflightResult{
			BlockingErrors: []workercreation.Issue{{
				Code: "invalid-draft", Field: "worker_spec.tool_model_resource_ids.seedance-video",
				Message: "must-not-appear", Severity: "blocking",
			}},
			OptionsRevision: "runtime-catalog-7",
		},
	}
	compiler := newWorkerCreationCompiler(service, nil)

	result, err := compiler.Compile(
		context.Background(),
		workerTemplateScope(),
		workercreation.Draft{OptionsRevision: "runtime-catalog-7"},
	)

	require.NoError(t, err)
	require.Len(t, result.Issues, 1)
	assert.Equal(
		t,
		"/spec/toolModelResourceIds/seedance-video",
		result.Issues[0].Path,
	)
	assert.NotContains(t, result.Issues[0].Message, "must-not-appear")
}

func TestWorkerCreationCompilerExplainsIncompatibleModelSelection(t *testing.T) {
	service := &workerPreflightStub{
		revision: "runtime-catalog-7",
		result: workercreation.PreflightResult{
			BlockingErrors: []workercreation.Issue{{
				Code: "invalid-draft", Field: "worker_spec.model_resource_id",
				Message: "provider credential must-not-appear", Severity: "blocking",
			}},
			OptionsRevision: "runtime-catalog-7",
		},
	}
	compiler := newWorkerCreationCompiler(service, nil)

	result, err := compiler.Compile(
		context.Background(),
		workerTemplateScope(),
		workercreation.Draft{OptionsRevision: "runtime-catalog-7"},
	)

	require.NoError(t, err)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "/spec/modelResourceId", result.Issues[0].Path)
	assert.Equal(t,
		"The selected model is incompatible with this Worker type.",
		result.Issues[0].Message,
	)
}

func TestWorkerCreationCompilerPropagatesInfrastructureFailure(t *testing.T) {
	service := &workerPreflightStub{
		revision: "runtime-catalog-7",
		err:      assert.AnError,
	}
	compiler := newWorkerCreationCompiler(service, nil)

	_, err := compiler.Compile(
		context.Background(),
		workerTemplateScope(),
		workercreation.Draft{OptionsRevision: "runtime-catalog-7"},
	)

	assert.ErrorIs(t, err, assert.AnError)
}

func TestWorkerCreationCompilerRejectsIncompleteSuccess(t *testing.T) {
	service := &workerPreflightStub{
		revision: "runtime-catalog-7",
		result: workercreation.PreflightResult{
			OptionsRevision: "runtime-catalog-7",
		},
	}
	compiler := newWorkerCreationCompiler(service, nil)

	_, err := compiler.Compile(
		context.Background(),
		workerTemplateScope(),
		workercreation.Draft{OptionsRevision: "runtime-catalog-7"},
	)

	assert.ErrorIs(t, err, control.ErrCorrupt)
}

func TestNewWorkerCreationCompilerRejectsUnavailableService(t *testing.T) {
	_, err := NewWorkerCreationCompiler(nil)
	assert.Error(t, err)
}

type workerPreflightStub struct {
	revision string
	result   workercreation.PreflightResult
	err      error
}

func (stub *workerPreflightStub) Revision() string { return stub.revision }

func (stub *workerPreflightStub) Preflight(
	context.Context,
	specservice.Scope,
	workercreation.Draft,
) (workercreation.PreflightResult, error) {
	return stub.result, stub.err
}
