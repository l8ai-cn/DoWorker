import { describe, expect, it } from "vitest";
import {
  parseCanonicalResourceJson,
  parseResourceYaml,
  stringifyResourceYaml,
} from "./resource-yaml-codec";
import {
  createExpertDraft,
  createGoalLoopDraft,
  createPromptDraft,
  createWorkflowDraft,
} from "./resource-definition-drafts";
import { createResourceBindingDraft } from "./resource-binding-draft";
import { createWorkerInvocationDraft } from "./worker-invocation-draft";
import { createWorkerTemplateDraft } from "./worker-template-draft";
import { createResourceDraft } from "./resource-draft-factory";

const RESOURCE_KINDS = [
  "WorkerTemplate",
  "Worker",
  "Prompt",
  "Expert",
  "Workflow",
  "GoalLoop",
  "ModelBinding",
  "Repository",
  "Skill",
  "KnowledgeBase",
  "EnvironmentBundle",
  "ComputeTarget",
  "ResourceProfile",
  "ToolBinding",
] as const;

describe("resource YAML codec", () => {
  it("round trips the WorkerTemplate typed draft", () => {
    const draft = createWorkerTemplateDraft("acme");
    draft.metadata.name = "code-reviewer";
    draft.spec.workerType = "codex";
    draft.spec.optionsRevision = "runtime-catalog-7";

    const yaml = stringifyResourceYaml(draft);
    const parsed = parseResourceYaml(yaml, "WorkerTemplate");

    expect(parsed).toEqual(draft);
    expect(yaml).toContain("kind: WorkerTemplate");
  });

  it("rejects duplicate keys without echoing their values", () => {
    const source = [
      "apiVersion: agentcloud.io/v1alpha1",
      "kind: WorkerTemplate",
      "metadata:",
      "  name: reviewer",
      "  name: pasted-secret-value",
      "spec: {}",
    ].join("\n");

    expect(() => parseResourceYaml(source, "WorkerTemplate"))
      .toThrow("YAML syntax error");
    try {
      parseResourceYaml(source, "WorkerTemplate");
    } catch (error) {
      expect(String(error)).not.toContain("pasted-secret-value");
    }
  });

  it("rejects multiple documents and oversized physical lines", () => {
    expect(() => parseResourceYaml(
      "kind: WorkerTemplate\n---\nkind: Skill",
      "WorkerTemplate",
    ))
      .toThrow("exactly one document");
    expect(() => parseResourceYaml(
      `value: ${"x".repeat(65 * 1024)}`,
      "WorkerTemplate",
    ))
      .toThrow("64 KiB");
  });

  it("rejects integers that JavaScript cannot represent exactly", () => {
    const source = [
      "apiVersion: agentcloud.io/v1alpha1",
      "kind: GoalLoop",
      "metadata:",
      "  name: release-loop",
      "spec:",
      "  tokenBudget: 9007199254740993",
    ].join("\n");

    expect(() => parseResourceYaml(source, "GoalLoop"))
      .toThrow("safe integer range");
  });

  it("round trips a Worker invocation without changing kind", () => {
    const draft = createWorkerInvocationDraft("acme");
    draft.metadata.name = "review-run";
    draft.spec.workerTemplateRef.name = "code-reviewer";

    const yaml = stringifyResourceYaml(draft);

    expect(parseResourceYaml(yaml, "Worker")).toEqual(draft);
    expect(() => parseResourceYaml(yaml, "WorkerTemplate"))
      .toThrow("YAML must be a WorkerTemplate resource");
  });

  it.each([
    ["Prompt", createPromptDraft("acme")],
    ["Expert", createExpertDraft("acme")],
    ["Workflow", createWorkflowDraft("acme")],
    ["GoalLoop", createGoalLoopDraft("acme")],
    ["ModelBinding", createResourceBindingDraft("ModelBinding", "acme")],
    ["ToolBinding", createResourceBindingDraft("ToolBinding", "acme")],
  ] as const)("round trips the %s typed draft", (kind, draft) => {
    const yaml = stringifyResourceYaml(draft);

    expect(parseResourceYaml(yaml, kind)).toEqual(draft);
    expect(yaml).toContain(`kind: ${kind}`);
  });

  it("rejects incomplete canonical resource documents", () => {
    const missingName = {
      ...createPromptDraft("acme"),
      metadata: { namespace: "acme" },
    };
    const missingContent = {
      ...createPromptDraft("acme"),
      metadata: { name: "review-prompt", namespace: "acme" },
      spec: { variables: {} },
    };

    expect(() => parseCanonicalResourceJson(
      encodeJson(missingName),
      "Prompt",
    )).toThrow("metadata.name must be a non-empty string");
    expect(() => parseCanonicalResourceJson(
      encodeJson(missingContent),
      "Prompt",
    )).toThrow("spec.content must be string");
  });

  it.each(RESOURCE_KINDS)(
    "accepts the complete %s canonical draft shape",
    (kind) => {
      const draft = createResourceDraft(kind, "acme");
      draft.metadata.name = "planned-resource";

      expect(parseCanonicalResourceJson(encodeJson(draft), kind))
        .toEqual(draft);
    },
  );

  it("validates canonical map values and array items recursively", () => {
    const prompt = createPromptDraft("acme");
    prompt.metadata.name = "review-prompt";
    const invalidPrompt = {
      ...prompt,
      spec: { ...prompt.spec, variables: { branch: null } },
    };
    const workerTemplate = createWorkerTemplateDraft("acme");
    workerTemplate.metadata.name = "review-worker";
    const invalidWorkerTemplate = {
      ...workerTemplate,
      spec: {
        ...workerTemplate.spec,
        workspace: {
          ...workerTemplate.spec.workspace,
          knowledgeMounts: [null],
        },
      },
    };

    expect(() => parseCanonicalResourceJson(
      encodeJson(invalidPrompt),
      "Prompt",
    )).toThrow("spec.variables entries must be object");
    expect(() => parseCanonicalResourceJson(
      encodeJson(invalidWorkerTemplate),
      "WorkerTemplate",
    )).toThrow("spec.workspace.knowledgeMounts[0] must be object");
  });

  it("rejects GoalLoop canonical drafts without a description", () => {
    const draft = createGoalLoopDraft("acme");
    draft.metadata.name = "release-loop";
    const spec: Record<string, unknown> = { ...draft.spec };
    delete spec.description;

    expect(() => parseCanonicalResourceJson(
      encodeJson({ ...draft, spec }),
      "GoalLoop",
    )).toThrow("spec.description must be string");
  });

  it("uses the WorkerTemplate canonical CPU field names", () => {
    const draft = createWorkerTemplateDraft("acme");
    draft.metadata.name = "review-worker";
    draft.spec.runtime.customResources = {
      cpuRequestMilliCPU: 500,
      cpuLimitMilliCPU: 1000,
      memoryRequestBytes: 536870912,
      memoryLimitBytes: 1073741824,
      storageRequestBytes: 1073741824,
      storageLimitBytes: 10737418240,
    };

    expect(parseCanonicalResourceJson(
      encodeJson(draft),
      "WorkerTemplate",
    )).toEqual(draft);
    expect(() => parseCanonicalResourceJson(
      encodeJson({
        ...draft,
        spec: {
          ...draft.spec,
          runtime: {
            ...draft.spec.runtime,
            customResources: {
              ...draft.spec.runtime.customResources,
              cpuRequestMillicpu: 500,
            },
          },
        },
      }),
      "WorkerTemplate",
    )).toThrow("customResources contains unsupported fields");
  });
});

function encodeJson(value: unknown): Uint8Array {
  return new TextEncoder().encode(JSON.stringify(value));
}
