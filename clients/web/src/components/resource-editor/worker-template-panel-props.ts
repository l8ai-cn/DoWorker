import type { WorkerTemplateDraft } from "./resource-editor-types";
import type { ResourceReferenceCatalog } from "./resource-reference-options";

export interface WorkerTemplatePanelProps {
  draft: WorkerTemplateDraft;
  catalog: ResourceReferenceCatalog;
  onChange: (draft: WorkerTemplateDraft) => void;
}
