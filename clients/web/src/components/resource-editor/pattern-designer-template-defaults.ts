import type { WorkerTemplateDraft } from "./resource-editor-types";

export function patternDesignerWorkspace(
  workerType: string,
  workspace: WorkerTemplateDraft["spec"]["workspace"],
): WorkerTemplateDraft["spec"]["workspace"] {
  if (workerType !== "pattern-designer") return workspace;
  return {
    ...workspace,
    skillRefs: [
      "pattern-generate",
      "canvas-compose",
      "pattern-seam-review",
      "lovart-api",
    ].map((name) => ({ kind: "Skill", name })),
  };
}

export function patternDesignerSecretRefs(
  workerType: string,
  refs: WorkerTemplateDraft["spec"]["typeConfig"]["secretRefs"],
): WorkerTemplateDraft["spec"]["typeConfig"]["secretRefs"] {
  if (workerType !== "pattern-designer") return refs;
  return {
    ...refs,
    LOVART_ACCESS_KEY: { kind: "EnvironmentBundle", name: "lovart" },
    LOVART_SECRET_KEY: { kind: "EnvironmentBundle", name: "lovart" },
  };
}
