package fixture

import (
	"context"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/tests/mcp-e2e/client"
)

type WorkflowResource struct {
	Name string
	ID   int64
}

func NewEchoWorkerTemplateResource(
	t *testing.T,
	env *Env,
	rest *client.REST,
	prefix string,
) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	templateName := uniqueAlias(prefix + "-template")
	targetName := uniqueAlias(prefix + "-target")
	profileName := uniqueAlias(prefix + "-profile")
	spec, err := rest.BuildEchoWorkerSpec(
		ctx,
		env.DevOrgSlug,
		"pty",
		templateName,
	)
	if err != nil {
		t.Fatalf("build resource-native echo worker spec: %v", err)
	}
	applyResource(t, ctx, rest, env.DevOrgSlug, "ComputeTarget", targetName, map[string]any{
		"computeTargetId": spec.ComputeTargetID,
	})
	applyResource(t, ctx, rest, env.DevOrgSlug, "ResourceProfile", profileName, map[string]any{
		"resourceProfileId": spec.ResourceProfileID,
	})
	applied := applyResource(
		t,
		ctx,
		rest,
		env.DevOrgSlug,
		"WorkerTemplate",
		templateName,
		map[string]any{
			"optionsRevision": spec.OptionsRevision,
			"workerType":      spec.WorkerTypeSlug,
			"toolRefs":        map[string]any{},
			"runtime": map[string]any{
				"runtimeImageId":  spec.RuntimeImageID,
				"placementPolicy": "automatic",
				"computeTargetRef": map[string]any{
					"kind": "ComputeTarget", "name": targetName,
				},
				"deploymentMode": "pooled",
				"resourceProfileRef": map[string]any{
					"kind": "ResourceProfile", "name": profileName,
				},
			},
			"typeConfig": map[string]any{
				"schemaVersion":   spec.TypeSchemaVersion,
				"values":          spec.TypeConfigValues,
				"secretRefs":      map[string]any{},
				"interactionMode": "pty",
				"automationLevel": "autonomous",
			},
			"workspace": map[string]any{
				"branch": "", "skillRefs": []any{},
				"knowledgeMounts": []any{}, "environmentBundleRefs": []any{},
				"configDocumentBindings": []any{}, "instructions": "",
			},
			"lifecycle": map[string]any{
				"terminationPolicy": "manual", "idleTimeoutMinutes": 0,
			},
			"metadata": map[string]any{"alias": templateName},
		},
	)
	if applied.WorkerSpecSnapshotID <= 0 {
		t.Fatal("worker template apply returned no immutable snapshot")
	}
	return templateName
}

func NewPromptResource(
	t *testing.T,
	env *Env,
	rest *client.REST,
	prefix, content string,
) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	name := uniqueAlias(prefix + "-prompt")
	applyResource(t, ctx, rest, env.DevOrgSlug, "Prompt", name, map[string]any{
		"content": content, "variables": map[string]any{},
	})
	return name
}

func NewWorkflowResource(
	t *testing.T,
	env *Env,
	rest *client.REST,
	prefix,
	prompt string,
) WorkflowResource {
	t.Helper()
	templateName := NewEchoWorkerTemplateResource(t, env, rest, prefix)
	promptName := NewPromptResource(t, env, rest, prefix, prompt)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	name := uniqueAlias(prefix)
	applied := applyResource(t, ctx, rest, env.DevOrgSlug, "Workflow", name, map[string]any{
		"workerTemplateRef": map[string]any{
			"kind": "WorkerTemplate", "name": templateName,
		},
		"promptRef": map[string]any{
			"kind": "Prompt", "name": promptName,
		},
		"inputs": map[string]any{}, "executionMode": "direct",
		"cronExpression": "", "sandboxStrategy": "fresh",
		"sessionPersistence": false, "concurrencyPolicy": "skip",
		"maxConcurrentRuns": 1, "maxRetainedRuns": 0,
		"timeoutMinutes": 1, "idleTimeoutSeconds": 30,
	})
	if applied.WorkflowID <= 0 {
		t.Fatal("workflow apply returned no workflow id")
	}
	return WorkflowResource{Name: name, ID: applied.WorkflowID}
}

func applyResource(
	t *testing.T,
	ctx context.Context,
	rest *client.REST,
	orgSlug, kind, name string,
	spec map[string]any,
) client.AppliedOrchestrationResource {
	t.Helper()
	applied, err := rest.ApplyOrchestrationResource(ctx, orgSlug, kind, map[string]any{
		"apiVersion": "agentsmesh.io/v1alpha1",
		"kind":       kind,
		"metadata": map[string]any{
			"name": name, "namespace": orgSlug, "displayName": name,
		},
		"spec": spec,
	})
	if err != nil {
		t.Fatalf("apply %s resource %s: %v", kind, name, err)
	}
	return applied
}
