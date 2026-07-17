import { describe, expect, it } from "vitest";
import {
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
      "apiVersion: agentsmesh.io/v1alpha1",
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
      "apiVersion: agentsmesh.io/v1alpha1",
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
});
