/** localStorage / sessionStorage key prefixes — SSOT for web-user prefs. */
export const DO_WORKER_STORAGE_PREFIX = "do-worker";
export const LEGACY_OMNIGENT_STORAGE_PREFIX = "omnigent";

export function doWorkerStorageKey(suffix: string): string {
  return `${DO_WORKER_STORAGE_PREFIX}:${suffix}`;
}

export function readStorageWithLegacy(suffix: string): string | null {
  try {
    return (
      localStorage.getItem(doWorkerStorageKey(suffix)) ??
      localStorage.getItem(`${LEGACY_OMNIGENT_STORAGE_PREFIX}:${suffix}`)
    );
  } catch {
    return null;
  }
}

export function writeStorageKey(suffix: string, value: string): void {
  try {
    localStorage.setItem(doWorkerStorageKey(suffix), value);
    localStorage.removeItem(`${LEGACY_OMNIGENT_STORAGE_PREFIX}:${suffix}`);
  } catch {
    // pass
  }
}
