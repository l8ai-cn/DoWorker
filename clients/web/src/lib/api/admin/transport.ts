// Direct Connect-RPC transport for the admin service. clients/web routes
// most RPCs through the wasm bridge, but proto.admin.v1.AdminService has no
// wasm binding — so admin pages talk to the backend Connect handlers over a
// plain binary fetch, same wire shape as web-admin's transport. The JWT comes
// from the wasm AuthManager (SSOT) rather than a Zustand token field.
import {
  create,
  toBinary,
  fromBinary,
  type DescMessage,
  type MessageInitShape,
  type MessageShape,
} from "@bufbuild/protobuf";

import { getApiBaseUrl } from "@/lib/env";
import { getAuthManager } from "@/lib/wasm-core";
import { useAuthStore } from "@/stores/auth";

export class AdminConnectError extends Error {
  readonly code: string;
  readonly status: number;
  constructor(message: string, code: string, status: number) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

function authToken(): string | null {
  try {
    return getAuthManager().get_token() ?? null;
  } catch {
    return null;
  }
}

// `service` is the fully-qualified proto name (e.g. "proto.admin.v1.AdminService");
// `method` is the PascalCase RPC name. Pass the generated *Schema constants.
export async function callAdminConnect<I extends DescMessage, O extends DescMessage>(
  service: string,
  method: string,
  inputSchema: I,
  outputSchema: O,
  input: MessageInitShape<I>,
): Promise<MessageShape<O>> {
  const body = toBinary(inputSchema, create(inputSchema, input));

  const headers: Record<string, string> = {
    "Content-Type": "application/proto",
    "connect-protocol-version": "1",
  };
  const token = authToken();
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const base = getApiBaseUrl().replace(/\/$/, "");
  const resp = await fetch(`${base}/${service}/${method}`, {
    method: "POST",
    headers,
    body,
  });

  if (resp.status === 401) {
    void useAuthStore.getState().logout();
    throw new AdminConnectError("Session expired. Please login again.", "unauthenticated", 401);
  }

  if (!resp.ok) {
    let detail = `HTTP ${resp.status}`;
    let code = "unknown";
    try {
      const err = (await resp.json()) as { code?: string; message?: string };
      detail = err.message || detail;
      code = err.code || code;
    } catch {
      /* body wasn't JSON — keep defaults */
    }
    throw new AdminConnectError(detail, code, resp.status);
  }

  return fromBinary(outputSchema, new Uint8Array(await resp.arrayBuffer()));
}
