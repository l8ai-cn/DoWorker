export const lifecycleItems = [
  {
    titleKey: "lifecycle.validate.title",
    descriptionKey: "lifecycle.validate.description",
  },
  {
    titleKey: "lifecycle.plan.title",
    descriptionKey: "lifecycle.plan.description",
  },
  {
    titleKey: "lifecycle.apply.title",
    descriptionKey: "lifecycle.apply.description",
  },
] as const;

export const resourceKindRows = [
  {
    nameKey: "kinds.bindings.name",
    applyKey: "kinds.bindings.apply",
    purposeKey: "kinds.bindings.purpose",
  },
  {
    nameKey: "kinds.prompt.name",
    applyKey: "kinds.prompt.apply",
    purposeKey: "kinds.prompt.purpose",
  },
  {
    nameKey: "kinds.workerTemplate.name",
    applyKey: "kinds.workerTemplate.apply",
    purposeKey: "kinds.workerTemplate.purpose",
  },
  {
    nameKey: "kinds.worker.name",
    applyKey: "kinds.worker.apply",
    purposeKey: "kinds.worker.purpose",
  },
  {
    nameKey: "kinds.expertWorkflow.name",
    applyKey: "kinds.expertWorkflow.apply",
    purposeKey: "kinds.expertWorkflow.purpose",
  },
  {
    nameKey: "kinds.goalLoop.name",
    applyKey: "kinds.goalLoop.apply",
    purposeKey: "kinds.goalLoop.purpose",
  },
] as const;

export const editorSteps = [
  "editor.steps.form",
  "editor.steps.yaml",
  "editor.steps.validate",
  "editor.steps.plan",
  "editor.steps.apply",
] as const;

export const securityItems = [
  "security.items.credentials",
  "security.items.authorization",
  "security.items.stale",
] as const;
