import { createResourceBindingDraft } from "./resource-binding-draft";
import {
  createExpertDraft,
  createGoalLoopDraft,
  createPromptDraft,
  createWorkflowDraft,
} from "./resource-definition-drafts";
import {
  isResourceBindingKind,
  type ResourceDraft,
  type ResourceEditorKind,
} from "./resource-editor-types";
import { createWorkerInvocationDraft } from "./worker-invocation-draft";
import { createWorkerTemplateDraft } from "./worker-template-draft";

export function createResourceDraft(
  kind: ResourceEditorKind,
  namespace: string,
): ResourceDraft {
  if (isResourceBindingKind(kind)) {
    return createResourceBindingDraft(kind, namespace);
  }
  switch (kind) {
    case "WorkerTemplate":
      return createWorkerTemplateDraft(namespace);
    case "Worker":
      return createWorkerInvocationDraft(namespace);
    case "Prompt":
      return createPromptDraft(namespace);
    case "Expert":
      return createExpertDraft(namespace);
    case "Workflow":
      return createWorkflowDraft(namespace);
    case "GoalLoop":
      return createGoalLoopDraft(namespace);
  }
}
