import { hostFetch } from "@/lib/host";

export interface EmbedSessionAccess {
  accessToken: string;
  expiresAt: number;
  sessionId: string;
  orgSlug: string;
  capabilities: string[];
  parentOrigins: string[];
}

export interface EmbedContextBootstrap {
  expiresAt: number;
  parentOrigins: string[];
}

const pendingInspections = new Map<string, Promise<EmbedContextBootstrap>>();
const pendingRedemptions = new Map<string, Promise<EmbedSessionAccess>>();

type EmbedContextResponse = {
  access_token?: unknown;
  expires_at?: unknown;
  session_id?: unknown;
  org_slug?: unknown;
  capabilities?: unknown;
  parent_origins?: unknown;
};

type EmbedContextBootstrapResponse = {
  expires_at?: unknown;
  parent_origins?: unknown;
};

const EMBED_CAPABILITIES = new Set(["read", "write", "approve", "terminal", "control"]);
type EmbedFetch = (path: string, init?: RequestInit) => Promise<Response>;

export function readEmbedContext(search: string): string {
  const values = new URLSearchParams(search).getAll("embed_context");
  if (values.length === 0 || values[0] === "") {
    throw new Error("embed_context is required");
  }
  if (values.length !== 1) {
    throw new Error("embed_context must appear exactly once");
  }
  return values[0];
}

export async function redeemEmbedContext(
  context: string,
  redemptionProof: string,
  fetcher: EmbedFetch = hostFetch,
): Promise<EmbedSessionAccess> {
  const response = await fetcher("/v1/embed-contexts/redeem", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${context}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ redemption_proof: redemptionProof }),
    cache: "no-store",
  });
  if (!response.ok) {
    throw new Error("Unable to open the embedded session");
  }
  return parseEmbedSessionAccess((await response.json()) as EmbedContextResponse);
}

export async function inspectEmbedContext(
  context: string,
  fetcher: EmbedFetch = hostFetch,
): Promise<EmbedContextBootstrap> {
  const response = await fetcher("/v1/embed-contexts/inspect", {
    method: "POST",
    headers: { Authorization: `Bearer ${context}` },
    cache: "no-store",
  });
  if (!response.ok) {
    throw new Error("Unable to open the embedded session");
  }
  return parseEmbedContextBootstrap((await response.json()) as EmbedContextBootstrapResponse);
}

export function inspectEmbedContextOnce(
  context: string,
  fetcher: EmbedFetch = hostFetch,
): Promise<EmbedContextBootstrap> {
  const pending = pendingInspections.get(context);
  if (pending) return pending;
  const inspection = inspectEmbedContext(context, fetcher);
  pendingInspections.set(context, inspection);
  const release = () => {
    if (pendingInspections.get(context) === inspection) {
      pendingInspections.delete(context);
    }
  };
  void inspection.then(release, release);
  return inspection;
}

export function redeemEmbedContextOnce(
  context: string,
  redemptionProof: string,
  fetcher: EmbedFetch = hostFetch,
): Promise<EmbedSessionAccess> {
  const key = `${context}\u0000${redemptionProof}`;
  const pending = pendingRedemptions.get(key);
  if (pending) return pending;
  const redemption = redeemEmbedContext(context, redemptionProof, fetcher);
  pendingRedemptions.set(key, redemption);
  const release = () => {
    if (pendingRedemptions.get(key) === redemption) {
      pendingRedemptions.delete(key);
    }
  };
  void redemption.then(release, release);
  return redemption;
}

export function clearEmbedContextFromLocation(): void {
  const url = new URL(window.location.href);
  url.searchParams.delete("embed_context");
  window.history.replaceState({}, "", url);
}

function parseEmbedSessionAccess(body: EmbedContextResponse): EmbedSessionAccess {
  if (
    typeof body.access_token !== "string" ||
    typeof body.expires_at !== "number" ||
    typeof body.session_id !== "string" ||
    typeof body.org_slug !== "string" ||
    body.org_slug === "" ||
    !Array.isArray(body.capabilities) ||
    !body.capabilities.every(
      (capability) => typeof capability === "string" && EMBED_CAPABILITIES.has(capability),
    ) ||
    !Array.isArray(body.parent_origins) ||
    !body.parent_origins.every((origin) => typeof origin === "string")
  ) {
    throw new Error("Embedded session response is invalid");
  }
  return {
    accessToken: body.access_token,
    expiresAt: body.expires_at,
    sessionId: body.session_id,
    orgSlug: body.org_slug,
    capabilities: body.capabilities,
    parentOrigins: body.parent_origins,
  };
}

function parseEmbedContextBootstrap(body: EmbedContextBootstrapResponse): EmbedContextBootstrap {
  if (
    typeof body.expires_at !== "number" ||
    !Array.isArray(body.parent_origins) ||
    !body.parent_origins.every((origin) => typeof origin === "string")
  ) {
    throw new Error("Embedded context response is invalid");
  }
  return {
    expiresAt: body.expires_at,
    parentOrigins: body.parent_origins,
  };
}
