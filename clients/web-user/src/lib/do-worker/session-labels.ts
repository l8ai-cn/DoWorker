export const WRAPPER_LABEL_KEY = "do-worker.wrapper";
export const UI_MODE_LABEL_KEY = "do-worker.ui";
export const UI_MODE_TERMINAL_VALUE = "terminal";
export const CLOSED_LABEL_KEY = "do-worker.closed";
export const FORK_SOURCE_LABEL_KEY = "do-worker.fork.source_id";

const LEGACY_LABEL_ALIASES: Record<string, string> = {
  "do-worker.ui": UI_MODE_LABEL_KEY,
  "omnigent.wrapper": WRAPPER_LABEL_KEY,
  "do-worker.closed": CLOSED_LABEL_KEY,
  "omnigent.fork.source_id": FORK_SOURCE_LABEL_KEY,
};

export function sessionLabel(
  labels: Record<string, string> | null | undefined,
  key: string,
): string | undefined {
  if (!labels) return undefined;
  if (labels[key] !== undefined) return labels[key];
  for (const [legacy, canonical] of Object.entries(LEGACY_LABEL_ALIASES)) {
    if (canonical === key && labels[legacy] !== undefined) {
      return labels[legacy];
    }
  }
  return undefined;
}

export function sessionLabelEquals(
  labels: Record<string, string> | null | undefined,
  key: string,
  value: string,
): boolean {
  return sessionLabel(labels, key) === value;
}
