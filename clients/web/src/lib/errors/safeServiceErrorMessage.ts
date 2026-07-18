import { getErrorMessage } from "./serviceError";

export function safeServiceErrorMessage(
  error: unknown,
  fallback: string,
): string {
  const message = getErrorMessage(error)
    .replace(/\s+for url \(\s*https?:\/\/[^)]+\)/gi, "")
    .replace(/\s+@\s+https?:\/\/\S+$/, "")
    .replace(/https?:\/\/\S+/gi, "[redacted]")
    .trim();
  return message || fallback;
}
