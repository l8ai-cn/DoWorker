export function isUnexpectedTerminalClose(code: number): boolean {
  return code === 1001 || code === 1006 || code === 1012 || code === 1013;
}

export const TERMINAL_RECONNECT_BACKOFF_MS = [500, 1000, 2000, 4000, 8000] as const;
