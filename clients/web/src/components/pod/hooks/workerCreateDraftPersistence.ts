import type { WorkerSpecDraft } from "@/lib/api/facade/podConnect";

const STORAGE_PREFIX = "agentcloud.worker-create-draft.v1";
const SENSITIVE_FIELD = /(api[-_]?key|access[-_]?key|token|secret|password|credential|private[-_]?key)/i;

export interface PersistedWorkerCreateDraft {
  step: 1 | 2 | 3 | 4;
  fillPrompt: string;
  draft: Partial<WorkerSpecDraft>;
}

export function loadWorkerCreateDraft(
  orgSlug: string,
): PersistedWorkerCreateDraft | null {
  if (!orgSlug || typeof window === "undefined") return null;
  try {
    const raw = window.sessionStorage.getItem(storageKey(orgSlug));
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Partial<PersistedWorkerCreateDraft>;
    if (!isStep(parsed.step) || typeof parsed.fillPrompt !== "string") return null;
    if (!parsed.draft || typeof parsed.draft !== "object") return null;
    return {
      step: parsed.step,
      fillPrompt: parsed.fillPrompt,
      draft: parsed.draft,
    };
  } catch {
    return null;
  }
}

export function persistWorkerCreateDraft(
  orgSlug: string,
  state: PersistedWorkerCreateDraft,
): void {
  if (!orgSlug || typeof window === "undefined") return;
  try {
    window.sessionStorage.setItem(
      storageKey(orgSlug),
      JSON.stringify({
        ...state,
        draft: sanitizeDraft(state.draft),
      }),
    );
  } catch {
    // Storage can be unavailable in private browsing; the form remains usable.
  }
}

export function clearWorkerCreateDraft(orgSlug: string): void {
  if (!orgSlug || typeof window === "undefined") return;
  try {
    window.sessionStorage.removeItem(storageKey(orgSlug));
  } catch {
    // Ignore storage cleanup failures.
  }
}

function sanitizeDraft(draft: Partial<WorkerSpecDraft>): Partial<WorkerSpecDraft> {
  const typeConfig = draft.type_config_values;
  const safeTypeConfig = typeConfig
    ? Object.fromEntries(
        Object.entries(typeConfig).filter(([field]) => !SENSITIVE_FIELD.test(field)),
      )
    : undefined;
  return {
    ...draft,
    type_config_values: safeTypeConfig,
  };
}

function storageKey(orgSlug: string): string {
  return `${STORAGE_PREFIX}:${orgSlug}`;
}

function isStep(value: unknown): value is 1 | 2 | 3 | 4 {
  return value === 1 || value === 2 || value === 3 || value === 4;
}
