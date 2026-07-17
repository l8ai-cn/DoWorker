package orchestrationresource

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterDefinitionSchemasRegistersEveryDefinitionKind(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterDefinitionSchemas(registry))

	for _, kind := range []string{
		KindPrompt,
		KindWorker,
		KindExpert,
		KindWorkflow,
		KindGoalLoop,
	} {
		require.True(t, registry.Has(TypeMeta{
			APIVersion: APIVersionV1Alpha1,
			Kind:       kind,
		}), kind)
	}
}

func TestDefinitionSchemasDecodeValidSpecs(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterDefinitionSchemas(registry))

	tests := []struct {
		kind string
		spec string
	}{
		{KindPrompt, `{
			"content":"Review {{change}}",
			"variables":{"change":{"required":true,"default":""}}
		}`},
		{KindWorker, `{
			"workerTemplateRef":{"kind":"WorkerTemplate","name":"reviewer"},
			"promptRef":{"kind":"Prompt","name":"review-task"},
			"inputs":{"change":"pr-42"},
			"alias":"reviewer-42"
		}`},
		{KindExpert, `{
			"workerTemplateRef":{"kind":"WorkerTemplate","name":"reviewer"},
			"promptRef":{"kind":"Prompt","name":"review-system"},
			"description":"Reviews changes",
			"category":"engineering",
			"releaseNotes":"Initial revision"
		}`},
		{KindWorkflow, `{
			"workerTemplateRef":{"kind":"WorkerTemplate","name":"reviewer"},
			"promptRef":{"kind":"Prompt","name":"nightly-review"},
			"inputs":{},
			"executionMode":"direct",
			"cronExpression":"0 2 * * *",
			"sandboxStrategy":"fresh",
			"sessionPersistence":false,
			"concurrencyPolicy":"skip",
			"maxConcurrentRuns":1,
			"maxRetainedRuns":30,
			"timeoutMinutes":60,
			"idleTimeoutSeconds":30
		}`},
		{KindGoalLoop, `{
			"workerTemplateRef":{"kind":"WorkerTemplate","name":"reviewer"},
			"description":"Repair checkout deterministically",
			"objective":"Fix checkout",
			"acceptanceCriteria":["Tests pass"],
			"verificationCommand":"go test ./...",
			"maxIterations":10,
			"timeoutMinutes":60,
			"noProgressLimit":3,
			"sameErrorLimit":2,
			"escalationPolicy":"pause"
		}`},
	}
	for _, test := range tests {
		t.Run(test.kind, func(t *testing.T) {
			_, err := registry.DecodeAndValidate(definitionManifest(
				t,
				test.kind,
				test.spec,
			))
			require.NoError(t, err)
		})
	}
}

func TestDefinitionSchemasRejectInvalidReferencesAndFields(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterDefinitionSchemas(registry))

	tests := []struct {
		name string
		kind string
		spec string
		want string
	}{
		{
			name: "wrong worker template kind",
			kind: KindWorker,
			spec: `{"workerTemplateRef":{"kind":"Expert","name":"reviewer"},
				"inputs":{},"alias":""}`,
			want: KindWorkerTemplate,
		},
		{
			name: "worker alias exceeds pod limit",
			kind: KindWorker,
			spec: `{"workerTemplateRef":{"kind":"WorkerTemplate","name":"reviewer"},
				"inputs":{},"alias":"` + strings.Repeat("a", 101) + `"}`,
			want: "alias",
		},
		{
			name: "invalid prompt variable",
			kind: KindPrompt,
			spec: `{"content":"Review","variables":{"Change_ID":{
				"required":true,"default":""}}}`,
			want: "prompt variable",
		},
		{
			name: "workflow persistent concurrency",
			kind: KindWorkflow,
			spec: `{
				"workerTemplateRef":{"kind":"WorkerTemplate","name":"reviewer"},
				"promptRef":{"kind":"Prompt","name":"nightly-review"},
				"inputs":{},"executionMode":"direct","sandboxStrategy":"persistent",
				"sessionPersistence":true,"concurrencyPolicy":"skip",
				"maxConcurrentRuns":2,"maxRetainedRuns":0,
				"timeoutMinutes":60,"idleTimeoutSeconds":30
			}`,
			want: "maxConcurrentRuns",
		},
		{
			name: "workflow unsupported concurrency policy",
			kind: KindWorkflow,
			spec: `{
				"workerTemplateRef":{"kind":"WorkerTemplate","name":"reviewer"},
				"promptRef":{"kind":"Prompt","name":"nightly-review"},
				"inputs":{},"executionMode":"direct","sandboxStrategy":"fresh",
				"sessionPersistence":false,"concurrencyPolicy":"queue",
				"maxConcurrentRuns":1,"maxRetainedRuns":0,
				"timeoutMinutes":60,"idleTimeoutSeconds":30
			}`,
			want: "currently only supports skip",
		},
		{
			name: "workflow private callback",
			kind: KindWorkflow,
			spec: `{
				"workerTemplateRef":{"kind":"WorkerTemplate","name":"reviewer"},
				"promptRef":{"kind":"Prompt","name":"nightly-review"},
				"inputs":{},"executionMode":"direct","sandboxStrategy":"fresh",
				"sessionPersistence":false,"concurrencyPolicy":"skip",
				"maxConcurrentRuns":1,"maxRetainedRuns":0,
				"timeoutMinutes":60,"idleTimeoutSeconds":30,
				"callbackUrl":"http://127.0.0.1/callback"
			}`,
			want: "private or local",
		},
		{
			name: "workflow invalid cron",
			kind: KindWorkflow,
			spec: `{
				"workerTemplateRef":{"kind":"WorkerTemplate","name":"reviewer"},
				"promptRef":{"kind":"Prompt","name":"nightly-review"},
				"inputs":{},"executionMode":"direct","sandboxStrategy":"fresh",
				"sessionPersistence":false,"concurrencyPolicy":"skip",
				"maxConcurrentRuns":1,"maxRetainedRuns":0,
				"timeoutMinutes":60,"idleTimeoutSeconds":30,
				"cronExpression":"not-a-cron"
			}`,
			want: "cronExpression is invalid",
		},
		{
			name: "goal loop unknown field",
			kind: KindGoalLoop,
			spec: `{
				"workerTemplateRef":{"kind":"WorkerTemplate","name":"reviewer"},
				"objective":"Fix","acceptanceCriteria":["Done"],
				"verificationCommand":"go test ./...","maxIterations":1,
				"timeoutMinutes":1,"noProgressLimit":1,"sameErrorLimit":1,
				"escalationPolicy":"pause","fallback":true
			}`,
			want: "unknown",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := registry.DecodeAndValidate(definitionManifest(
				t,
				test.kind,
				test.spec,
			))
			require.ErrorContains(t, err, test.want)
		})
	}
}

func definitionManifest(
	t *testing.T,
	kind string,
	spec string,
) Manifest {
	t.Helper()
	return Manifest{
		TypeMeta: TypeMeta{
			APIVersion: APIVersionV1Alpha1,
			Kind:       kind,
		},
		Metadata: Metadata{
			Name:      "definition",
			Namespace: "acme",
		},
		Spec: json.RawMessage(spec),
	}
}
