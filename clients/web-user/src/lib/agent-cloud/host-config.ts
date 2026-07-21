import type { ReactNode } from "react";

import { readAgentCloudJWT, readAgentCloudOrgSlug } from "./auth-session";

export interface UserSuggestion {
  userId: string;
  displayName?: string;
}

export interface AgentCloudHostConfig {
  fetcher?: (path: string, init?: RequestInit) => Promise<Response>;
  searchUsers?: (query: string, options?: { signal?: AbortSignal }) => Promise<UserSuggestion[]>;
  resolveWebSocketUrl?: (path: string) => string;
  transformShareLink?: (relativePath: string) => string;
  cliServerUrlSuffix?: string;
  docsLinks?: {
    newSandbox?: ReactNode;
    databricksGitCredentials?: ReactNode;
  };
}

let _config: AgentCloudHostConfig = {};
let _embedRoot: HTMLElement | null = null;

export function setAgentCloudHostConfig(config: AgentCloudHostConfig): void {
  if (!config?.fetcher && _config.fetcher) return;
  _config = config ?? {};
}

export function getAgentCloudHostConfig(): AgentCloudHostConfig {
  return _config;
}

export function getAgentCloudUserSearch(): AgentCloudHostConfig["searchUsers"] {
  return _config.searchUsers;
}

export function getAgentCloudTransformShareLink(): AgentCloudHostConfig["transformShareLink"] {
  return _config.transformShareLink;
}

export function setEmbedRoot(el: HTMLElement | null): void {
  _embedRoot = el;
}

export function getEmbedRoot(): HTMLElement | null {
  return _embedRoot;
}

export function hostFetch(path: string, init?: RequestInit): Promise<Response> {
  if (_config.fetcher) {
    return _config.fetcher(path, init);
  }
  return fetch(path, init);
}

export function resolveWebSocketUrl(path: string): string {
  if (_config.resolveWebSocketUrl) {
    return _config.resolveWebSocketUrl(path);
  }
  const scheme = window.location.protocol === "https:" ? "wss:" : "ws:";
  let url = `${scheme}//${window.location.host}${path}`;
  const params = new URLSearchParams();
  const jwt = readAgentCloudJWT();
  if (jwt) params.set("token", jwt);
  const org = readAgentCloudOrgSlug();
  if (org) params.set("org_slug", org);
  const qs = params.toString();
  if (qs) url += (path.includes("?") ? "&" : "?") + qs;
  return url;
}

export function getCliServerUrl(): string {
  const origin = typeof window !== "undefined" ? window.location.origin : "";
  return origin + (_config.cliServerUrlSuffix ?? "");
}
