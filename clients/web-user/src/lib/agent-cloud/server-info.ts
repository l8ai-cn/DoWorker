import { hostFetch } from "./host-config";

export interface ServerInfo {
  accounts_enabled: boolean;
  login_url: string | null;
  needs_setup: boolean;
  databricks_features: boolean;
  managed_sandboxes_enabled: boolean;
  sandbox_provider: string | null;
  server_version: string | null;
  smart_routing_enabled: boolean;
}

const OFF: ServerInfo = {
  accounts_enabled: false,
  login_url: null,
  needs_setup: false,
  databricks_features: false,
  managed_sandboxes_enabled: false,
  sandbox_provider: null,
  server_version: null,
  smart_routing_enabled: false,
};

let _cached: ServerInfo | null = null;
let _pending: Promise<ServerInfo> | null = null;

export async function resolveServerInfo(): Promise<ServerInfo> {
  if (_cached !== null) return _cached;
  if (_pending !== null) return _pending;
  _pending = (async () => {
    try {
      const res = await hostFetch("/v1/info");
      if (res.ok) {
        const data = (await res.json()) as Partial<ServerInfo>;
        _cached = {
          accounts_enabled: data.accounts_enabled === true,
          login_url: typeof data.login_url === "string" ? data.login_url : null,
          needs_setup: data.needs_setup === true,
          databricks_features: data.databricks_features === true,
          managed_sandboxes_enabled: data.managed_sandboxes_enabled === true,
          sandbox_provider:
            typeof data.sandbox_provider === "string" ? data.sandbox_provider : null,
          server_version: typeof data.server_version === "string" ? data.server_version : null,
          smart_routing_enabled: data.smart_routing_enabled === true,
        };
        return _cached;
      }
    } catch {
      // Network failure — fall through to the off sentinel.
    }
    _cached = OFF;
    return _cached;
  })();
  return _pending;
}

export function getCachedServerInfo(): ServerInfo | null {
  return _cached;
}

const SANDBOX_PROVIDER_NAMES: Record<string, string> = {
  modal: "Modal",
  lakebox: "Databricks",
  daytona: "Daytona",
  e2b: "E2B",
};

export function sandboxOptionLabel(provider: string | null): string {
  if (!provider) return "New Sandbox";
  const name =
    SANDBOX_PROVIDER_NAMES[provider] ?? provider.charAt(0).toUpperCase() + provider.slice(1);
  return `${name} Sandbox`;
}
