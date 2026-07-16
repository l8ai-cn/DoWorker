import initWasm, { WasmAuthManager } from "do-worker-wasm";
import { apiBaseUrl } from "./api-config";

let managerPromise: Promise<WasmAuthManager> | undefined;

export function mobileAuthBaseUrl(): string {
  const configured = apiBaseUrl();
  if (configured) return configured;
  if (typeof window !== "undefined") return window.location.origin;
  return "";
}

export function mobileAuthUrlSlug(baseUrl: string): string {
  const trimmed = baseUrl.replace(/\/+$/, "");
  const separator = trimmed.indexOf("://");
  const scheme = separator === -1 ? "" : trimmed.slice(0, separator);
  const rest = separator === -1 ? trimmed : trimmed.slice(separator + 3);
  const authority = rest.split(/[/?#]/, 1)[0] ?? "";
  const normalized = scheme
    ? `${scheme.toLowerCase()}_${authority.toLowerCase()}`
    : authority.toLowerCase();
  return normalized.replace(/[^a-zA-Z0-9]/g, "_").slice(0, 64);
}

export function mobileAuthSessionStorageKey(): string {
  return `do-worker-auth/${mobileAuthUrlSlug(mobileAuthBaseUrl())}/session`;
}

export function getMobileAuthManager(): Promise<WasmAuthManager> {
  if (!managerPromise) {
    const baseUrl = mobileAuthBaseUrl();
    managerPromise = initWasm().then(() => new WasmAuthManager(baseUrl));
  }
  return managerPromise;
}
